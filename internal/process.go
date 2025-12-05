package internal

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"
)

const (
	oauth2TokenURL    = "%s/oauth/token?grant_type=client_credentials"
	updateScriptURL   = "%s/api/v1/IntegrationDesigntimeArtifacts(Id='%s',Version='%s')/$links/Resources(Name='%s',ResourceType='%s')"
	fetchCSRFTokenURL = "%s/api/v1/"
	timeout           = time.Second * 10
	contentType       = "Content-Type"
	applicationJSON   = "application/json"
	xcsrfToken        = "x-csrf-token"
	xcsrfFetch        = "fetch"

	ENV_CPI_USER     = "CPI_USER"
	ENV_CPI_PASSWORD = "CPI_PASSWORD"
)

// ExecuteTests run groovy testing scripts.
func ExecuteTests(paths []string) bool {
	status := true
	for _, path := range paths {
		cmd := exec.Command("groovy", "-cp", "src", path)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		fmt.Println("Running: ", path)
		if err := cmd.Run(); err != nil {
			fmt.Println("Error: ", err)
			status = false
		}
	}
	return status
}

// UploadScripts authenticates over oauth2, then upload iflow scripts.
func UploadScripts(tenantURL, apiURL string, iflows []Iflow) error {
	client := &http.Client{Timeout: timeout}
	user := os.Getenv(ENV_CPI_USER)
	password := os.Getenv(ENV_CPI_PASSWORD)

	url := fmt.Sprintf(oauth2TokenURL, tenantURL)
	request, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	request.SetBasicAuth(user, password)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := client.Do(request)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected oauth2 token response - %d", res.StatusCode)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	m := map[string]any{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		return err
	}
	accessToken, ok := m["access_token"]
	if !ok {
		return fmt.Errorf("access token empty")
	}
	request, _ = http.NewRequest(http.MethodGet, fmt.Sprintf(fetchCSRFTokenURL, apiURL), nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	request.Header.Add(xcsrfToken, xcsrfFetch)

	res, err = client.Do(request)
	if err != nil {
		fmt.Println("cannot fetch csrf token")
		return err
	}
	token := res.Header.Get(xcsrfToken)
	var uploadErr error
	for _, iflow := range iflows {
		for _, script := range iflow.Scripts {
			url := fmt.Sprintf(updateScriptURL, apiURL, iflow.ID, iflow.Version, script.ID, script.Type)
			data, err := os.ReadFile(script.Path)
			if err != nil {
				return err
			}
			payload := fmt.Sprintf("{\"ResourceContent\": \"%s\"}", base64.StdEncoding.EncodeToString(data))

			request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader([]byte(payload)))
			if err != nil {
				return err
			}
			request.Header.Add(contentType, applicationJSON)
			request.Header.Add(xcsrfToken, token)
			request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))
			res, err := client.Do(request)
			if err != nil {
				return err
			}
			if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
				body, err := io.ReadAll(res.Body)
				if err != nil {
					fmt.Println("failed to parse Body")
				}
				fmt.Printf("FAILURE %d:\n %s\n", res.StatusCode, body)
				uploadErr = fmt.Errorf("failed to upload scripts")
				continue
			}
			fmt.Printf("SUCCESS uploading %s\n", script.ID)
		}
	}
	return uploadErr
}
