package client

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

// MockClient pretends to be a vault client, however, it only stores data to a cache
type MockClient struct {
	cache map[string]interface{}
}

// CreateMockClient creates a mock vault client
func CreateMockClient() (*MockClient, error) {
	return &MockClient{make(map[string]interface{})}, nil
}

type vaultResponse struct {
	Auth          interface{} `json:"auth"`
	Data          interface{} `json:"data"`
	LeaseDuration int         `json:"lease_duration"`
	LeaseID       string      `json:"lease_id"`
	Renewable     bool        `json:"renewable"`
}

func (vaultClient *MockClient) get(path string) ([]byte, error) {
	data, ok := vaultClient.cache[path]
	if !ok {
		return nil, newHTTPStatusError(404)
	}
	return json.Marshal(vaultResponse{nil, data, 2764800, "", false})
}

func (vaultClient *MockClient) list(path string) ([]byte, error) {
	prefix := path + "/"
	items := []string{}
	for key := range vaultClient.cache {
		if strings.HasPrefix(key, prefix) {
			subPath := strings.TrimPrefix(key, prefix)
			segments := strings.Split(subPath, "/")
			if len(segments) > 1 {
				items = append(items, segments[0]+"/")
			} else {
				items = append(items, segments[0])
			}
		}
	}
	if len(items) == 0 {
		return nil, newHTTPStatusError(404)
	}
	return json.Marshal(vaultResponse{nil, map[string][]string{"keys": items}, 2764800, "", false})
}

func (vaultClient *MockClient) remove(path string) error {
	_, ok := vaultClient.cache[path]
	if !ok {
		return newHTTPStatusError(404)
	}
	delete(vaultClient.cache, path)
	return nil
}

func (vaultClient *MockClient) store(path string, body io.Reader) error {

	data, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	var ud interface{}
	err = json.Unmarshal(data, &ud)
	if err != nil {
		return newHTTPStatusError(400)
	}
	vaultClient.cache[path] = ud
	return nil
}

func (vaultClient *MockClient) vaultHTTP(method, path string, body io.Reader) ([]byte, error) {

	path = strings.TrimSuffix(path, "/")

	var responseBytes []byte
	var err error

	switch method {
	case "GET":
		responseBytes, err = vaultClient.get(path)
	case "LIST":
		responseBytes, err = vaultClient.list(path)
	case "DELETE":
		err = vaultClient.remove(path)
	case "POST":
		err = vaultClient.store(path, body)
	case "PUT":
		err = vaultClient.store(path, body)
	}
	if err != nil {
		return nil, err
	}
	return responseBytes, nil
}

// VaultGet handles a GET request and returns a response
func (vaultClient *MockClient) VaultGet(path string) ([]byte, error) {
	return vaultClient.vaultHTTP("GET", path, nil)
}

// VaultList handles a LIST request and returns a response
func (vaultClient *MockClient) VaultList(path string) ([]byte, error) {
	return vaultClient.vaultHTTP("LIST", path, nil)
}

// VaultDelete handles a DELETE request and returns a response
func (vaultClient *MockClient) VaultDelete(path string) ([]byte, error) {
	return vaultClient.vaultHTTP("DELETE", path, nil)
}

// VaultPost handles a POST request and returns a response
func (vaultClient *MockClient) VaultPost(path string, body io.Reader) ([]byte, error) {
	return vaultClient.vaultHTTP("POST", path, body)
}

// VaultPut handles a PUT request and returns a response
func (vaultClient *MockClient) VaultPut(path string, body io.Reader) ([]byte, error) {
	return vaultClient.vaultHTTP("PUT", path, body)
}

type httpStatusError struct {
	statusCode int
}

func (e httpStatusError) Error() string {
	return fmt.Sprintf("request resulted in status: %d", e.statusCode)
}

func (e httpStatusError) HTTPStatusCode() int {
	return e.statusCode
}

func newHTTPStatusError(statusCode int) httpStatusError {
	return httpStatusError{statusCode}
}
