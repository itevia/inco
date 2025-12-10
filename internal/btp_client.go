package internal

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const (
	tokenURLGrantType = "%s?grant_type=client_credentials"
	updateScriptURL   = "%s/api/v1/IntegrationDesigntimeArtifacts(Id='%s',Version='%s')/$links/Resources(Name='%s',ResourceType='%s')"
	fetchCSRFTokenURL = "%s/api/v1/"
	contentType       = "Content-Type"
	applicationJSON   = "application/json"
	xcsrfToken        = "x-csrf-token"
	xcsrfFetch        = "fetch"
)

var (
	ErrUnexpectedStatusCode = errors.New("unexpected status code")
	ErrEmptyAccessToken     = errors.New("empty access token")
	ErrNoAccessToken        = errors.New("no access token")
	ErrNoCSRFToken          = errors.New("no csrf token")
)

func NewBTPClient(httpClient httpClient, tokenURL, apiURL, clientID, clientSecret string) *BTPClient {
	return &BTPClient{
		tokenURL:     tokenURL,
		apiURL:       apiURL,
		clientID:     clientID,
		clientSecret: clientSecret,
		hc:           httpClient,
	}
}

type IBTPClient interface {
	RequestToken() error
	FetchCSRFToken() error
	UpdateIflowResource(data []byte, iflow Iflow, script Script) error
}

type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

type BTPClient struct {
	tokenURL     string
	apiURL       string
	clientID     string
	clientSecret string
	accessToken  string
	csrfToken    string

	hc httpClient
}

func (c *BTPClient) RequestToken() error {
	req, err := buildOauth2AuthRequest(c.tokenURL, c.clientID, c.clientSecret)
	if err != nil {
		return err
	}
	res, err := c.hc.Do(req)
	if err != nil {
		return err
	}
	accessToken, err := getAccessTokenFromResponse(res)
	if err != nil {
		return err
	}
	c.accessToken = accessToken
	return nil
}

func (c *BTPClient) FetchCSRFToken() error {
	if c.accessToken == "" {
		return fmt.Errorf("%w: request access token first", ErrNoAccessToken)
	}
	req, err := buildFetchCSRFRequest(c.apiURL, c.accessToken)
	if err != nil {
		return err
	}
	res, err := c.hc.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("%w - %d", ErrUnexpectedStatusCode, res.StatusCode)
	}
	token := res.Header.Get(xcsrfToken)
	if token == "" {
		return fmt.Errorf("%w: not found in xcsrf response header", ErrNoCSRFToken)
	}
	c.csrfToken = token
	return nil
}

func (c *BTPClient) UpdateIflowResource(data []byte, iflow Iflow, script Script) error {
	if c.accessToken == "" {
		return fmt.Errorf("%w: request access token first", ErrNoAccessToken)
	}
	if c.csrfToken == "" {
		return fmt.Errorf("%w: request csrf token first", ErrNoCSRFToken)
	}
	payload := fmt.Sprintf("{\"ResourceContent\": \"%s\"}", base64.StdEncoding.EncodeToString(data))
	request, err := buildUpdateResourceRequest(c.apiURL, iflow, script, []byte(payload), c.accessToken, c.csrfToken)
	if err != nil {
		return err
	}
	res, err := c.hc.Do(request)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("%w - %s", ErrUnexpectedStatusCode, body)
	}
	return nil
}

// buildOauth2AuthRequest creates http request with BasicAuth authentication.
func buildOauth2AuthRequest(tokenURL, clientID, clientSecret string) (*http.Request, error) {
	url := fmt.Sprintf(tokenURLGrantType, tokenURL)
	request, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}
	request.SetBasicAuth(clientID, clientSecret)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return request, nil
}

// getAccessTokenFromResponse checks response Status Code and read response Body.
func getAccessTokenFromResponse(res *http.Response) (string, error) {
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w - %d", ErrUnexpectedStatusCode, res.StatusCode)
	}
	body, _ := io.ReadAll(res.Body)
	m := map[string]any{}
	err := json.Unmarshal(body, &m)
	if err != nil {
		return "", err
	}
	accessToken, ok := m["access_token"].(string)
	if !ok {
		return "", ErrEmptyAccessToken
	}
	return accessToken, nil
}

func buildFetchCSRFRequest(apiURL, accessToken string) (*http.Request, error) {
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf(fetchCSRFTokenURL, apiURL), nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	request.Header.Add(xcsrfToken, xcsrfFetch)
	return request, nil
}

func buildUpdateResourceRequest(apiURL string, iflow Iflow, script Script, payload []byte, accessToken, token string) (*http.Request, error) {
	url := fmt.Sprintf(updateScriptURL, apiURL, iflow.ID, iflow.Version, script.ID, script.Type)
	request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader([]byte(payload)))
	if err != nil {
		return nil, err
	}
	request.Header.Add(contentType, applicationJSON)
	request.Header.Add(xcsrfToken, token)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	return request, nil
}
