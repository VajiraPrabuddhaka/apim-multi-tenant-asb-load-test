package apis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// EnvironmentRequest represents the JSON structure of the environment creation request
type EnvironmentRequest struct {
	Name                     string   `json:"name"`
	DisplayName              string   `json:"displayName"`
	Provider                 string   `json:"provider"`
	Description              string   `json:"description"`
	IsReadOnly               bool     `json:"isReadOnly"`
	DataPlaneId              string   `json:"dataPlaneId"`
	GatewayAccessibilityType string   `json:"gatewayAccessibilityType"`
	VHosts                   []VHost  `json:"vhosts"`
	EndpointURIs             []string `json:"endpointURIs"`
	AdditionalProperties     []string `json:"additionalProperties"`
}

// VHost represents a virtual host configuration
type VHost struct {
	Host        string `json:"host"`
	HttpContext string `json:"httpContext"`
	HttpPort    int    `json:"httpPort"`
	HttpsPort   int    `json:"httpsPort"`
	WsPort      int    `json:"wsPort"`
	WssPort     int    `json:"wssPort"`
}

// CreateEnvironment sends a POST request to create an environment
func CreateEnvironment(orgID, name, dataPlaneID, authToken string) error {
	// Define the request payload
	payload := EnvironmentRequest{
		Name:                     name,
		DisplayName:              name,
		Provider:                 "wso2",
		Description:              "Development US east azure",
		IsReadOnly:               false,
		DataPlaneId:              dataPlaneID,
		GatewayAccessibilityType: "external",
		VHosts: []VHost{
			{
				Host:        fmt.Sprintf("%s-dev.choreo-dv.pdp.dev", name),
				HttpContext: "",
				HttpPort:    80,
				HttpsPort:   443,
				WsPort:      9099,
				WssPort:     8099,
			},
		},
		EndpointURIs:         []string{},
		AdditionalProperties: []string{},
	}

	// Marshal payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	// Create the HTTP request
	url := fmt.Sprintf("%s?organizationId=%s", envsBasePath, orgID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", authToken))

	// Send the request
	resp, err := insecureClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	// Check for non-2xx status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("non-2xx status code: %d, response: %s", resp.StatusCode, string(body))
	}

	fmt.Println("Environment successfully created!")
	return nil
}
