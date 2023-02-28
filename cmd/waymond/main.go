package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/scriptnull/waymond/requester"
)

func main() {
	// read waymond config
	configFile, err := os.ReadFile("waymond.json")
	if err != nil {
		fmt.Println("unable to read waymond config file", err)
		os.Exit(1)
	}

	// parse config contents
	var config Config
	err = json.Unmarshal(configFile, &config)
	if err != nil {
		fmt.Println("unable to parse waymond config file", err)
	}
	fmt.Println("loaded waymond config successfully")

	// register the auto-scaling requesters
	for idx, requester := range config.Requesters {
		err = requester.Register()
		if err != nil {
			fmt.Println("unable to register the requester:", err)
			os.Exit(1)
		}

		fmt.Printf("registered requester[%d]: (type: %s)\n", idx, requester.Type)
	}
}

type Config struct {
	Requesters []requester.Instance `json:"requesters"`
}
