package main

import (
	"flag"
	log "github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"strconv"
)

const (
	version = "0.01"
	jobName = "tesla_collector"
)

var (
	configPath      = flag.String("config_path", "config.json", "Path to configuration file.")
	pushgatewayAddr = flag.String("pushgateway_address", "", "URL of pushgateway.")
)

var (
	odometerMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "odometric_metric",
		Help: "Current odometer value per vehicle.",
	}, []string{"vehicle_id", "vehicle_name"})
	firmwareMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "firmware_metric",
		Help: "Current firmware version per vehicle.",
	}, []string{"vehicle_id", "vehicle_name", "version"})
	batteryRangeMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "battery_range_metric",
		Help: "Reported battery range per vehicle.",
	}, []string{"vehicle_id", "vehicle_name"})
	batteryLevelMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "battery_level_metric",
		Help: "Reported battery charge percentage per vehicle.",
	}, []string{"vehicle_id", "vehicle_name"})
)

func main() {
	flag.Parse()
	defer log.Flush()

	log.Infof("Tesla Go version %v\n", version)

	config := LoadConfig(*configPath)
	client := NewTeslaClient()

	if err := client.Authenticate(config); err != nil {
		log.Fatalf("Failed to authenticate: %s", err)
	} else {
		log.Infof("Authentication complete.")
	}

	vehicles, err := client.ListVehicles()
	if err != nil {
		log.Fatalf("ListVehicles operation failed: %s", err)
	}

	for _, v := range vehicles {
		vid := strconv.FormatInt(v.ID, 10)
		vehicleState, err := client.GetVehicleState(v)
		if err != nil {
			log.Fatalf("Error while getting vehicle state of v %s: %s", v.ID, err)
		}
		chargeState, err := client.GetChargeState(v)
		if err != nil {
			log.Fatalf("Error while getting charge state of v %s: %s", v.ID, err)
		}
		odometerMetric.WithLabelValues(vid, v.DisplayName).Set(vehicleState.Odometer)
		firmwareMetric.WithLabelValues(vid, v.DisplayName, vehicleState.FirmwareVersion).Set(1)
		batteryRangeMetric.WithLabelValues(vid, v.DisplayName).Set(chargeState.BatteryRange)
		batteryLevelMetric.WithLabelValues(vid, v.DisplayName).Set(float64(chargeState.BatteryLevel))
	}

	err = push.Collectors(
		jobName, nil, *pushgatewayAddr,
		odometerMetric, firmwareMetric, batteryRangeMetric, batteryLevelMetric)
	if err != nil {
		log.Fatalf("Failed to push metrics: %s", err)
	} else {
		log.Infof("Sucessfully pushed metrics.")
	}
}
