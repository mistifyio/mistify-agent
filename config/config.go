package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type (
	ActionType int

	Service struct {
		MaxPending uint   `json:"max_pending"`
		Port       uint   `json:"port"`
		Path       string `json:"path"`
	}

	Stage struct {
		Service string            `json:"service"`
		Method  string            `json:"method"`
		Args    map[string]string `json:"args"`
	}

	Action struct {
		Type   ActionType
		Stages []Stage `json:"stages"`
	}

	Config struct {
		Actions  map[string]Action  `json:"actions"`
		Services map[string]Service `json:"services"`
		DBPath   string             `json:"dbpath"`
	}
)

const (
	InfoAction ActionType = iota
	StreamAction
	AsyncAction
)

var (
	// Valid actions and their associated type
	ValidActions map[string]ActionType = map[string]ActionType{
		"create":           AsyncAction,
		"delete":           AsyncAction,
		"reboot":           AsyncAction,
		"restart":          AsyncAction,
		"poweroff":         AsyncAction,
		"shutdown":         AsyncAction,
		"run":              AsyncAction,
		"cpuMetrics":       InfoAction,
		"nicMetrics":       InfoAction,
		"diskMetrics":      InfoAction,
		"listImages":       InfoAction,
		"getImage":         InfoAction,
		"deleteImage":      AsyncAction,
		"fetchImage":       AsyncAction,
		"listSnapshots":    InfoAction,
		"getSnapshot":      InfoAction,
		"createSnapshot":   AsyncAction,
		"deleteSnapshot":   AsyncAction,
		"rollbackSnapshot": AsyncAction,
		"downloadSnapshot": StreamAction,
	}
)

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
