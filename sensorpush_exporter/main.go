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

var usernameEnvVar = "SENSORPUSH_USERNAME"
var passwordEnvVar = "SENSORPUSH_PASSWORD"

var sensorNameMap SensorNameMap

func main() {
	flag.Parse()

	username, usernameSet := os.LookupEnv(usernameEnvVar)
	password, passwordSet := os.LookupEnv(passwordEnvVar)
	if !usernameSet && !passwordSet {
		log.Fatalf("You must set %s and %s", usernameEnvVar, passwordEnvVar)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := getClient()
	authCtx := getAuthContext(ctx, client, username, password)

	sensorNameMap = *getSensorNameMap(authCtx, client)
	getSamples(authCtx, client, sensorNameMap)

	initMetrics(ctx)
	go pollForSamples(authCtx, client)

	http.Handle("/metrics", promhttp.Handler())
	var listenOn = fmt.Sprintf(":%d", *port)
	log.Printf("Serving metrics on %s", listenOn)
	log.Fatal(http.ListenAndServe(listenOn, nil))
}

var startTime = time.Now()
var uptimeGauge prometheus.Gauge
var temperatureGaugeVec *prometheus.GaugeVec
var humidityGaugeVec *prometheus.GaugeVec

func initMetrics(ctx context.Context) {
	uptimeGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Subsystem: "sensorpush_exporter",
		Name:      "uptime_seconds",
		Help:      "Time in seconds since the OPCUA exporter started",
	})
	uptimeGauge.Set(time.Now().Sub(startTime).Seconds())
	prometheus.MustRegister(uptimeGauge)
	go watchUptime(ctx)

	labelNames := []string{"device_name"}
	temperatureGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "sensorpush_exporter",
		Name:      "temperature_celsius",
		Help:      "Temperature at the sensor",
	}, labelNames)
	prometheus.MustRegister(temperatureGaugeVec)

	humidityGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "sensorpush_exporter",
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

func getClient() *sensorpush.APIClient {
	config := sensorpush.NewConfiguration()
	client := sensorpush.NewAPIClient(config)

	return client
}

func getAuthContext(ctx context.Context, client *sensorpush.APIClient, username string, password string) context.Context {
	authResp, _, err := client.ApiApi.OauthAuthorizePost(ctx, sensorpush.AuthorizeRequest{
		Email:    username,
		Password: password,
	})
	token, _, err := client.ApiApi.AccessToken(ctx, sensorpush.AccessTokenRequest{
		Authorization: authResp.Authorization,
	})
	if err != nil {
		log.Fatal(err)
	}

	authCtx := context.WithValue(ctx, sensorpush.ContextAccessToken, token.Accesstoken)

	return authCtx
}

type SensorNameMap map[string]string

func getSensorNameMap(authCtx context.Context, client *sensorpush.APIClient) *SensorNameMap {
	sensors, _, err := client.ApiApi.Sensors(authCtx, sensorpush.SensorsRequest{})
	if err != nil {
		log.Fatal(err)
	}

	nameMap := make(SensorNameMap)
	for _, v := range sensors {
		nameMap[v.Id] = v.Name
	}
	return &nameMap
}

func getSamples(authCtx context.Context, client *sensorpush.APIClient, sensorNameMap SensorNameMap) map[string]sensorpush.Sample {

	samples, resp, err := client.ApiApi.Samples(authCtx, sensorpush.SamplesRequest{
		Limit: 1,
	})
	if err != nil {
		log.Print("CODE: ", resp.StatusCode)
		log.Fatal(err)
	}

	result := make(map[string]sensorpush.Sample)
	for sensorId, samples := range samples.Sensors {
		sensorName := sensorNameMap[sensorId]
		result[sensorName] = samples[0]
	}
	return result

}

func pollForSamples(authCtx context.Context, client *sensorpush.APIClient) {
	for {
		samples := getSamples(authCtx, client, sensorNameMap)

		for sensorName, sample := range samples {
			labels := prometheus.Labels{
				"device_name": sensorName,
			}
			temperatureGaugeVec.With(labels).Set(float64(sample.Temperature))

			humidityGaugeVec.With(labels).Set(float64(sample.Humidity))
			log.Printf("device_name: %s  temp: %fC  humidity: %f%%", sensorName, sample.Temperature, sample.Humidity)
		}

		time.Sleep(time.Duration(*pollingInterval) * time.Second)
	}
}
