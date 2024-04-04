package config

import (
	"gopkg.in/yaml.v2"
	"log"
	"os"
)

type Config struct {
	Github struct {
		Token string `yaml:"token"`
	} `yaml:"github"`
}

func Load() *Config {
	yamlFile, err := os.ReadFile("internal/config/config.yaml")
	if err != nil {
		log.Fatalf("Error reading YAML file: %v", err)
	}

	// Parse YAML content
	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("Error parsing YAML: %v", err)
	}

	log.Println(config)

	return &config
}
