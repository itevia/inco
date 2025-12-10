package internal

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	ttokenURL     = "https://itevia.com/oauth/token"
	tapiURL       = "https://api.itevia.com"
	tclientID     = "clientid"
	tclientSecret = "clientsecret"
	tapiBadURL    = "://bad-url"
)

func TestBuildOauth2AuthRequest(t *testing.T) {
	_, err := buildOauth2AuthRequest("://bad-url", "clientid", "clientsecret")
	require.Error(t, err)
	request, err := buildOauth2AuthRequest("https://itevia.com/oauth/token", "clientid", "clientsecret")
	require.Nil(t, err)
	require.Equal(t, "/oauth/token?grant_type=client_credentials", request.URL.RequestURI())
	require.Equal(t, fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte("clientid:clientsecret"))), request.Header.Get("Authorization"))
	require.Equal(t, "application/x-www-form-urlencoded", request.Header.Get("Content-Type"))
}

func TestBuildFetchCSRFRequest(t *testing.T) {
	_, err := buildFetchCSRFRequest("://bad-url", "myaccesstoken")
	require.Error(t, err)
	request, err := buildFetchCSRFRequest("https://itevia.com", "myaccesstoken")
	require.Nil(t, err)
	require.Equal(t, "/api/v1/", request.URL.RequestURI())
	require.Equal(t, "Bearer myaccesstoken", request.Header.Get("Authorization"))
	require.Equal(t, xcsrfFetch, request.Header.Get(xcsrfToken))
}

func TestBuildUpdateResourceRequest(t *testing.T) {
	_, err := buildUpdateResourceRequest("://bad-url", Iflow{ID: "iflowid", Version: "iflowversion"}, Script{ID: "scriptid", Type: "scripttype", Path: "scriptpath"}, []byte("{\"a\":\"b\"}"), "myaccesstoken", "myxscrftoken")
	require.Error(t, err)
	request, err := buildUpdateResourceRequest("https://api.itevia.com", Iflow{ID: "iflowid", Version: "iflowversion"}, Script{ID: "scriptid", Type: "scripttype", Path: "scriptpath"}, []byte("{\"a\":\"b\"}"), "myaccesstoken", "myxscrftoken")
	require.Nil(t, err)
	require.Equal(t, "/api/v1/IntegrationDesigntimeArtifacts(Id='iflowid',Version='iflowversion')/$links/Resources(Name='scriptid',ResourceType='scripttype')", request.URL.RequestURI())
	require.Equal(t, "Bearer myaccesstoken", request.Header.Get("Authorization"))
	require.Equal(t, "myxscrftoken", request.Header.Get(xcsrfToken))
	require.Equal(t, "application/json", request.Header.Get("Content-Type"))
}

func TestGetAccessTokenFromResponse(t *testing.T) {
	_, err := getAccessTokenFromResponse(&http.Response{
		StatusCode: http.StatusBadRequest,
	})
	require.ErrorIs(t, err, ErrUnexpectedStatusCode)

	_, err = getAccessTokenFromResponse(&http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader([]byte(`{"a":"b"}`))),
	})
	require.ErrorIs(t, err, ErrEmptyAccessToken)

	_, err = getAccessTokenFromResponse(&http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader([]byte(`{access_token": "myaccesstoken"}`))),
	})
	require.Error(t, err)

	accessToken, err := getAccessTokenFromResponse(&http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader([]byte(`{"access_token": "myaccesstoken"}`))),
	})
	require.NoError(t, err)
	require.Equal(t, "myaccesstoken", accessToken)
}

func TestBTPClientRequestToken(t *testing.T) {
	t.Run("FailBuildRequest", func(t *testing.T) {
		mockedHTTPClient := newMockedHTTPClient([]mockedResponse{})
		client := BTPClient{
			tokenURL:     "://bad-url",
			apiURL:       "https://api.itevia.com",
			clientID:     "clientid",
			clientSecret: "clientsecret",
			hc:           mockedHTTPClient,
		}
		err := client.RequestToken()
		require.Error(t, err)
	})

	t.Run("FailRequestSend", func(t *testing.T) {
		mockedHTTPClient := newMockedHTTPClient([]mockedResponse{{err: fmt.Errorf("http do failed")}})
		client := BTPClient{
			tokenURL:     "https://itevia.com/oauth/token",
			apiURL:       "https://api.itevia.com",
			clientID:     "clientid",
			clientSecret: "clientsecret",
			hc:           mockedHTTPClient,
		}

		err := client.RequestToken()
		require.ErrorContains(t, err, "http do failed")
	})

	t.Run("InvalidResponse", func(t *testing.T) {
		mockedHTTPClient := newMockedHTTPClient([]mockedResponse{{res: &http.Response{StatusCode: http.StatusBadGateway}}})
		client := BTPClient{
			tokenURL:     "https://itevia.com/oauth/token",
			apiURL:       "https://api.itevia.com",
			clientID:     "clientid",
			clientSecret: "clientsecret",
			hc:           mockedHTTPClient,
		}

		err := client.RequestToken()
		require.ErrorIs(t, err, ErrUnexpectedStatusCode)
	})

	t.Run("Valid", func(t *testing.T) {
		mockedHTTPClient := newMockedHTTPClient([]mockedResponse{{res: &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader([]byte(`{"access_token":"myaccesstoken"}`)))}}})
		client := BTPClient{
			tokenURL:     "https://itevia.com/oauth/token",
			apiURL:       "https://api.itevia.com",
			clientID:     "clientid",
			clientSecret: "clientsecret",
			hc:           mockedHTTPClient,
		}

		require.Nil(t, client.RequestToken())
		require.Equal(t, "myaccesstoken", client.accessToken)
	})
}

func TestBTPClientUpdateResource(t *testing.T) {
	t.Run("NoAccessToken", func(t *testing.T) {
		client := NewBTPClient(nil, ttokenURL, tapiURL, tclientID, tclientSecret)
		require.ErrorIs(t, client.UpdateIflowResource([]byte(`data`), Iflow{}, Script{}), ErrNoAccessToken)
	})

	t.Run("NoCSRFToken", func(t *testing.T) {
		client := NewBTPClient(nil, ttokenURL, tapiURL, tclientID, tclientSecret)
		client.accessToken = "myaccesstoken"
		require.ErrorIs(t, client.UpdateIflowResource([]byte(`data`), Iflow{}, Script{}), ErrNoCSRFToken)
	})

	t.Run("FailBuildRequest", func(t *testing.T) {
		mockedHTTPClient := newMockedHTTPClient([]mockedResponse{})
		client := NewBTPClient(mockedHTTPClient, ttokenURL, tapiBadURL, tclientID, tclientSecret)
		client.accessToken = "myaccesstoken"
		client.csrfToken = "mycsrftoken"
		require.Error(t, client.UpdateIflowResource([]byte(`data`), Iflow{}, Script{}))
	})

	t.Run("FailHttpDo", func(t *testing.T) {
		mockedHTTPClient := newMockedHTTPClient([]mockedResponse{{err: fmt.Errorf("http do send")}})
		client := NewBTPClient(mockedHTTPClient, ttokenURL, tapiURL, tclientID, tclientSecret)
		client.accessToken = "myaccesstoken"
		client.csrfToken = "mycsrftoken"
		require.ErrorContains(t, client.UpdateIflowResource([]byte(`data`), Iflow{ID: "iid", Version: "iversion"}, Script{ID: "sid", Type: "stype"}), "http do send")
	})

	t.Run("FailInvalidResponseStatus", func(t *testing.T) {
		mockedHTTPClient := newMockedHTTPClient([]mockedResponse{{res: &http.Response{StatusCode: http.StatusBadGateway, Body: io.NopCloser(bytes.NewReader([]byte(`{"message":"error cause"}`)))}}})
		client := NewBTPClient(mockedHTTPClient, ttokenURL, tapiURL, tclientID, tclientSecret)
		client.accessToken = "myaccesstoken"
		client.csrfToken = "mycsrftoken"
		require.ErrorIs(t, client.UpdateIflowResource([]byte(`data`), Iflow{ID: "iid", Version: "iversion"}, Script{ID: "sid", Type: "stype"}), ErrUnexpectedStatusCode)
	})
	t.Run("Valid", func(t *testing.T) {
		mockedHTTPClient := newMockedHTTPClient([]mockedResponse{{res: &http.Response{StatusCode: http.StatusCreated}}})
		client := NewBTPClient(mockedHTTPClient, ttokenURL, tapiURL, tclientID, tclientSecret)
		client.accessToken = "myaccesstoken"
		client.csrfToken = "mycsrftoken"
		require.Nil(t, client.UpdateIflowResource([]byte(`data`), Iflow{ID: "iid", Version: "iversion"}, Script{ID: "sid", Type: "stype"}))
	})
}

func TestBTPClientFetchCSRFToken(t *testing.T) {
	t.Run("NoAccessToken", func(t *testing.T) {
		client := BTPClient{
			tokenURL:     "https://itevia.com/oauth/token",
			apiURL:       "https://api.itevia.com",
			clientID:     "clientid",
			clientSecret: "clientsecret",
		}
		require.ErrorIs(t, client.FetchCSRFToken(), ErrNoAccessToken)
	})

	t.Run("FailBuildRequest", func(t *testing.T) {
		mockedHTTPClient := newMockedHTTPClient([]mockedResponse{})
		client := BTPClient{
			tokenURL:     "https://itevia.com/oauth/token",
			apiURL:       "https://api.itevia.com",
			clientID:     "clientid",
			clientSecret: "clientsecret",
			accessToken:  "",
			hc:           mockedHTTPClient,
		}
		require.Error(t, client.FetchCSRFToken())
	})

	t.Run("FailRequestBuild", func(t *testing.T) {
		mockedHTTPClient := newMockedHTTPClient([]mockedResponse{})
		client := BTPClient{
			tokenURL:     "https://itevia.com/oauth/token",
			apiURL:       "://bad-url",
			clientID:     "clientid",
			clientSecret: "clientsecret",
			accessToken:  "myaccesstoken",
			hc:           mockedHTTPClient,
		}
		require.Error(t, client.FetchCSRFToken())
	})

	t.Run("FailRequestSend", func(t *testing.T) {
		mockedHTTPClient := newMockedHTTPClient([]mockedResponse{{err: fmt.Errorf("http do send")}})
		client := BTPClient{
			tokenURL:     "https://itevia.com/oauth/token",
			apiURL:       "https://api.itevia.com",
			clientID:     "clientid",
			clientSecret: "clientsecret",
			accessToken:  "myaccesstoken",
			hc:           mockedHTTPClient,
		}
		require.ErrorContains(t, client.FetchCSRFToken(), "http do send")
	})

	t.Run("InvalidResponse", func(t *testing.T) {
		mockedHTTPClient := newMockedHTTPClient([]mockedResponse{{res: &http.Response{StatusCode: http.StatusBadGateway}}})
		client := BTPClient{
			tokenURL:     "https://itevia.com/oauth/token",
			apiURL:       "https://api.itevia.com",
			clientID:     "clientid",
			clientSecret: "clientsecret",
			accessToken:  "myaccesstoken",
			hc:           mockedHTTPClient,
		}
		require.ErrorIs(t, client.FetchCSRFToken(), ErrUnexpectedStatusCode)
	})

	t.Run("CSRFNotSet", func(t *testing.T) {
		mockedHTTPClient := newMockedHTTPClient([]mockedResponse{{res: &http.Response{StatusCode: http.StatusOK}}})
		client := BTPClient{
			tokenURL:     "https://itevia.com/oauth/token",
			apiURL:       "https://api.itevia.com",
			clientID:     "clientid",
			clientSecret: "clientsecret",
			accessToken:  "myaccesstoken",
			hc:           mockedHTTPClient,
		}
		require.ErrorIs(t, client.FetchCSRFToken(), ErrNoCSRFToken)
	})

	t.Run("Valid", func(t *testing.T) {
		header := http.Header{}
		header.Add(xcsrfToken, "mycsrftoken")
		mockedHTTPClient := newMockedHTTPClient([]mockedResponse{{res: &http.Response{StatusCode: http.StatusOK, Header: header}}})
		client := NewBTPClient(mockedHTTPClient, "https://itevia.com/oauth/token", "https://api.itevia.com", "clientid", "clientsecret")
		client.accessToken = "myaccesstoken"
		client.hc = mockedHTTPClient
		require.Nil(t, client.FetchCSRFToken())
		require.Equal(t, "mycsrftoken", client.csrfToken)
	})
}

type mockedResponse struct {
	err error
	res *http.Response
}

type httpClientMock struct {
	responses []mockedResponse
}

func (c *httpClientMock) Do(req *http.Request) (*http.Response, error) {
	if len(c.responses) == 0 {
		panic("unexpected call to mocked http client")
	}
	res := c.responses[0]
	c.responses = c.responses[1:len(c.responses)]
	if res.err != nil {
		return nil, res.err
	}
	return res.res, nil
}

func newMockedHTTPClient(responses []mockedResponse) *httpClientMock {
	return &httpClientMock{
		responses: responses,
	}
}
