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
