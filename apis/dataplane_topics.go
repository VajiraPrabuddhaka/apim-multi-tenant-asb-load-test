package apis

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

// Topic represents the structure of topics in the response.
type Topic struct {
	TopicName        string `json:"topicName"`
	ConnectionString string `json:"connectionString"`
}

// RegisterResponse represents the full response structure.
type RegisterResponse struct {
	Message string  `json:"message"`
	Topics  []Topic `json:"topics"`
}

// RegisterDataplaneTopics sends a POST request to register dataplane topics.
func RegisterDataplaneTopics(orgID, dataPlaneID, authToken string) ([]Topic, error) {
	url := fmt.Sprintf(
		"%s/%s/register-dataplane-topics?organizationId=%s", dataplanesBasePath, dataPlaneID, orgID,
	)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", authToken)) // Use your base64-encoded credentials

	// Execute the request
	resp, err := insecureClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse JSON response
	var response RegisterResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	fmt.Printf("Successfully registered dataplane topics for DataPlaneID: %s, OrgID: %s\n", dataPlaneID, orgID)
	return response.Topics, nil
}

// appendTopicsToFile writes topics to the file atomically using a mutex.
func appendTopicsToFile(topics []Topic, filename string, mutex *sync.Mutex) error {
	// Lock the mutex to ensure atomic writes
	mutex.Lock()
	defer mutex.Unlock()

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	for _, topic := range topics {
		_, err := file.WriteString(fmt.Sprintf("%s\n%s\n", topic.TopicName, topic.ConnectionString))
		if err != nil {
			return fmt.Errorf("failed to write to file: %v", err)
		}
	}
	return nil
}

// processRegistration handles registering topics and writing them to the file.
func processRegistration(orgID, dataPlaneID, authToken, outputFileName string, mutex *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()

	topics, err := RegisterDataplaneTopics(orgID, dataPlaneID, authToken)
	if err != nil {
		log.Printf("Error registering topics for OrgID: %s, DataPlaneID: %s - %v", orgID, dataPlaneID, err)
		return
	}

	if err := appendTopicsToFile(topics, outputFileName, mutex); err != nil {
		log.Printf("Error writing topics for OrgID: %s, DataPlaneID: %s - %v", orgID, dataPlaneID, err)
	}
}

// CreateDataplaneTopicsFromFile reads the input file and processes registrations in parallel.
func CreateDataplaneTopicsFromFile(filename, authToken, outputFileName string, maxParallel int) error {
	file, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read input file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(file)), "\n")

	var wg sync.WaitGroup                   // WaitGroup to wait for all goroutines
	mutex := &sync.Mutex{}                  // Mutex to ensure atomic file writes
	sem := make(chan struct{}, maxParallel) // Semaphore to control parallelism

	for _, line := range lines {
		parts := strings.Split(line, ",")
		if len(parts) != 2 {
			log.Printf("Invalid line: %s", line)
			continue
		}
		orgID, dataPlaneID := parts[0], parts[1]

		// Acquire semaphore slot
		sem <- struct{}{}
		wg.Add(1)

		go func(orgID, dataPlaneID string) {
			defer func() { <-sem }() // Release semaphore slot
			processRegistration(orgID, dataPlaneID, authToken, outputFileName, mutex, &wg)
		}(orgID, dataPlaneID)
	}

	wg.Wait() // Wait for all goroutines to finish
	return nil
}

//// appendTopicsToFile writes topic names and connection strings to the specified file.
//func appendTopicsToFile(topics []Topic, filename string) error {
//	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
//	if err != nil {
//		return fmt.Errorf("failed to open file: %v", err)
//	}
//	defer file.Close()
//
//	for _, topic := range topics {
//		_, err := file.WriteString(fmt.Sprintf("%s\n%s\n", topic.TopicName, topic.ConnectionString))
//		if err != nil {
//			return fmt.Errorf("failed to write to file: %v", err)
//		}
//	}
//
//	return nil
//}
