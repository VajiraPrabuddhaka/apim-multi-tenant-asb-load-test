package apis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// CreateRevision sends the revision creation request and returns the revision ID.
func CreateRevision(apiID, orgID, authToken string) (string, error) {
	jsonPayload := `{
		"description": "first revision"
	}`

	url := fmt.Sprintf(revisionPath, apisBasePath, apiID) + fmt.Sprintf("?organizationId=%s", orgID)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonPayload)))
	if err != nil {
		return "", fmt.Errorf("failed to create revision request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := insecureClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("revision request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var revResp RevisionResponse
	if err := json.NewDecoder(resp.Body).Decode(&revResp); err != nil {
		return "", fmt.Errorf("failed to parse revision response: %w", err)
	}

	return revResp.ID, nil
}

// DeployAPIRevision sends a POST request to deploy an API revision.
func DeployAPIRevision(apiID, revisionID, organizationID, dataPlaneID, authToken string) error {
	url := fmt.Sprintf(
		"%s/%s/deploy-revision?revisionId=%s&organizationId=%s", apisBasePath, apiID, revisionID, organizationID,
	)

	// Derive name and vhost using dataPlaneID
	name := fmt.Sprintf("development-%s", dataPlaneID[len(dataPlaneID)-6:])
	vhost := fmt.Sprintf("%s-dev.choreo-dv.pdp.dev", name)

	// Prepare request body
	requestBody := []map[string]interface{}{
		{
			"name":               name,
			"displayOnDevportal": true,
			"vhost":              vhost,
		},
	}
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

	// Send request
	resp, err := insecureClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	fmt.Println("API revision deployed successfully.")
	return nil
}
