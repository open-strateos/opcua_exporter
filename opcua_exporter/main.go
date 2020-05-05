package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gopcua/opcua"
	opcua_debug "github.com/gopcua/opcua/debug"
	"github.com/gopcua/opcua/monitor"
	"github.com/gopcua/opcua/ua"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var port = flag.Int("port", 9686, "Port to publish metrics on.")
var endpoint = flag.String("endpoint", "opc.tcp://localhost:4096", "OPC UA Endpoint to connect to.")
var promPrefix = flag.String("prom-prefix", "", "Prefix will be appended to emitted prometheus metrics")
var nodeListFile = flag.String("config", "", "Path to a file from which to read the list of OPC UA nodes to monitor")
var configB64 = flag.String("config-b64", "", "Base64-encoded config JSON. Overrides -config")
var debug = flag.Bool("debug", false, "Enable debug logging")
var readTimeout = flag.Duration("read-timeout", 5*time.Second, "Timeout when waiting for OPCUA subscription messages")
var maxTimeouts = flag.Int("max-timeouts", 30, "The exporter will quit trying after this many read timeouts (0 to disable).")

// NodeConfig : Structure for representing OPCUA nodes to monitor.
type NodeConfig struct {
	NodeName   string      // OPC UA node identifier
	MetricName string      // Prometheus metric name to emit
	ExtractBit interface{} // Optional numeric value. If present and positive, extract just this bit and emit it as a boolean metric
}

// MsgHandler interface can convert OPC UA Variant objects
// and emit prometheus metrics
type MsgHandler interface {
	FloatValue(v ua.Variant) (float64, error) // metric value to be emitted
	Handle(v ua.Variant) error                // compute the metric value and publish it
}

// HandlerMap maps OPC UA channel names to MsgHandlers
type HandlerMap map[string]handlerMapRecord

type handlerMapRecord struct {
	config  NodeConfig
	handler MsgHandler
}

var startTime = time.Now()
var uptimeGauge prometheus.Gauge

func init() {
	uptimeGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Subsystem: "opcua_exporter",
		Name:      "uptime_seconds",
		Help:      "Time in seconds since the OPCUA exporter started",
	})
	uptimeGauge.Set(time.Now().Sub(startTime).Seconds())
	prometheus.MustRegister(uptimeGauge)
}

func main() {
	log.Print("Starting up.")
	flag.Parse()
	opcua_debug.Enable = *debug

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var nodes []NodeConfig
	var readError error
	if *configB64 != "" {
		log.Print("Using base64-encoded config")
		nodes, readError = readConfigBase64(configB64)
	} else if *nodeListFile != "" {
		log.Printf("Reading config from %s", *nodeListFile)
		nodes, readError = readConfigFile(*nodeListFile)
	} else {
		log.Fatal("Requires -config or -config-b64")
	}

	if readError != nil {
		log.Fatalf("Error reading config JSON: %v", readError)
	}

	client := getClient(endpoint)
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Error connecting to OPC UA client: %v", err)
	}
	defer client.Close()

	metricMap := createMetrics(&nodes)
	go setupMonitor(ctx, client, &nodes, metricMap)

	http.Handle("/metrics", promhttp.Handler())
	var listenOn = fmt.Sprintf(":%d", *port)
	log.Printf("Serving metrics on %s", listenOn)
	log.Fatal(http.ListenAndServe(listenOn, nil))
}

func getClient(endpoint *string) *opcua.Client {
	client := opcua.NewClient(*endpoint)
	return client
}

// Subscribe to all the nodes and update the appropriate prometheus metrics on change
func setupMonitor(ctx context.Context, client *opcua.Client, nodes *[]NodeConfig, handlerMap HandlerMap) {
	m, err := monitor.NewNodeMonitor(client)
	if err != nil {
		log.Fatal(err)
	}

	var nodeList []string
	for _, node := range *nodes {
		nodeList = append(nodeList, node.NodeName)
	}

	ch := make(chan *monitor.DataChangeMessage, 16)
	params := opcua.SubscriptionParameters{Interval: time.Second}
	sub, err := m.ChanSubscribe(ctx, &params, ch, nodeList...)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup(sub)

	lag := time.Millisecond * 10
	timeoutCount := 0
	for {
		uptimeGauge.Set(time.Now().Sub(startTime).Seconds())
		select {
		case <-ctx.Done():
			return
		case msg := <-ch:
			if msg.Error != nil {
				log.Printf("[channel ] sub=%d error=%s", sub.SubscriptionID(), msg.Error)
			} else if msg.Value == nil {
				log.Printf("nil value received for node %s", msg.NodeID)
			} else {
				log.Printf("[channel ] sub=%d ts=%s node=%s value=%v", sub.SubscriptionID(), msg.SourceTimestamp.UTC().Format(time.RFC3339), msg.NodeID, msg.Value.Value())
				handler := handlerMap[msg.NodeID.String()].handler
				value := msg.Value
				err = handler.Handle(*value)
				if err != nil {
					log.Printf("Error handling opcua value: %s\n", err)
				}
			}
			time.Sleep(lag)
		case <-time.After(*readTimeout):
			timeoutCount++
			log.Printf("Timeout %d wating for subscription messages", timeoutCount)
			if *maxTimeouts > 0 && timeoutCount >= *maxTimeouts {
				log.Fatalf("Max timeouts (%d) exceeded. Quitting.", *maxTimeouts)
			}
		}
	}

}

func cleanup(sub *monitor.Subscription) {
	log.Printf("stats: sub=%d delivered=%d dropped=%d", sub.SubscriptionID(), sub.Delivered(), sub.Dropped())
	sub.Unsubscribe()
}

// Initialize a Prometheus gauge for each node. Return them as a map.
func createMetrics(nodeConfigs *[]NodeConfig) HandlerMap {
	handlerMap := make(HandlerMap)
	for _, nodeConfig := range *nodeConfigs {
		nodeName := nodeConfig.NodeName
		metricName := nodeConfig.MetricName
		handlerMap[nodeName] = handlerMapRecord{nodeConfig, createHandler(nodeConfig)}
		log.Printf("Created prom metric %s for OPC UA node %s", metricName, nodeName)
	}

	return handlerMap
}

func createHandler(nodeConfig NodeConfig) MsgHandler {
	metricName := nodeConfig.MetricName
	if *promPrefix != "" {
		metricName = fmt.Sprintf("%s_%s", *promPrefix, metricName)
	}
	g := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: metricName,
		Help: "From OPC UA",
	})
	prometheus.MustRegister(g)

	var handler MsgHandler
	if nodeConfig.ExtractBit != nil {
		extractBit := int(nodeConfig.ExtractBit.(float64)) // JSON numbers are float64 by default
		handler = OpcuaBitVectorHandler{g, extractBit}
	} else {
		handler = OpcValueHandler{g}
	}
	return handler
}

func readConfigFile(path string) ([]NodeConfig, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(absPath)
	if err != nil {
		return nil, err
	}

	return parseConfigJSON(f)
}

func readConfigBase64(encodedConfig *string) ([]NodeConfig, error) {
	config, decodeErr := base64.StdEncoding.DecodeString(*encodedConfig)
	if decodeErr != nil {
		log.Fatal(decodeErr)
	}
	return parseConfigJSON(bytes.NewReader(config))
}

func parseConfigJSON(config io.Reader) ([]NodeConfig, error) {
	content, err := ioutil.ReadAll(config)
	if err != nil {
		return nil, err
	}

	var nodes []NodeConfig
	err = json.Unmarshal(content, &nodes)
	log.Printf("Found %d nodes in config file.", len(nodes))
	return nodes, err
}
