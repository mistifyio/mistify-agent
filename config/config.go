package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type (
	Service struct {
		MaxPending uint `json:"max_pending"`
		Port       uint `json:"port"`
	}

	Stage struct {
		Service string            `json:"service"`
		Method  string            `json:"method"`
		Args    map[string]string `json:"args"`
	}

	Action struct {
		Sync  []Stage `json:"sync"`
		Async []Stage `json:"async"`
	}

	Config struct {
		Actions  map[string]Action  `json:"actions"`
		Services map[string]Service `json:"services"`
		Metrics  map[string]Stage   `json:"metrics"`
		DBPath   string             `json:"dbpath"`
	}
)

var (
	valid_actions map[string]bool = map[string]bool{
		"create":   true,
		"delete":   true,
		"reboot":   true,
		"restart":  true,
		"poweroff": true,
		"shutdown": true,
		"run":      true,
	}

	valid_metrics map[string]bool = map[string]bool{
		"cpu":  true,
		"nic":  true,
		"disk": true,
	}
)

func NewConfig() *Config {
	c := &Config{
		Actions:  make(map[string]Action),
		Services: make(map[string]Service),
		Metrics:  make(map[string]Stage),
		DBPath:   "/tmp/mistify-agent.db",
	}

	/*
		for name, _ := range valid_actions {
			c.addAction(name)
		}
	*/
	return c
}

func validateStage(stage *Stage, prefix string) error {
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

		if action.Sync == nil {
			action.Sync = make([]Stage, 0)
		}

		for _, s := range action.Sync {
			if err := validateStage(&s, name+" sync"); err != nil {
				return err
			}
		}
		if action.Async == nil {
			action.Async = make([]Stage, 0)
		}
		for _, s := range action.Async {
			if err := validateStage(&s, name+" async"); err != nil {
				return err
			}
		}

		c.Actions[name] = action
	}

	for name, metric := range newConfig.Metrics {
		if _, ok := c.Metrics[name]; ok {
			return fmt.Errorf("metric %s has already been defined", name)
		}
		if _, ok := valid_metrics[name]; !ok {
			return fmt.Errorf("metric %s is not a valid metric", name)
		}

		if err := validateStage(&metric, name+" metric"); err != nil {
			return err
		}

		c.Metrics[name] = metric
	}

	return nil
}

func (c *Config) Fixup() error {
	/*
		for _, service := range c.Services {
			 anything to do here?
		}
	*/

	for name, _ := range valid_actions {
		action, ok := c.Actions[name]
		if !ok {
			return fmt.Errorf("nothing defined for action %s", name)
		}
		if len(action.Sync) == 0 && len(action.Async) == 0 {
			return fmt.Errorf("no pipelines defined for action %s", name)
		}
		/*
			c.Actions[name] = Action{
				Sync:  make([]Stage, 0),
				Async: make([]Stage, 0),
			}
		*/
		//}

	}
	for name, action := range c.Actions {
		for j, _ := range action.Sync {
			stage := &action.Sync[j]
			if _, ok := c.Services[stage.Service]; !ok {
				return fmt.Errorf("%s unable to find service %s", name, stage.Service)
			}
			if stage.Args == nil {
				stage.Args = make(map[string]string)
			}
		}

		for j, _ := range action.Async {
			stage := &action.Async[j]
			if _, ok := c.Services[stage.Service]; !ok {
				return fmt.Errorf("%s unable to find service %s", name, stage.Service)
			}
			if stage.Args == nil {
				stage.Args = make(map[string]string)
			}
		}
	}

	// TODO: add builtins for create and delete

	for name, _ := range valid_metrics {
		_, ok := c.Metrics[name]
		if !ok {
			return fmt.Errorf("nothing defined for metric %s", name)
		}
	}

	for name, metric := range c.Metrics {
		if _, ok := c.Services[metric.Service]; !ok {
			return fmt.Errorf("%s unable to find service %s", name, metric.Service)
		}
		if metric.Args == nil {
			metric.Args = make(map[string]string)
		}
		c.Metrics[name] = metric
	}

	return nil

}
