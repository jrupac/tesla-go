package main

import (
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"log"
	"strconv"
)

const (
	version = "0.01"
	jobName = "tesla_collector"
)

var (
	configPath      = flag.String("configPath", "config.json", "Path to configuration file.")
	pushgatewayAddr = flag.String("pushgateway_address", "", "URL of pushgateway.")
	odometerMetric  = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "odometric_metric",
		Help: "Current odometer value per vehicle.",
	}, []string{"vehicle_id", "vehicle_name"})
)

func main() {
	flag.Parse()
	log.Printf("Tesla Go version %v\n", version)

	config := LoadConfig(*configPath)
	client := NewTeslaClient()

	if err := client.Authenticate(config); err != nil {
		log.Fatalf("Failed to authenticate: %s", err)
	} else {
		log.Println("Authentication complete.")
	}

	vehicles, err := client.ListVehicles()
	if err != nil {
		log.Fatalf("ListVehicles operation failed: %s", err)
	}

	for _, v := range vehicles {
		resp, err := client.GetVehicleState(v)
		if err != nil {
			log.Fatalf("Error while getting drive state of v %s: %s", v.ID, err)
		}
		odometerMetric.WithLabelValues(strconv.FormatInt(v.ID, 10), v.DisplayName).Set(resp.Odometer)
	}

	if err := push.Collectors(jobName, nil, *pushgatewayAddr, odometerMetric); err != nil {
		log.Fatalf("Failed to push metrics: %s", err)
	}
}
