package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type (
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
		Type   string
		Stages []Stage `json:"stages"`
	}

	Config struct {
		Actions  map[string]Action  `json:"actions"`
		Services map[string]Service `json:"services"`
		DBPath   string             `json:"dbpath"`
	}
)

var (
	// Valid actions and their associated type
	valid_actions map[string]string = map[string]string{
		"create":           "async",
		"delete":           "async",
		"reboot":           "async",
		"restart":          "async",
		"poweroff":         "async",
		"shutdown":         "async",
		"run":              "async",
		"cpuMetrics":       "info",
		"nicMetrics":       "info",
		"diskMetrics":      "info",
		"listImages":       "info",
		"getImage":         "info",
		"deleteImage":      "async",
		"fetchImage":       "async",
		"listSnapshots":    "info",
		"getSnapshot":      "info",
		"createSnapshot":   "async",
		"deleteSnapshot":   "async",
		"rollbackSnapshot": "async",
		"downloadSnapshot": "stream",
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
		if _, ok := valid_actions[name]; !ok {
			return fmt.Errorf("action %s is not a valid action", name)
		}

		for _, s := range action.Stages {
			if err := s.validate(name); err != nil {
				return err
			}
		}

		action.Type = valid_actions[name]
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
