package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sensorpush_exporter/sensorpush"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var port = flag.Int("port", 9687, "Port to publish metrics on.")
var pollingInterval = flag.Int("interval", 60, "Polling interval, in seconds.")

const usernameEnvVar = "SENSORPUSH_USERNAME"
const passwordEnvVar = "SENSORPUSH_PASSWORD"
const promSubsystemName = "sensorpush_exporter" // For labelling prometheus metrics

var sensorNameMap map[string]string // maps sensor IDs to display names

func main() {
	flag.Parse()

	username, usernameSet := os.LookupEnv(usernameEnvVar)
	password, passwordSet := os.LookupEnv(passwordEnvVar)
	if !usernameSet || !passwordSet {
		log.Fatalf("You must set %s and %s", usernameEnvVar, passwordEnvVar)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := getClient()
	authCtx, err := getAuthContext(ctx, client, username, password)
	if err != nil {
		log.Fatal("Error authenticating with Sensorpush: ", err)
	}

	sensorNameMap = getSensorNameMap(authCtx, client)

	initMetrics(ctx)
	go serveMetrics()

	for {
		err := pollForSamples(authCtx, client)
		if err != nil {
			log.Print("Error getting samples: ", err)
		}

		log.Println("Refreshing auth token...")
		authCtx, err = getAuthContext(ctx, client, username, password)
		reauthCounter.Inc()
		if err != nil {
			log.Fatal("Error authenticating with Sensorpush: ", err)
		}
	}

}

var startTime = time.Now()
var uptimeGauge prometheus.Gauge
var temperatureGaugeVec *prometheus.GaugeVec
var humidityGaugeVec *prometheus.GaugeVec
var reauthCounter prometheus.Counter

func initMetrics(ctx context.Context) {
	uptimeGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Subsystem: promSubsystemName,
		Name:      "uptime_seconds",
		Help:      "Time in seconds since the OPCUA exporter started",
	})
	uptimeGauge.Set(time.Now().Sub(startTime).Seconds())
	prometheus.MustRegister(uptimeGauge)
	go watchUptime(ctx)

	reauthCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Subsystem: "sensorpush_exporter",
		Name:      "reauth_count",
		Help:      "Number of times the exporter has refreshed its Sensorpush auth token.",
	})
	prometheus.MustRegister(reauthCounter)

	labelNames := []string{"device_name"}
	temperatureGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: promSubsystemName,
		Name:      "temperature_celsius",
		Help:      "Temperature at the sensor",
	}, labelNames)
	prometheus.MustRegister(temperatureGaugeVec)

	humidityGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: promSubsystemName,
		Name:      "relative_humidity",
		Help:      "Relative humidity at the sensor.",
	}, labelNames)
	prometheus.MustRegister(humidityGaugeVec)
}

func watchUptime(ctx context.Context) {
	for {
		uptimeGauge.Set(time.Now().Sub(startTime).Seconds())
		time.Sleep(time.Second)
	}
}

func serveMetrics() {
	http.Handle("/metrics", promhttp.Handler())
	var listenOn = fmt.Sprintf(":%d", *port)
	log.Printf("Serving metrics on %s", listenOn)
	log.Fatal(http.ListenAndServe(listenOn, nil))
}

func getClient() *sensorpush.APIClient {
	config := sensorpush.NewConfiguration()
	client := sensorpush.NewAPIClient(config)
	return client
}

func getAuthContext(ctx context.Context, client *sensorpush.APIClient, username string, password string) (context.Context, error) {
	authResp, _, err := client.ApiApi.OauthAuthorizePost(ctx, sensorpush.AuthorizeRequest{
		Email:    username,
		Password: password,
	})
	if err != nil {
		return nil, err
	}

	token, _, err := client.ApiApi.AccessToken(ctx, sensorpush.AccessTokenRequest{
		Authorization: authResp.Authorization,
	})
	if err != nil {
		return nil, err
	}

	authCtx := context.WithValue(ctx, sensorpush.ContextAccessToken, token.Accesstoken)

	return authCtx, nil
}

func getSensorNameMap(authCtx context.Context, client *sensorpush.APIClient) map[string]string {
	sensors, _, err := client.ApiApi.Sensors(authCtx, sensorpush.SensorsRequest{})
	if err != nil {
		log.Fatal(err)
	}

	nameMap := make(map[string]string)
	for _, v := range sensors {
		nameMap[v.Id] = v.Name
	}
	return nameMap
}

func getSamples(authCtx context.Context, client *sensorpush.APIClient, sensorNameMap map[string]string) (map[string]sensorpush.Sample, error) {

	samples, resp, err := client.ApiApi.Samples(authCtx, sensorpush.SamplesRequest{
		Limit: 1,
	})
	if err != nil {
		log.Printf("Error from sensorpush, response code %d", resp.StatusCode)
		return nil, err
	}

	result := make(map[string]sensorpush.Sample)
	for sensorID, samples := range samples.Sensors {
		sensorName := sensorNameMap[sensorID]
		result[sensorName] = samples[0]
	}
	return result, nil

}

func pollForSamples(authCtx context.Context, client *sensorpush.APIClient) error {
	for {
		samples, err := getSamples(authCtx, client, sensorNameMap)
		if err != nil {
			return err
		}

		for sensorName, sample := range samples {
			labels := prometheus.Labels{
				"device_name": sensorName,
			}
			temperatureGaugeVec.With(labels).Set(float64(sample.Temperature))

			humidityGaugeVec.With(labels).Set(float64(sample.Humidity))
			log.Printf("device_name: %s\ttemp: %fC\thumidity: %f%%", sensorName, sample.Temperature, sample.Humidity)
		}

		time.Sleep(time.Duration(*pollingInterval) * time.Second)
	}
}
