package main

import (
	"bytes"
	"context"
	"encoding/base64"
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
	"gopkg.in/yaml.v2"
)

var port = flag.Int("port", 9686, "Port to publish metrics on.")
var endpoint = flag.String("endpoint", "opc.tcp://localhost:4096", "OPC UA Endpoint to connect to.")
var promPrefix = flag.String("prom-prefix", "", "Prefix will be appended to emitted prometheus metrics")
var nodeListFile = flag.String("config", "", "Path to a file from which to read the list of OPC UA nodes to monitor")
var configB64 = flag.String("config-b64", "", "Base64-encoded config JSON. Overrides -config")
var debug = flag.Bool("debug", false, "Enable debug logging")
var readTimeout = flag.Duration("read-timeout", 5*time.Second, "Timeout when waiting for OPCUA subscription messages")
var maxTimeouts = flag.Int("max-timeouts", 0, "The exporter will quit trying after this many read timeouts (0 to disable).")
var bufferSize = flag.Int("buffer-size", 64, "Maximum number of messages in the receive buffer")
var summaryInterval = flag.Duration("summary-interval", 5*time.Minute, "How frequently to print an event count summary")

// NodeConfig : Structure for representing OPCUA nodes to monitor.
type NodeConfig struct {
	NodeName   string      `yaml:"nodeName"`             // OPC UA node identifier
	MetricName string      `yaml:"metricName"`           // Prometheus metric name to emit
	ExtractBit interface{} `yaml:"extractBit,omitempty"` // Optional numeric value. If present and positive, extract just this bit and emit it as a boolean metric
}

// MsgHandler interface can convert OPC UA Variant objects
// and emit prometheus metrics
type MsgHandler interface {
	FloatValue(v ua.Variant) (float64, error) // metric value to be emitted
	Handle(v ua.Variant) error                // compute the metric value and publish it
}

// HandlerMap maps OPC UA channel names to MsgHandlers
type HandlerMap map[string][]handlerMapRecord

type handlerMapRecord struct {
	config  NodeConfig
	handler MsgHandler
}

var startTime = time.Now()
var uptimeGauge prometheus.Gauge
var messageCounter prometheus.Counter
var eventSummaryCounter *EventSummaryCounter

func init() {
	subsystem := "opcua_exporter"
	uptimeGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Subsystem: subsystem,
		Name:      "uptime_seconds",
		Help:      "Time in seconds since the OPCUA exporter started",
	})
	uptimeGauge.Set(time.Now().Sub(startTime).Seconds())
	prometheus.MustRegister(uptimeGauge)

	messageCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Subsystem: subsystem,
		Name:      "message_count",
		Help:      "Total number of OPCUA channel updates received by the exporter",
	})
	prometheus.MustRegister(messageCounter)

	eventSummaryCounter = NewEventSummaryCounter(*summaryInterval)
}

func main() {
	log.Print("Starting up.")
	flag.Parse()
	opcua_debug.Enable = *debug

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	eventSummaryCounter.Start(ctx)

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
	log.Printf("Connecting to OPCUA server at %s", *endpoint)
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Error connecting to OPC UA client: %v", err)
	} else {
		log.Print("Connected successfully")
	}
	defer client.Close()

	metricMap := createMetrics(&nodes)
	go setupMonitor(ctx, client, metricMap, *bufferSize)

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
func setupMonitor(ctx context.Context, client *opcua.Client, handlerMap HandlerMap, bufferSize int) {
	m, err := monitor.NewNodeMonitor(client)
	if err != nil {
		log.Fatal(err)
	}

	var nodeList []string
	for nodeName := range handlerMap { // Node names are keys of handlerMap
		nodeList = append(nodeList, nodeName)
	}

	ch := make(chan *monitor.DataChangeMessage, bufferSize)
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
				log.Printf("[error ] sub=%d error=%s", sub.SubscriptionID(), msg.Error)
			} else if msg.Value == nil {
				log.Printf("nil value received for node %s", msg.NodeID)
			} else {
				if *debug {
					log.Printf("[message ] sub=%d ts=%s node=%s value=%v", sub.SubscriptionID(), msg.SourceTimestamp.UTC().Format(time.RFC3339), msg.NodeID, msg.Value.Value())
				}

				messageCounter.Inc()
				nodeID := msg.NodeID.String()
				eventSummaryCounter.Inc(nodeID)

				handleMessage(msg, handlerMap)
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

func handleMessage(msg *monitor.DataChangeMessage, handlerMap HandlerMap) {
	nodeID := msg.NodeID.String()
	for _, handlerMapRec := range handlerMap[nodeID] {
		handler := handlerMapRec.handler
		value := msg.Value
		if *debug {
			log.Printf("Handling %s --> %s", nodeID, handlerMapRec.config.MetricName)
		}
		err := handler.Handle(*value)
		if err != nil {
			log.Printf("Error handling opcua value: %s (%s)\n", err, handlerMapRec.config.MetricName)
		}
	}
}

// Initialize a Prometheus gauge for each node. Return them as a map.
func createMetrics(nodeConfigs *[]NodeConfig) HandlerMap {
	handlerMap := make(HandlerMap)
	for _, nodeConfig := range *nodeConfigs {
		nodeName := nodeConfig.NodeName
		metricName := nodeConfig.MetricName
		mapRecord := handlerMapRecord{nodeConfig, createHandler(nodeConfig)}
		handlerMap[nodeName] = append(handlerMap[nodeName], mapRecord)
		log.Printf("Created prom metric %s for OPC UA node %s", metricName, nodeName)
	}
	log.Printf("Registered %d handlers for %d unique UPCUA nodes", len(nodeConfigs), len(handlerMap))

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
		extractBit := nodeConfig.ExtractBit.(int) // coerce interface to an integer
		handler = OpcuaBitVectorHandler{g, extractBit, *debug}
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

	return parseConfigYAML(f)
}

func readConfigBase64(encodedConfig *string) ([]NodeConfig, error) {
	config, decodeErr := base64.StdEncoding.DecodeString(*encodedConfig)
	if decodeErr != nil {
		log.Fatal(decodeErr)
	}
	return parseConfigYAML(bytes.NewReader(config))
}

func parseConfigYAML(config io.Reader) ([]NodeConfig, error) {
	content, err := ioutil.ReadAll(config)
	if err != nil {
		return nil, err
	}

	var nodes []NodeConfig
	err = yaml.Unmarshal(content, &nodes)
	log.Printf("Found %d nodes in config file.", len(nodes))
	return nodes, err
}
