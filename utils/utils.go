package utils

import (
	"apim-multi-tenant-asb-load-test/apis"
	"bufio"
	"bytes"
	"fmt"
	"github.com/google/uuid"
	"log"
	"os"
	"strings"
	"sync"
)

// LoadLinesFromFile reads lines from a text file and returns them as a slice of strings.
func LoadLinesFromFile(filename string) ([]string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	lines := bytes.Split(data, []byte("\n"))
	var result []string
	for _, line := range lines {
		if len(line) > 0 {
			result = append(result, string(line))
		}
	}
	return result, nil
}

// SaveLinesToFile saves a slice of strings to a text file, one per line.
func SaveLinesToFile(filename string, lines []string) error {
	var buffer bytes.Buffer
	for _, line := range lines {
		buffer.WriteString(line + "\n")
	}
	return os.WriteFile(filename, buffer.Bytes(), 0644)
}

// GenerateOrgAndDataPlaneIDs generates a specified number of UUIDs and writes them to a file.
func GenerateOrgAndDataPlaneIDs(numUUIDs int) {
	file, err := os.Create("organization_ids.txt")
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	for i := 0; i < numUUIDs; i++ {
		orgID := uuid.New().String()
		dataPlaneID := uuid.New().String()

		// Write <org_uuid>,<dataplane_id> to file
		_, err := file.WriteString(fmt.Sprintf("%s,%s\n", orgID, dataPlaneID))
		if err != nil {
			log.Fatalf("Failed to write to file: %v", err)
		}
	}

	fmt.Println("UUIDs successfully written to organization_ids.txt")
}

// ReadAsbTopicAndConnectionStringsFromFile Reads topics and connection strings from the file.
func ReadAsbTopicAndConnectionStringsFromFile(filename string) ([][2]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var configs [][2]string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		topic := strings.TrimSpace(scanner.Text())
		if !scanner.Scan() {
			return nil, fmt.Errorf("missing connection string for topic: %s", topic)
		}
		connStr := strings.TrimSpace(scanner.Text())
		configs = append(configs, [2]string{topic, connStr})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	return configs, nil
}

// ReadOrgAndDataPlaneIDs reads the org UUIDs and data plane IDs from a file.
func ReadOrgAndDataPlaneIDs(filename string) ([][2]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	var orgDataPlanePairs [][2]string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ",")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid line format: %s", line)
		}
		orgDataPlanePairs = append(orgDataPlanePairs, [2]string{parts[0], parts[1]})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	return orgDataPlanePairs, nil
}

// CreateEnvironmentsFromFile reads org and data plane IDs and creates environments in parallel.
func CreateEnvironmentsFromFile(filename, authToken string, maxParallel int) {
	orgDataPlanePairs, err := ReadOrgAndDataPlaneIDs(filename)
	if err != nil {
		log.Fatalf("Error reading org and data plane IDs: %v", err)
	}

	// Create a semaphore to control the number of parallel goroutines.
	sem := make(chan struct{}, maxParallel)
	var wg sync.WaitGroup

	for _, pair := range orgDataPlanePairs {
		orgID, dataPlaneID := pair[0], pair[1]
		name := fmt.Sprintf("development-%s", dataPlaneID[len(dataPlaneID)-6:]) // Example name generation

		// Acquire a semaphore slot.
		sem <- struct{}{}
		wg.Add(1)

		go func(orgID, name, dataPlaneID string) {
			defer wg.Done()

			// Create environment and handle errors.
			err := apis.CreateEnvironment(orgID, name, dataPlaneID, authToken)
			if err != nil {
				log.Printf("Failed to create environment for Org: %s, DataPlane: %s, Error: %v", orgID, dataPlaneID, err)
			} else {
				fmt.Printf("Successfully created environment for Org: %s, DataPlane: %s\n", orgID, dataPlaneID)
			}

			// Release the semaphore slot.
			<-sem
		}(orgID, name, dataPlaneID)
	}

	// Wait for all goroutines to complete.
	wg.Wait()
}

// LoadAPIData loads the API data from the given file.
func LoadAPIData(filename string) ([][]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	var apiData [][]string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ",")
		if len(parts) != 4 {
			return nil, fmt.Errorf("invalid data format in line: %s", line)
		}
		apiData = append(apiData, parts)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}
	return apiData, nil
}
