# config

[![config](https://godoc.org/github.com/mistifyio/mistify-agent/config?status.png)](https://godoc.org/github.com/mistifyio/mistify-agent/config)

Package config handles agent configuration from a config file. It also defines
the valid actions along with their types.

## Usage

```go
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
```

#### type Action

```go
type Action struct {
	Type   ActionType
	Stages []Stage `json:"stages"`
}
```

Action is a set of stages and how they should be handled

#### type ActionType

```go
type ActionType int
```

ActionType describes synchronicity of an action

```go
const (
	// InfoAction is for synchronous information retrieval in JSON format
	InfoAction ActionType = iota
	// StreamAction is for synchronous data streaming
	StreamAction
	// AsyncAction is for asynchronous actions
	AsyncAction
)
```

#### type Config

```go
type Config struct {
	Actions  map[string]Action  `json:"actions"`
	Services map[string]Service `json:"services"`
	DBPath   string             `json:"dbpath"`
}
```

Config contains all of the configuration data

#### func  NewConfig

```go
func NewConfig() *Config
```
NewConfig creates a new Config

#### func (*Config) AddConfig

```go
func (c *Config) AddConfig(path string) error
```
AddConfig loads a configuration file

#### func (*Config) Fixup

```go
func (c *Config) Fixup() error
```
Fixup does a bit of validation and initializtion

#### type Service

```go
type Service struct {
	MaxPending uint   `json:"max_pending"`
	Port       uint   `json:"port"`
	Path       string `json:"path"`
}
```

Service is an HTTP service

#### type Stage

```go
type Stage struct {
	Service string            `json:"service"`
	Method  string            `json:"method"`
	Args    map[string]string `json:"args"`
}
```

Stage is a single step for an action

--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
