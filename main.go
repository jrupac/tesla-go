package main

import (
	"flag"
	"log"
)

const (
	version = "0.01"
)

var config_path = flag.String("config_path", "config.json", "Path to configuration file.")

func main() {
	log.Printf("Tesla Go version %v\n", version)

	config := LoadConfig(*config_path)
	client := NewTeslaClient()

	if err := client.Authenticate(config); err != nil {
		log.Fatalf("Failed to authenticate: %s", err)
	} else {
		log.Println("Authentication complete.")
	}

	vehicles, err := client.ListVehicles()
	if err != nil {
		log.Fatalf("Operation failed: %s", err)
	}

	for _, vehicle := range vehicles {
		resp, err := client.GetVehicleState(vehicle)
		if err != nil {
			log.Fatalf("Error while getting drive state of vehicle %s: %s", vehicle.ID, err)
		}
		log.Printf("%d: %+v\n", vehicle.ID, resp)
	}
}
