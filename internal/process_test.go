package internal

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUploadScripts(t *testing.T) {
	t.Run("FailingRequestToken", func(t *testing.T) {
		mockedClient := BTPClientMock{
			requestTokenError: ErrUnexpectedStatusCode,
		}
		require.ErrorIs(t, UploadScripts(&mockedClient, nil, nil), ErrUnexpectedStatusCode)
	})
	t.Run("FailingFetchCSRFToken", func(t *testing.T) {
		mockedClient := BTPClientMock{
			fetchCSRFTokenError: ErrUnexpectedStatusCode,
		}
		require.ErrorIs(t, UploadScripts(&mockedClient, nil, nil), ErrUnexpectedStatusCode)
	})
	t.Run("OneReadFileFailed", func(t *testing.T) {
		header := http.Header{}
		header.Add(xcsrfToken, xcsrfFetch)
		mockedHTTPClient := newMockedHTTPClient([]mockedResponse{
			{res: &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader([]byte(`{"access_token":"myaccesstoken"}`)))}},
			{res: &http.Response{StatusCode: http.StatusOK, Header: header}},
		})
		client := NewBTPClient(mockedHTTPClient, ttokenURL, tapiURL, tclientID, tclientSecret)
		err := UploadScripts(client, func(string) ([]byte, error) {
			return nil, fmt.Errorf("read failed")
		}, []Iflow{{ID: "iflow1", Version: "iflowv1", Scripts: []Script{{ID: "script1", Type: "groovy", Path: "path1"}}}})
		require.ErrorContains(t, err, "some reading/uploading")
	})
	t.Run("AllReadFileFailed", func(t *testing.T) {
		header := http.Header{}
		header.Add(xcsrfToken, xcsrfFetch)
		mockedHTTPClient := newMockedHTTPClient([]mockedResponse{
			{res: &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader([]byte(`{"access_token":"myaccesstoken"}`)))}},
			{res: &http.Response{StatusCode: http.StatusOK, Header: header}},
		})
		client := NewBTPClient(mockedHTTPClient, ttokenURL, tapiURL, tclientID, tclientSecret)
		err := UploadScripts(client, func(string) ([]byte, error) {
			return nil, fmt.Errorf("read failed")
		}, []Iflow{{ID: "iflow1", Version: "iflowv1", Scripts: []Script{{ID: "script1", Type: "groovy", Path: "path1"}, {ID: "script2", Type: "js", Path: "path2"}}}})
		require.ErrorContains(t, err, "some reading/uploading")
	})
	t.Run("PartReadFileFailed", func(t *testing.T) {
		mockedClient := BTPClientMock{
			updateIflowResourceErrors: []error{nil},
		}
		err := UploadScripts(&mockedClient, func(path string) ([]byte, error) {
			if path == "path1" {
				return nil, fmt.Errorf("read failed")
			}
			return []byte(`data`), nil
		}, []Iflow{{ID: "iflow1", Version: "iflowv1", Scripts: []Script{{ID: "script1", Type: "groovy", Path: "path1"}, {ID: "script2", Type: "js", Path: "path2"}}}})
		require.ErrorContains(t, err, "some reading/uploading")
	})
	t.Run("PartUpdateFailed", func(t *testing.T) {
		mockedClient := BTPClientMock{
			updateIflowResourceErrors: []error{ErrUnexpectedStatusCode, nil},
		}
		err := UploadScripts(&mockedClient, func(path string) ([]byte, error) {
			if path == "path1" {
				return nil, fmt.Errorf("read failed")
			}
			return []byte(`data`), nil
		}, []Iflow{{ID: "iflow1", Version: "iflowv1", Scripts: []Script{{ID: "script1", Type: "groovy", Path: "path2"}, {ID: "script2", Type: "js", Path: "path3"}}}})
		require.ErrorContains(t, err, "some reading/uploading")
	})
	t.Run("AllSucceed", func(t *testing.T) {
		mockedClient := BTPClientMock{
			updateIflowResourceErrors: []error{nil, nil},
		}
		err := UploadScripts(&mockedClient, func(path string) ([]byte, error) {
			if path == "path1" {
				return nil, fmt.Errorf("read failed")
			}
			return []byte(`data`), nil
		}, []Iflow{{ID: "iflow1", Version: "iflowv1", Scripts: []Script{{ID: "script1", Type: "groovy", Path: "path2"}, {ID: "script2", Type: "js", Path: "path3"}}}})
		require.Nil(t, err)
	})
}

type BTPClientMock struct {
	requestTokenError         error
	fetchCSRFTokenError       error
	updateIflowResourceErrors []error
}

func (c *BTPClientMock) RequestToken() error {
	return c.requestTokenError
}
func (c *BTPClientMock) FetchCSRFToken() error {
	return c.fetchCSRFTokenError
}
func (c *BTPClientMock) UpdateIflowResource(data []byte, iflow Iflow, script Script) error {
	if len(c.updateIflowResourceErrors) == 0 {
		panic("")
	}
	err := c.updateIflowResourceErrors[0]
	c.updateIflowResourceErrors = c.updateIflowResourceErrors[1:]
	return err
}
