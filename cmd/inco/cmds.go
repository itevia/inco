package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/najeal/gvy/internal"
)

const (
	// Env variables to access CPI
	ENV_CPI_USER     = "CPI_CLIENT_ID"
	ENV_CPI_PASSWORD = "CPI_CLIENT_SECRET"

	timeout = time.Second * 10
)

const configPath = "inco.yaml"

func runTests() error {
	cfgBytes, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	config := internal.LoadConfig(cfgBytes)
	if !internal.ExecuteTests(config.TestPaths, os.Stdout, os.Stderr) {
		return fmt.Errorf("tests failed")
	}
	return nil
}

func runUploads() error {
	cfgBytes, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	clientID := os.Getenv(ENV_CPI_USER)
	clientSecret := os.Getenv(ENV_CPI_PASSWORD)
	config := internal.LoadConfig(cfgBytes)
	btpclient := internal.NewBTPClient(&http.Client{Timeout: timeout}, config.IntegrationSuiteTokenURL, config.IntegrationSuiteAPIURL, clientID, clientSecret)
	if err := internal.UploadScripts(btpclient, os.ReadFile, config.UploadScripts); err != nil {
		return err
	}
	fmt.Println("Upload completed !")
	return nil
}
