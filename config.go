package main

import (
	"encoding/json"
	"io/ioutil"

	log "github.com/golang/glog"
)

type Configuration struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func LoadConfig(path string) Configuration {
	log.Infof("Loading configuration from %s\n", path)

	contents, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Unable to read config file: %s", err)
	}

	var config Configuration

	err = json.Unmarshal(contents, &config)
	if err != nil {
		log.Fatalf("Malformed config file: %s", err)
	}

	return config
}
