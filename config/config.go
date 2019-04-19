package config

import (
  "log"
  "sync"

  "github.com/kelseyhightower/envconfig"
)

type Config struct {
  Port                     string `envconfig:"port" default:"4444"`
  BrowserConfigFile        string `envconfig:"browser_config_file" default:"config/browsers.yaml"`
  StartupTimeout           int32  `envconfig:"startup_timeout" default:"900000"`
  NewSessionAttemptTimeout int32  `envconfig:"new_session_attempt_timeout" default:"60000"`
  SeleniumStartupTimeout   int32  `envconfig:"selenium_startup_timeout" default:"60000"`
  RetryCount               int32  `envconfig:"retry_count" default:"30"`
  SigningKey               string `envconfig:"signing_key" default:"secret_key"`
  GridLabel                string `envconfig:"grid_label" default:"dev"`
  NodeSelectorKey          string `envconfig:"node_selector_key"`
  NodeSelectorValue        string `envconfig:"node_selector_value"`
  CPURequest               string `envconfig:"cpu_request" default:"400m"`
  MemoryRequest            string `envconfig:"memory_request" default:"600Mi"`
  CPULimit                 string `envconfig:"cpu_limit" default:"600m"`
  MemoryLimit              string `envconfig:"memory_limit" default:"1000Mi"`
  GridTimeout              int    `envconfig:"sersan_grid_timeout" default:"300"`
}

var conf Config
var once sync.Once

// GetConfig returns the singleton config instance
func Get() Config {
  once.Do(func() {
    err := envconfig.Process("", &conf)
    if err != nil {
      log.Fatal("Can't load config: ", err)
    }
  })

  return conf
}
