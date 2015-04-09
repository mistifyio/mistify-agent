// Package config handles agent configuration from a config file. It also
// defines the valid actions along with their types.
package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type (
	// ActionType describes synchronicity of an action
	ActionType int

	// Service is an HTTP service
	Service struct {
		MaxPending uint   `json:"max_pending"`
		Port       uint   `json:"port"`
		Path       string `json:"path"`
	}

	// Stage is a single step for an action
	Stage struct {
		Service string            `json:"service"`
		Method  string            `json:"method"`
		Args    map[string]string `json:"args"`
	}

	// Action is a set of stages and how they should be handled
	Action struct {
		Type   ActionType
		Stages []Stage `json:"stages"`
	}

	// Config contains all of the configuration data
	Config struct {
		Actions  map[string]Action  `json:"actions"`
		Services map[string]Service `json:"services"`
		DBPath   string             `json:"dbpath"`
	}
)

const (
	// InfoAction is for synchronous information retrieval in JSON format
	InfoAction ActionType = iota
	// StreamAction is for synchronous data streaming
	StreamAction
	// AsyncAction is for asynchronous actions
	AsyncAction
)

var (
	// ValidActions is a whitelist of configurable actions and their types
	ValidActions = map[string]ActionType{
		"create":               AsyncAction,
		"containerCreate":      AsyncAction,
		"delete":               AsyncAction,
		"containerDelete":      AsyncAction,
		"containerStart":       AsyncAction,
		"reboot":               AsyncAction,
		"containerReboot":      AsyncAction,
		"restart":              AsyncAction,
		"containerRestart":     AsyncAction,
		"poweroff":             AsyncAction,
		"containerPoweroff":    AsyncAction,
		"shutdown":             AsyncAction,
		"containerShutdown":    AsyncAction,
		"run":                  AsyncAction,
		"cpuMetrics":           InfoAction,
		"nicMetrics":           InfoAction,
		"diskMetrics":          InfoAction,
		"listImages":           InfoAction,
		"containerListImages":  InfoAction,
		"getImage":             InfoAction,
		"containerGetImage":    InfoAction,
		"deleteImage":          AsyncAction,
		"containerDeleteImage": AsyncAction,
		"fetchImage":           AsyncAction,
		"containerFetchImage":  AsyncAction,
		"listSnapshots":        InfoAction,
		"getSnapshot":          InfoAction,
		"createSnapshot":       AsyncAction,
		"deleteSnapshot":       AsyncAction,
		"rollbackSnapshot":     AsyncAction,
		"downloadSnapshot":     StreamAction,
	}
)

// NewConfig creates a new Config
func NewConfig() *Config {
	c := &Config{
		Actions:  make(map[string]Action),
		Services: make(map[string]Service),
		DBPath:   "/tmp/mistify-agent.db",
	}

	return c
}

func (stage *Stage) validate(prefix string) error {
	if stage == nil {
		return nil
	}
	if stage.Method == "" {
		return fmt.Errorf("%s: method cannot be empty", prefix)
	}
	if stage.Service == "" {
		return fmt.Errorf("%s: service cannot be empty", prefix)
	}
	return nil
}

// AddConfig loads a configuration file
func (c *Config) AddConfig(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	newConfig := Config{}
	err = json.Unmarshal(data, &newConfig)
	if err != nil {
		return err
	}

	for name, service := range newConfig.Services {
		if _, ok := c.Services[name]; ok {
			return fmt.Errorf("service %s has already been defined", name)
		}
		if service.Port <= 0 {
			return fmt.Errorf("service %s has no port", name)
		}

		if service.MaxPending == 0 {
			service.MaxPending = 4
		}
		c.Services[name] = service

	}

	for name, action := range newConfig.Actions {
		if _, ok := c.Actions[name]; ok {
			return fmt.Errorf("action %s has already been defined", name)
		}
		if _, ok := ValidActions[name]; !ok {
			return fmt.Errorf("action %s is not a valid action", name)
		}

		for _, s := range action.Stages {
			if err := s.validate(name); err != nil {
				return err
			}
		}

		action.Type = ValidActions[name]
		c.Actions[name] = action
	}

	return nil
}

// Fixup does a bit of validation and initializtion
func (c *Config) Fixup() error {
	for name, action := range c.Actions {
		for _, stage := range action.Stages {
			if _, ok := c.Services[stage.Service]; !ok {
				return fmt.Errorf("%s unable to find service %s", name, stage.Service)
			}
			if stage.Args == nil {
				stage.Args = make(map[string]string)
			}
		}
	}

	// TODO: add builtins for create and delete

	return nil
}
