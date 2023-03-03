package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

var configPath string
var k = koanf.New(".")

func main() {
	// set command line flags
	flag.StringVar(&configPath, "config", "", "file path to waymond config file (.toml)")
	flag.Parse()
	if configPath == "" {
		configPath = "waymond.toml"
	}

	// read waymond config file
	if err := k.Load(file.Provider(configPath), toml.Parser()); err != nil {
		fmt.Println("error loading config:", err)
		os.Exit(1)
	}

	// var config Config
	triggers := k.Slices("trigger")
	for _, trigger := range triggers {
		fmt.Printf("trigger: type = %s, id = %s \n", trigger.String("type"), trigger.String("id"))
	}

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
