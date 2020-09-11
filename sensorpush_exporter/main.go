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

//Command-line Flags
var port = flag.Int("port", 9687, "Port to publish metrics on.")
var pollingInterval = flag.Int("interval", 60, "Polling interval, in seconds.")
var sensorNameRefreshInterval = flag.Int("name-refresh-interval", 5*60, "How frequently to automatically refresh the sensor names table, in seconds")

// Constants
const usernameEnvVar = "SENSORPUSH_USERNAME"
const passwordEnvVar = "SENSORPUSH_PASSWORD"
const promSubsystemName = "sensorpush_exporter" // For labelling prometheus metrics

// Global vars
var startTime = time.Now()
var globalAuthCtx *context.Context // holds auth token for sensorpush

var sensorNameMap map[string]string      // maps sensor IDs to display names
var sensorNamesRefresh = make(chan bool) // send to this channel to force-refresh the sensor names
var sensorNamesReady = make(chan bool)   // signal that sensor names are ready after a forced refresh

// Prometheus Metrics
var uptimeGauge prometheus.Gauge
var temperatureGaugeVec *prometheus.GaugeVec
var humidityGaugeVec *prometheus.GaugeVec
var reauthCounter prometheus.Counter
var numberOfSensors prometheus.Gauge

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
		Subsystem: promSubsystemName,
		Name:      "reauth_count",
		Help:      "Number of times the exporter has refreshed its Sensorpush auth token.",
	})
	prometheus.MustRegister(reauthCounter)

	numberOfSensors = prometheus.NewGauge(prometheus.GaugeOpts{
		Subsystem: promSubsystemName,
		Name:      "number_of_sensors",
		Help:      "Number of sensors being monitored",
	})
	prometheus.MustRegister(numberOfSensors)

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

// Update the global auth context by fetching a new token
// This allows us to share an auth context between pollforSamples() and sensorNameRefreshLoop(),
// at the cost of having some global state. I'm not sure this is ideal, but can't think
// of another way do it without each goroutine managing its own token.
func authenticateGlobal(client *sensorpush.APIClient, username string, password string) {
	authCtx, err := getAuthContext(context.Background(), client, username, password)
	if err != nil {
		log.Fatal("Unable to authenticate: ", err)
	}
	log.Print("Authentication success.")
	globalAuthCtx = authCtx
	reauthCounter.Inc()
}

func getAuthContext(ctx context.Context, client *sensorpush.APIClient, username string, password string) (*context.Context, error) {
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

	return &authCtx, nil
}

// Refresh the sensor name map periodically,
// or when a signal is received on the sensorNamesRefresh channel.
// A triggered refresh will reset the timer.
func sensorNameRefreshLoop(client *sensorpush.APIClient, interval time.Duration) {
	var tmpSensorNameMap map[string]string
	var err error
	var triggered = false
	for {

		// Block waiting for a trigger or timer
		select {
		case <-sensorNamesRefresh:
			triggered = true
			log.Println("Refreshing the sensor name map (triggered)")
		case <-time.After(interval):
			triggered = false
			log.Printf("Refreshing the sensor name map (scheduled)")
		}

		tmpSensorNameMap, err = getSensorNameMap(*globalAuthCtx, client)
		if err == nil {
			sensorNameMap = tmpSensorNameMap
		} else {
			log.Printf("Unable to refresh sensor names. Main loop should reauth before the next attempt.")
		}
		if triggered {
			sensorNamesReady <- true
		}
		numberOfSensors.Set(float64(len(sensorNameMap)))
	}
}

func getSensorNameMap(authCtx context.Context, client *sensorpush.APIClient) (map[string]string, error) {
	sensors, _, err := client.ApiApi.Sensors(authCtx, sensorpush.SensorsRequest{})
	if err != nil {
		return nil, err
	}

	nameMap := make(map[string]string)
	for _, v := range sensors {
		nameMap[v.Id] = v.Name
	}
	return nameMap, nil
}

func getSamples(authCtx *context.Context, client *sensorpush.APIClient, sensorNameMap map[string]string) (map[string]sensorpush.Sample, error) {

	samples, resp, err := client.ApiApi.Samples(*authCtx, sensorpush.SamplesRequest{
		Limit: 1,
	})
	if err != nil {
		log.Printf("Error from sensorpush, response code %d", resp.StatusCode)
		return nil, err
	}

	result := make(map[string]sensorpush.Sample)
	for sensorID, samples := range samples.Sensors {

		// If an unknown sensor ID is encountered, force-refresh the name map
		if _, known := sensorNameMap[sensorID]; !known {
			sensorNamesRefresh <- true
			<-sensorNamesReady
		}

		sensorName := sensorNameMap[sensorID]
		result[sensorName] = samples[0]
	}
	return result, nil

}

func fahrenheitToCelcius(tempF float32) float64 {
	return float64(tempF-32) / 1.8
}

// Update prometheus metrics
func updateMetrics(samples map[string]sensorpush.Sample) {
	for sensorName, sample := range samples {
		labels := prometheus.Labels{
			"device_name": sensorName,
		}

		temperatureCelcius := fahrenheitToCelcius(sample.Temperature)
		humidityPct := float64(sample.Humidity)
		temperatureGaugeVec.With(labels).Set(temperatureCelcius)
		humidityGaugeVec.With(labels).Set(humidityPct)

		log.Printf("device_name: %s\ttemp: %fC\thumidity: %f%%", sensorName, temperatureCelcius, humidityPct)
	}
}

func main() {
	flag.Parse()

	username, usernameSet := os.LookupEnv(usernameEnvVar)
	password, passwordSet := os.LookupEnv(passwordEnvVar)
	if !usernameSet || !passwordSet {
		log.Fatalf("You must set %s and %s", usernameEnvVar, passwordEnvVar)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Println("Initializing metrics")
	initMetrics(ctx)
	go serveMetrics()

	client := getClient()
	log.Println("Authenticating...")
	authenticateGlobal(client, username, password)

	go sensorNameRefreshLoop(client, time.Duration(*sensorNameRefreshInterval)*time.Second)
	// Trigger and wait for an initial fetch
	sensorNamesRefresh <- true
	<-sensorNamesReady

	// Sample polling loop
	for {
		samples, err := getSamples(globalAuthCtx, client, sensorNameMap)
		if err != nil {
			log.Print("Failed to fetch samples: ", err)
			log.Print("Reauthenticating...")
			authenticateGlobal(client, username, password)
			continue
		}
		updateMetrics(samples)

		time.Sleep(time.Duration(*pollingInterval) * time.Second)
	}
}
