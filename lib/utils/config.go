// Copyright (c) 2019 Daniel Oaks <daniel@danieloaks.net>
// released under the MIT license

package utils

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

// AccountConfig is the configuration for a single IRC user account.
type AccountConfig struct {
	Username string
	Password string
}

// ServerConfig holds the configuration for a specific IRC server.
type ServerConfig struct {
	Address           string
	TLS               bool
	Password          *string
	ResetBetweenTests bool `yaml:"reset-between-tests"`
}

// Config holds our configuration information.
type Config struct {
	Server   ServerConfig
	Accounts []AccountConfig
}

// LoadConfig loads the given YAML configuration file.
func LoadConfig(filename string) (config *Config, err error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	// test config here

	return config, nil
}
