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
	"github.com/scriptnull/waymond/internal/scaler"
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

	// track available trigger configuration parsers available out of the box in waymond
	triggerConfigParsers := make(map[trigger.Type]func(*koanf.Koanf) (trigger.Interface, error))
	triggerConfigParsers[cron.Type] = cron.ParseConfig

	// extract triggers from trigger configurations
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

	// track available trigger configuration parsers available out of the box in waymond
	scalerConfigParsers := make(map[scaler.Type]func(*koanf.Koanf) (scaler.Interface, error))
	scalerConfigParsers[docker.Type] = docker.ParseConfig

	// extract scalers from scaler configurations
	scalerConfigs := k.Slices("scaler")
	scalers := make(map[string]scaler.Interface)
	for _, scalerConfig := range scalerConfigs {
		ttype := scalerConfig.String("type")
		if ttype == "" {
			errs = append(errs, fmt.Errorf("expected a non-empty 'type' field for scaler: %+v", scalerConfig))
			continue
		}

		id := scalerConfig.String("id")
		if id == "" {
			errs = append(errs, fmt.Errorf("expected a non-empty 'id' field for scaler: %+v", scalerConfig))
			continue
		}

		parseConfig, found := scalerConfigParsers[scaler.Type(ttype)]
		if !found {
			errs = append(errs, fmt.Errorf("unknown 'type' value in scaler: %s in %+v", ttype, scalerConfig))
			continue
		}

		scaler, err := parseConfig(scalerConfig)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		scalers[id] = scaler
	}
	if len(errs) > 0 {
		fmt.Println(errs)
		os.Exit(1)
	}

	var registerErrs []error

	// register all the triggers in the config
	for id, trigger := range triggers {
		fmt.Printf("starting to register trigger: id:%s type:%s \n", id, trigger.Type())
		err := trigger.Register()
		if err != nil {
			registerErrs = append(registerErrs, err)
		}
		fmt.Printf("registered trigger: id:%s type:%s \n", id, trigger.Type())
	}
	if len(registerErrs) > 0 {
		fmt.Println("error while registering triggers:", registerErrs)
		os.Exit(1)
	}

	// register all the scalers in the config
	for id, scaler := range scalers {
		fmt.Printf("starting to register scaler: id:%s type:%s \n", id, scaler.Type())
		err := scaler.Register()
		if err != nil {
			registerErrs = append(registerErrs, err)
		}
		fmt.Printf("registered scaler: id:%s type:%s \n", id, scaler.Type())
	}
	if len(registerErrs) > 0 {
		fmt.Println("error while registering scalers:", registerErrs)
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
