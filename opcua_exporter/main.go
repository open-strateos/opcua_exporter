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
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var port = flag.Int("port", 9686, "Port to publish metrics on.")
var endpoint = flag.String("endpoint", "opc.tcp://localhost:4096", "OPC UA Endpoint to connect to.")
var promPrefix = flag.String("prom-prefix", "", "Prefix will be appended to emitted prometheus metrics")
var nodeListFile = flag.String("config", "", "Path to a file from which to read the list of OPC UA nodes to monitor")
var configB64 = flag.String("config-b64", "", "Base64-encoded config JSON. Overrides -config")
var debug = flag.Bool("debug", false, "Enable debug logging")

// Maps OPC UA channel names to prometheus Gauge instances
type gaugeMap map[string]prometheus.Gauge

// Structure for representing OPCUA nodes to monitor.
type Node struct {
	NodeName   string // OPC UA node identifier
	MetricName string // Prometheus metric name to emit
}

func main() {
	fmt.Println("Starting up.")
	flag.Parse()
	opcua_debug.Enable = *debug

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var nodes []Node
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
	fmt.Println(fmt.Sprintf("Serving metrics on %s", listenOn))
	log.Fatal(http.ListenAndServe(listenOn, nil))
}

func getClient(endpoint *string) *opcua.Client {
	client := opcua.NewClient(*endpoint)
	return client
}

// Subscribe to all the nodes and update the appropriate prometheus metrics on change
func setupMonitor(ctx context.Context, client *opcua.Client, nodes *[]Node, metricMap gaugeMap) {
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
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-ch:
			if msg.Error != nil {
				log.Printf("[channel ] sub=%d error=%s", sub.SubscriptionID(), msg.Error)
			} else {
				log.Printf("[channel ] sub=%d ts=%s node=%s value=%v", sub.SubscriptionID(), msg.SourceTimestamp.UTC().Format(time.RFC3339), msg.NodeID, msg.Value.Value())
				metric := metricMap[msg.NodeID.String()]
				value := msg.Value.Value()
				var floatVal float64
				switch v := value.(type) {
				case bool:
					if value.(bool) {
						floatVal = 1.0
					} else {
						floatVal = 0.0
					}
				case int32:
					floatVal = float64(value.(int32))
				default:
					log.Printf("Node %s has unhandled type %T", msg.NodeID.String(), v)
					continue
				}
				metric.Set(floatVal)
			}
			time.Sleep(lag)
		}
	}

}

func cleanup(sub *monitor.Subscription) {
	log.Printf("stats: sub=%d delivered=%d dropped=%d", sub.SubscriptionID(), sub.Delivered(), sub.Dropped())
	sub.Unsubscribe()
}

// Initialize a Prometheus gauge for each node. Return them as a map.
func createMetrics(nodeList *[]Node) gaugeMap {
	metricMap := make(gaugeMap)
	for _, node := range *nodeList {
		nodeName := node.NodeName
		metricName := node.MetricName
		if *promPrefix != "" {
			metricName = fmt.Sprintf("%s_%s", *promPrefix, metricName)
		}
		g := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: metricName,
			Help: "From OPC UA",
		})
		prometheus.MustRegister(g)
		metricMap[nodeName] = g
		log.Printf("Created prom metric %s for OPC UA node %s", metricName, nodeName)
	}

	return metricMap
}

func readConfigFile(path string) ([]Node, error) {
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

func readConfigBase64(encodedConfig *string) ([]Node, error) {
	config, decodeErr := base64.StdEncoding.DecodeString(*encodedConfig)
	if decodeErr != nil {
		log.Fatal(decodeErr)
	}
	return parseConfigJSON(bytes.NewReader(config))
}

func parseConfigJSON(config io.Reader) ([]Node, error) {
	content, err := ioutil.ReadAll(config)
	if err != nil {
		return nil, err
	}

	var nodes []Node
	err = json.Unmarshal(content, &nodes)
	log.Printf("Found %d nodes in config file.", len(nodes))
	return nodes, err
}
