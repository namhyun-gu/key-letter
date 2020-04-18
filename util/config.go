package util

import (
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Port  int
	CertFile string
	CertKeyFile string
	Timeout uint
	Redis struct {
		Addr string
		Password string
	}
	Opts struct {
		Issuer string
		// Number of seconds a Code hash is valid for. Defaults to 30 seconds.s
		Period uint
		// Code Digits length. Available 6 or 8. Defaults to 6.
		Digits int
		// Generate code algorithm option. Available SHA1, SHA256, SHA512, MD5 algorithm. Defaults to SHA1.
		Algorithm string
	}
}

func GetConfig() *Config {
	var config Config
	filename, err := filepath.Abs("config.yaml")
	if err != nil {
		return nil
	}
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return nil
	}

	if config.Timeout == 0 {
		config.Timeout = 30
	}
	if config.Port == 0 {
		config.Port = 8000
	}
	return &config
}