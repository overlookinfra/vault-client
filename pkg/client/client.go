package client

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	vault_api "github.com/hashicorp/vault/api"
)

// Cert - certificate
type Cert struct {
	CA         string
	Cert       string
	PrivateKey string
}

// HTTPStatusError - HTTP status error
type HTTPStatusError struct {
	Path       string
	StatusCode int
}

// VaultClient - Vault client
type VaultClient struct {
	client     *vault_api.Client
	token      string
	apiVersion string
}

const apiVersion = "v1"

// CreateClient - creates the default Vault client
func CreateClient(TLSCaFile string, vaultToken string) (*VaultClient, error) {
	vaultCACert, err := ioutil.ReadFile(TLSCaFile)
	if err != nil {
		return nil, err
	}

	return CreateClientUsingParams(vaultCACert, vaultToken, apiVersion)
}

// CreateClientUsingParams - creates a Vault client
func CreateClientUsingParams(caCert []byte, token string, apiVersion string) (*VaultClient, error) {
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:            caCertPool,
				InsecureSkipVerify: true, // TODO - the cert validation is failing with the new certs.  Need to address why.
			},
		},
	}
	clientConfig := &vault_api.Config{
		Address:    "https://vault:8200",
		HttpClient: httpClient,
	}
	log.Print("Building vault client from vault_api")
	vaultClient, err := vault_api.NewClient(clientConfig)
	if err != nil {
		return nil, err
	}

	return &VaultClient{vaultClient, token, apiVersion}, nil
}

func unmarshalPostBody(postBody io.Reader) (map[string]interface{}, error) {
	if postBody != nil {
		buf := new(bytes.Buffer)
		buf.ReadFrom(postBody)

		var reqBody map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &reqBody); err != nil {
			return nil, err
		}
		return reqBody, nil
	}
	return nil, nil
}

func vaultHTTP(method, path string, body io.Reader, vaultClient *vault_api.Client, token string) ([]byte, error) {
	request := vaultClient.NewRequest(method, fmt.Sprintf("/%s/%s", apiVersion, path))
	postBody, err := unmarshalPostBody(body)
	if err != nil {
		return nil, err
	}
	if postBody != nil {
		request.SetJSONBody(postBody)
	}

	resp, err := vaultClient.RawRequest(request)
	if resp == nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, NewHTTPStatusError(path, resp.StatusCode)
	}
	responseBytes, err := ioutil.ReadAll(resp.Body)
	return responseBytes, nil
}

func (vaultClient *VaultClient) vaultHTTP(method, path string, body io.Reader) ([]byte, error) {
	return vaultHTTP(method, path, body, vaultClient.client, vaultClient.token)
}

// VaultGet - sends GET request to Vault
func (vaultClient *VaultClient) VaultGet(path string) ([]byte, error) {
	return vaultClient.vaultHTTP("GET", path, nil)
}

// VaultList - sends LIST request to Vault
func (vaultClient *VaultClient) VaultList(path string) ([]byte, error) {
	return vaultClient.vaultHTTP("LIST", path, nil)
}

// VaultDelete - sends DELETE request to Vault
func (vaultClient *VaultClient) VaultDelete(path string) ([]byte, error) {
	return vaultClient.vaultHTTP("DELETE", path, nil)
}

// VaultPost - sends POST request to Vault
func (vaultClient *VaultClient) VaultPost(path string, body io.Reader) ([]byte, error) {
	return vaultClient.vaultHTTP("POST", path, body)
}

// VaultPut - sends PUT request to Vault
func (vaultClient *VaultClient) VaultPut(path string, body io.Reader) ([]byte, error) {
	return vaultClient.vaultHTTP("PUT", path, body)
}

// Error - gets error message from HTTPStatusError
func (e HTTPStatusError) Error() string {
	return fmt.Sprintf("request to %s resulted in status: %d", e.Path, e.StatusCode)
}

// HTTPStatusCode - gets status code from HTTPStatusError
func (e HTTPStatusError) HTTPStatusCode() int {
	return e.StatusCode
}

// NewHTTPStatusError - creates HTTPStatusError using path and statusCode
func NewHTTPStatusError(path string, statusCode int) HTTPStatusError {
	return HTTPStatusError{
		Path:       path,
		StatusCode: statusCode,
	}
}

// ErrMissingCertAndKey - a specific error
var ErrMissingCertAndKey = errors.New("unable to request a cert from Vault")

// PkiRoot - PKI root
const PkiRoot = "pki" // this should probably reference the PkiRoot decalred in pki.go

// GetCert - get certificate
func GetCert() (*Cert, error) {
	// TODO: function can be removed after switch to swarm installer
	vaultClient, err := CreateClient("/etc/ssl/certs/puppet-discovery/shared.ca", os.Getenv("VAULT_TOKEN"))
	if err != nil {
		return nil, err
	}

	subdomain := os.Getenv("SERVICE_NAME")
	tenantID := os.Getenv("TENANT_ID")

	// TODO: replace the body block below with this one. Holding off until pipelines stablizes
	// body := vault.V1pkiIssueBody{
	// 	CommonName:        fmt.Sprintf("%s.%s.puppetdiscovery.com", subdomain, tenantId),
	// 	TTL:               "87500h",
	// 	ExcludeCNFromSans: true,
	// 	AltNames:          fmt.Sprintf("%s,localhost", subdomain),
	// }

	body := map[string]interface{}{
		"common_name": fmt.Sprintf("%s.%s.puppetdiscovery.com", subdomain, tenantID),
		"ttl":         "8765h", // expire after a year
		// "ip_sans":              "0.0.0.0,127.0.0.1,192.168.33.11",
		"exclude_cn_from_sans": true,
		"alt_names":            fmt.Sprintf("%s,localhost", subdomain),
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	certPath := fmt.Sprintf("%s/issue/vault-%s", PkiRoot, tenantID)
	reqBytes, err := vaultClient.VaultPut(certPath, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	var certRes map[string]interface{}
	if err := json.Unmarshal(reqBytes, &certRes); err != nil {
		return nil, err
	}

	if certRes["data"] == nil {
		return nil, ErrMissingCertAndKey
	}

	certResData := certRes["data"].(map[string]interface{})

	return &Cert{
		CA:         certResData["issuing_ca"].(string),
		Cert:       certResData["certificate"].(string),
		PrivateKey: certResData["private_key"].(string),
	}, err
}
