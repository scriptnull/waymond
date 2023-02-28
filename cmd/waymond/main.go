package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/scriptnull/waymond/requester"
	"github.com/scriptnull/waymond/schedule"
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

	// start global schedulers
	schedule.CronScheduler.Start()

	// wait for signals to quit program
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)
	go func() {
		sig := <-sigs
		fmt.Println("received signal", sig)
		done <- true
	}()
	fmt.Println("started waymond successfully")
	fmt.Println("press CTRL+C if you would like to quit")
	<-done
	fmt.Println("stopped waymond")
}

type Config struct {
	Requesters []requester.Instance `json:"requesters"`
}
