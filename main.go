package main

import (
	"apim-multi-tenant-asb-load-test/apis"
	"apim-multi-tenant-asb-load-test/asb_client"
	"apim-multi-tenant-asb-load-test/messaging"
	"apim-multi-tenant-asb-load-test/utils"
	"apim-multi-tenant-asb-load-test/worker"
	"context"
	"fmt"
	"log"
	"os"
	"sync"
)

const (
	authToken      = "17a56c85-f70b-3c83-ba8f-3cc75c50292d"
	authTokenBasic = "YWRtaW46YWRtaW4="

	orgIDsFile = "organization_ids.txt"
	apiIDsFile = "api_ids.txt"
	topicsFile = "topics.txt"
)

func main() {
	utils.GenerateOrgAndDataPlaneIDs(500)
	log.Printf("Organization IDs and Data Plane IDs generated and saved to %s\n", orgIDsFile)
	time.Sleep(10 * time.Second)

	utils.CreateEnvironmentsFromFile(orgIDsFile, authTokenBasic, 10)
	log.Printf("Environments created and saved to %s\n", topicsFile)
	time.Sleep(10 * time.Second)

	if err := apis.CreateDataplaneTopicsFromFile(orgIDsFile, authTokenBasic, topicsFile, 10); err != nil {
		log.Fatalf("Error: %v", err)
	}
	log.Printf("Topics created and saved to %s\n", topicsFile)
	time.Sleep(10 * time.Second)

	CreateApisAndRevisions(10)
	log.Printf("APIs and revisions created and saved to %s\n", apiIDsFile)
	time.Sleep(10 * time.Second)

	log.Printf("Starting random deployments...\n")
	apiData, err := utils.LoadAPIData(apiIDsFile)
	if err != nil {
		fmt.Printf("Error loading API data: %v\n", err)
		return
	}

	// Create a buffered channel for messages.
	messageChan := make(chan asb_client.Message, 20)

	// Create a wait group to synchronize all goroutines.
	var wg sync.WaitGroup

	// Context with cancellation for graceful shutdown.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	messaging.CreateTopicListeners(ctx, topicsFile, messageChan, &wg)

	outputFileFaulty, err := os.Create("time_differences_faulty.txt")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}

	outputFile, err := os.Create("time_differences.txt")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}

	// Start a goroutine to listen on the common channel.
	go messaging.ListenToChannel(messageChan, outputFileFaulty, outputFile)

	go worker.StartRandomDeployments(apiData, authToken, &messaging.SentTimes, 70)

	// Wait for all goroutines to finish.
	wg.Wait()
}

func CreateApisAndRevisions(maxParallel int) {
	// Load organization IDs from file.
	orgDataPlanePairs, err := utils.ReadOrgAndDataPlaneIDs(orgIDsFile)
	if err != nil {
		fmt.Println("Error reading organization IDs:", err)
		return
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	// Create a semaphore to control the number of parallel goroutines.
	sem := make(chan struct{}, maxParallel)

	var apiRevisions []string

	for i, orgDataplaneIDPair := range orgDataPlanePairs {
		sem <- struct{}{}
		wg.Add(1)
		go func(i int, orgDataplaneIDPair [2]string) {
			defer wg.Done()

			name := fmt.Sprintf("location%s", orgDataplaneIDPair[0][len(orgDataplaneIDPair[0])-6:])
			apiID, err := apis.CreateAPI(name, orgDataplaneIDPair[0], authToken)
			if err != nil {
				fmt.Printf("Failed to create API for %s: %v\n", name, err)
				return
			}

			revisionID, err := apis.CreateRevision(apiID, orgDataplaneIDPair[0], authToken)
			if err != nil {
				fmt.Printf("Failed to create revision for API %s: %v\n", apiID, err)
				return
			}

			// Collect API ID and revision ID together.
			entry := fmt.Sprintf("%s,%s,%s,%s", orgDataplaneIDPair[0], orgDataplaneIDPair[1], apiID, revisionID)

			mu.Lock()
			apiRevisions = append(apiRevisions, entry)
			mu.Unlock()

			<-sem
		}(i, orgDataplaneIDPair)
	}

	wg.Wait()

	// Save API IDs to file for future use if needed.
	if err := utils.SaveLinesToFile(apiIDsFile, apiRevisions); err != nil {
		fmt.Println("Error saving API IDs:", err)
	}
	fmt.Println("Finished creating APIs and their revisions.")
}
