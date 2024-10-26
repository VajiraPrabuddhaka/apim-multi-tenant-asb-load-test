package apis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// CreateAPI sends the API creation request and returns the API ID.
func CreateAPI(name, orgID, authToken string) (string, error) {
	jsonPayload := fmt.Sprintf(`{
		"name": "%s",
		"description": "This API is used to connect to the TestAPI service",
		"context": "/%s",
		"version": "1.0.0",
		"lifeCycleStatus": "CREATED",
		"type": "HTTP",
		"transport": ["http", "https"],
		"policies": ["Bronze"],
		"visibility": "PUBLIC",
		"corsConfiguration": {
			"corsConfigurationEnabled": true,
			"accessControlAllowOrigins": ["*"],
			"accessControlAllowCredentials": false,
			"accessControlAllowHeaders": [
				"authorization", "Access-Control-Allow-Origin", "Content-Type",
				"SOAPAction", "testKey", "api-key", "X-Authorization"
			],
			"accessControlAllowMethods": ["GET", "PUT", "POST", "DELETE", "PATCH", "OPTIONS"]
		},
		"endpointConfig": {
			"endpoint_type": "http",
			"sandbox_endpoints": {
				"url": "https://geolocation.onetrust.com/cookieconsentpub/v1"
			},
			"production_endpoints": {
				"url": "https://geolocation.onetrust.com/cookieconsentpub/v1"
			}
		},
		"operations": [{
			"target": "/geo/location",
			"verb": "GET",
			"authType": "Application & Application User"
		}],
		"keyManagers": ["all"],
		"advertiseInfo": {
			"advertised": false,
			"apiOwner": "ca0c41b4-5bbd-48c8-b319-cf64d98e85b1",
			"vendor": "WSO2"
		}
	}`, name, name)

	url := fmt.Sprintf("%s?organizationId=%s&openAPIVersion=%s", apisBasePath, orgID, openAPIVersion)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonPayload)))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := insecureClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return apiResp.ID, nil
}
