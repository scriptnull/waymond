package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/scriptnull/waymond/internal/connector"
	"github.com/scriptnull/waymond/internal/connector/direct"
	"github.com/scriptnull/waymond/internal/event"
	"github.com/scriptnull/waymond/internal/log"
	"github.com/scriptnull/waymond/internal/scaler"
	"github.com/scriptnull/waymond/internal/scaler/docker"
	"github.com/scriptnull/waymond/internal/scaler/noop"
	"github.com/scriptnull/waymond/internal/trigger"
	"github.com/scriptnull/waymond/internal/trigger/buildkite"
	"github.com/scriptnull/waymond/internal/trigger/cron"
)

var configPath string
var k = koanf.New(".")
var corelog = log.New("waymond.core")

func main() {
	// set command line flags
	flag.StringVar(&configPath, "config", "", "file path to waymond config file (.toml)")
	flag.Parse()
	if configPath == "" {
		configPath = "waymond.toml"
	}

	// read waymond config file
	if err := k.Load(file.Provider(configPath), toml.Parser()); err != nil {
		corelog.Error("error loading config:", err)
		os.Exit(1)
	}

	// track available trigger configuration parsers available out of the box in waymond
	triggerConfigParsers := make(map[trigger.Type]func(*koanf.Koanf) (trigger.Interface, error))
	triggerConfigParsers[cron.Type] = cron.ParseConfig
	triggerConfigParsers[buildkite.Type] = buildkite.ParseConfig

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
		corelog.Error(errs)
		os.Exit(1)
	}

	// track available trigger configuration parsers available out of the box in waymond
	scalerConfigParsers := make(map[scaler.Type]func(*koanf.Koanf) (scaler.Interface, error))
	scalerConfigParsers[docker.Type] = docker.ParseConfig
	scalerConfigParsers[noop.Type] = noop.ParseConfig

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

	// track available connector configuration parsers available out of the box in waymond
	connectorConfigParsers := make(map[connector.Type]func(*koanf.Koanf) (connector.Interface, error))
	connectorConfigParsers[direct.Type] = direct.ParseConfig

	// extract connector from connector configurations
	connectorConfigs := k.Slices("connect")
	connectors := make(map[string]connector.Interface)
	for _, connectorConfig := range connectorConfigs {
		ttype := connectorConfig.String("type")
		if ttype == "" {
			errs = append(errs, fmt.Errorf("expected a non-empty 'type' field for connector: %+v", connectorConfig))
			continue
		}

		id := connectorConfig.String("id")
		if id == "" {
			errs = append(errs, fmt.Errorf("expected a non-empty 'id' field for connector: %+v", connectorConfig))
			continue
		}

		parseConfig, found := connectorConfigParsers[connector.Type(ttype)]
		if !found {
			errs = append(errs, fmt.Errorf("unknown 'type' value in connector: %s in %+v", ttype, connectorConfig))
			continue
		}

		connector, err := parseConfig(connectorConfig)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		connectors[id] = connector
	}
	if len(errs) > 0 {
		corelog.Error(errs)
		os.Exit(1)
	}

	ctx := context.Background()

	err := event.Init()
	if err != nil {
		corelog.Error("error initializing the event bus", err)
		os.Exit(1)
	}

	var registerErrs []error

	// register all the triggers in the config
	for id, trigger := range triggers {
		corelog.Verbosef("starting to register trigger: id:%s type:%s \n", id, trigger.Type())
		err := trigger.Register(ctx)
		if err != nil {
			registerErrs = append(registerErrs, err)
		}
		corelog.Verbosef("registered trigger: id:%s type:%s \n", id, trigger.Type())
	}
	if len(registerErrs) > 0 {
		corelog.Error("error while registering triggers:", registerErrs)
		os.Exit(1)
	}

	// register all the scalers in the config
	for id, scaler := range scalers {
		corelog.Verbosef("starting to register scaler: id:%s type:%s \n", id, scaler.Type())
		err := scaler.Register(ctx)
		if err != nil {
			registerErrs = append(registerErrs, err)
		}
		corelog.Verbosef("registered scaler: id:%s type:%s \n", id, scaler.Type())
	}
	if len(registerErrs) > 0 {
		corelog.Error("error while registering scalers:", registerErrs)
		os.Exit(1)
	}

	// register all the connectors in the config
	for id, connector := range connectors {
		corelog.Verbosef("starting to register connector: id:%s type:%s \n", id, connector.Type())
		err := connector.Register(ctx)
		if err != nil {
			registerErrs = append(registerErrs, err)
		}
		corelog.Verbosef("registered connector: id:%s type:%s \n", id, connector.Type())
	}
	if len(registerErrs) > 0 {
		corelog.Error("error while registering connectors:", registerErrs)
		os.Exit(1)
	}

	// wait for signals to quit program
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)
	go func() {
		sig := <-sigs
		corelog.Verbose("received signal", sig)
		done <- true
	}()
	corelog.Verbose("started waymond successfully")
	corelog.Verbose("press CTRL+C if you would like to quit")
	<-done
	corelog.Verbose("stopped waymond")
}
