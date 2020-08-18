package main

import (
	"log"
	"sensorpush_exporter/sensorpush"
)

func main() {
	config := sensorpush.NewConfiguration()
	client := sensorpush.NewAPIClient(config)
	log.Println(client)
}
