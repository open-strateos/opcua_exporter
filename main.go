package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/monitor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var port = flag.Int("port", 9100, "Port to publish metrics on.")
var endpoint = flag.String("endpoint", "opc.tcp://localhost:4096", "OPC UA Endpoint to connect to.")
var promPrefix = flag.String("prom-prefix", "", "Prefix will be appended to emitted prometheus metrics")
var nodeListFile = flag.String("file", "", "Path to a file from which to read the list of OPC UA nodes to monitor")

type gaugeMap map[string]prometheus.Gauge

var defaultNodeList = []string{
	"ns=1;s=[L2S2_TMCP]Lift_Station_Consume.Alarms[0]",
	"ns=1;s=[L2S2_TMCP]Lift_Station_Consume.Alarms[1]",
	"ns=1;s=[L2S2_TMCP]Lift_Station_Consume.Alarms[2]",
	"ns=1;s=[L2S2_TMCP]Lift_Station_Consume.Alarms[3]",
}

func main() {
	fmt.Println("Starting up.")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var nodeList []string
	if *nodeListFile == "" {
		nodeList = defaultNodeList
	} else {
		nodeList = readLines(*nodeListFile)
	}

	client := getClient(endpoint)
	if err := client.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	metricMap := createMetrics(&nodeList)
	go setupMonitor(ctx, client, &nodeList, metricMap)

	http.Handle("/metrics", promhttp.Handler())
	var listenOn = fmt.Sprintf(":%d", *port)
	fmt.Println(fmt.Sprintf("Serving metrics on %s", listenOn))
	log.Fatal(http.ListenAndServe(listenOn, nil))
}

func getClient(endpoint *string) *opcua.Client {
	client := opcua.NewClient(*endpoint)
	return client
}

func setupMonitor(ctx context.Context, client *opcua.Client, nodeList *[]string, metricMap gaugeMap) {
	m, err := monitor.NewNodeMonitor(client)
	if err != nil {
		log.Fatal(err)
	}

	ch := make(chan *monitor.DataChangeMessage, 16)
	params := opcua.SubscriptionParameters{Interval: time.Second}
	sub, err := m.ChanSubscribe(ctx, &params, ch, *nodeList...)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup(sub)

	lag := time.Duration(0)
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
				metric.Set(float64(msg.Value.Value().(int32)))
			}
			time.Sleep(lag)
		}
	}

}

func cleanup(sub *monitor.Subscription) {
	log.Printf("stats: sub=%d delivered=%d dropped=%d", sub.SubscriptionID(), sub.Delivered(), sub.Dropped())
	sub.Unsubscribe()
}

func createMetrics(nodeList *[]string) gaugeMap {
	metricMap := make(gaugeMap)
	for _, nodeName := range *nodeList {
		metricName := nodeNameToMetricName(&nodeName)
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

// This assumes a very specific node name format
var nodeNameMatcher = regexp.MustCompile(`s=\[([^\]]+)\]([^;]+)`)

func nodeNameToMetricName(node *string) string {
	match := nodeNameMatcher.FindStringSubmatch(*node)
	if len(match) < 3 {
		log.Fatalf("Unable to parse node name: \"%s\". Is it valid?", *node)
	}
	result := fmt.Sprintf("%s_%s", match[1], match[2])
	for _, sym := range "[]." {
		result = strings.ReplaceAll(result, string(sym), "_")
	}
	result = strings.Trim(result, "_")
	return result
}

func readLines(path string) []string {
	fullPath, err := filepath.Abs(path)
	if err != nil {
		log.Fatal(err)
	}

	file, err := os.Open(fullPath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	lines := make([]string, 0, 4)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if scanner.Err() != nil {
		log.Fatal(scanner.Err())
	}
	return lines
}
