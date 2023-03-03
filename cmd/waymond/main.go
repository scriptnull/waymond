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
	"github.com/scriptnull/waymond/internal/scaler/docker"
	"github.com/scriptnull/waymond/internal/trigger"
	"github.com/scriptnull/waymond/internal/trigger/cron"
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

	// register all triggers provided by waymond out of the box
	triggerConfigParsers := make(map[trigger.Type]func(*koanf.Koanf) (trigger.Interface, error))
	triggerConfigParsers[cron.Type] = cron.ParseConfig

	// register all scalers provided by waymond out of the box
	sysScalers := make(map[string]any)
	sysScalers[docker.Type] = docker.Scaler{}

	triggerConfigs := k.Slices("trigger")
	triggers := make(map[string]trigger.Interface)
	var errs []error
	for _, triggerConfig := range triggerConfigs {
		ttype := triggerConfig.String("type")
		if ttype == "" {
			errs = append(errs, fmt.Errorf("expected a non-empty 'type' field for trigger: %+v", triggerConfig))
			continue
		}

		id := triggerConfig.String("id")
		if id == "" {
			errs = append(errs, fmt.Errorf("expected a non-empty 'id' field for trigger: %+v", triggerConfig))
			continue
		}

		parseConfig, found := triggerConfigParsers[trigger.Type(ttype)]
		if !found {
			errs = append(errs, fmt.Errorf("unknown 'type' value in trigger: %s in %+v", ttype, triggerConfig))
			continue
		}

		trigger, err := parseConfig(triggerConfig)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		triggers[id] = trigger
	}
	if len(errs) > 0 {
		fmt.Println(errs)
		os.Exit(1)
	}

	var registerErrs []error
	for id, trigger := range triggers {
		fmt.Printf("registering trigger: id:%s type:%s \n", id, trigger.Type())
		err := trigger.Register()
		if err != nil {
			registerErrs = append(registerErrs, err)
		}
	}
	if len(registerErrs) > 0 {
		fmt.Println("error while registering triggers:", registerErrs)
		os.Exit(1)
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
