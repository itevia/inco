package main

import (
	"fmt"
	"os"

	"github.com/najeal/gvy/internal"
)

const configPath = "inco.yaml"

func runTests() error {
	cfgBytes, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	config := internal.LoadConfig(cfgBytes)
	if !internal.ExecuteTests(config.TestPaths) {
		return fmt.Errorf("tests failed")
	}
	return nil
}

func runUploads() error {
	cfgBytes, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	config := internal.LoadConfig(cfgBytes)
	if err := internal.UploadScripts(config.IntegrationSuiteTenantURL, config.IntegrationSuiteAPIURL, config.UploadScripts); err != nil {
		return err
	}
	fmt.Println("Upload completed !")
	return nil
}
