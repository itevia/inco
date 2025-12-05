package internal

import (
	"github.com/goccy/go-yaml"
)

type Config struct {
	IntegrationSuiteTenantURL string   `yaml:"tenantURL"`
	IntegrationSuiteAPIURL    string   `yaml:"apiURL"`
	TestPaths                 []string `yaml:"testPaths"`
	UploadScripts             []Iflow  `yaml:"uploadScripts"`
}

type Iflow struct {
	ID      string   `yaml:"id"`
	Version string   `yaml:"version"`
	Scripts []Script `yaml:"scripts"`
}

type Script struct {
	ID   string `yaml:"id"`
	Type string `yaml:"type"`
	Path string `yaml:"path"`
}

func LoadConfig(data []byte) Config {
	var cfg Config
	yaml.Unmarshal(data, &cfg)
	return cfg
}
