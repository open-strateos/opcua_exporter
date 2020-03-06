package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/monitor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var port = flag.Int("port", 9100, "Port to publish metrics on.")
var endpoint = flag.String("endpoint", "opc.tcp://localhost:4096", "OPC UA Endpoint to connect to.")

var nodeList = []string{
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

	client := getClient(endpoint)
	if err := client.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	setupMonitor(ctx, client, &nodeList)

	var tempGuage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "foo",
		Help: "A test metric",
	})
	prometheus.MustRegister(tempGuage)
	tempGuage.Set(24)

	http.Handle("/metrics", promhttp.Handler())
	var listenOn = fmt.Sprintf(":%d", *port)
	fmt.Println(fmt.Sprintf("Serving metrics on %s", listenOn))
	log.Fatal(http.ListenAndServe(listenOn, nil))
}

func getClient(endpoint *string) *opcua.Client {
	client := opcua.NewClient(*endpoint)
	return client
}

func setupMonitor(ctx context.Context, client *opcua.Client, nodeList *[]string) {
	m, err := monitor.NewNodeMonitor(client)
	if err != nil {
		log.Fatal(err)
	}

	metricMap := *createMetrics(nodeList)

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
		// case <-ctx.Done():
		// 	return
		case msg := <-ch:
			if msg.Error != nil {
				log.Printf("[channel ] sub=%d error=%s", sub.SubscriptionID(), msg.Error)
			} else {
				log.Printf("[channel ] sub=%d ts=%s node=%s value=%v", sub.SubscriptionID(), msg.SourceTimestamp.UTC().Format(time.RFC3339), msg.NodeID, msg.Value.Value())
				metric := metricMap[msg.NodeID.String()]
				metric.Set(msg.Value.Value().(float64))
			}
			time.Sleep(lag)
		}
	}

}

func cleanup(sub *monitor.Subscription) {
	log.Printf("stats: sub=%d delivered=%d dropped=%d", sub.SubscriptionID(), sub.Delivered(), sub.Dropped())
	sub.Unsubscribe()
}

func createMetrics(nodeList *[]string) *map[string]prometheus.Gauge {
	metricMap := make(map[string]prometheus.Gauge)
	for _, nodeName := range *nodeList {
		g := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: nodeName,
			Help: "From OPC UA",
		})
		prometheus.MustRegister(g)
		metricMap[nodeName] = g
	}

	return &metricMap
}
