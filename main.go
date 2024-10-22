package main

import (
	"apim-multi-tenant-asb-load-test/apis"
	"apim-multi-tenant-asb-load-test/asb_client"
	"apim-multi-tenant-asb-load-test/messaging"
	"apim-multi-tenant-asb-load-test/utils"
	"context"
	"fmt"
	"sync"
)

const (
	apisBasePath   = "https://localhost:9444/api/am/publisher/v2/apis"
	revisionPath   = "%s/%s/revisions"
	authToken      = "e8c9f506-3cd6-3abd-b4eb-2cd1501bc651"
	authTokenBasic = "YWRtaW46YWRtaW4="
	concurrentReq  = 100 // Number of concurrent requests
	openAPIVersion = "v3"
	orgIDsFile     = "organization_ids.txt"
	apiIDsFile     = "api_ids.txt"
	topicsFile     = "topics.txt"
)

// ##########################Start Listeners############################
func main() {
	// Create a buffered channel for messages.
	messageChan := make(chan asb_client.Message, 100)

	// Create a wait group to synchronize all goroutines.
	var wg sync.WaitGroup

	// Context with cancellation for graceful shutdown.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	messaging.CreateTopicListeners(ctx, "topics.txt", messageChan, &wg)

	// Start a goroutine to listen on the common channel.
	go messaging.ListenToChannel(messageChan)

	// Wait for all goroutines to finish.
	wg.Wait()
}

//func main() {
//	//utils.CreateEnvironmentsFromFile("or")
//	//utils.GenerateOrgAndDataPlaneIDs(10)
//	//utils.CreateEnvironmentsFromFile(orgIDsFile, authTokenBasic, 10)
//	if err := apis.CreateDataplaneTopicsFromFile(orgIDsFile, authTokenBasic, topicsFile, 5); err != nil {
//		log.Fatalf("Error: %v", err)
//	}
//}

func CreateApisAndRevisions() {
	// Load organization IDs from file.
	orgIDs, err := utils.LoadLinesFromFile(orgIDsFile)
	if err != nil {
		fmt.Println("Error reading organization IDs:", err)
		return
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var apiRevisions []string

	for i, orgID := range orgIDs {
		wg.Add(1)
		go func(i int, orgID string) {
			defer wg.Done()

			location := fmt.Sprintf("location%06d", i+1)
			apiID, err := apis.CreateAPI(location, orgID, authToken)
			if err != nil {
				fmt.Printf("Failed to create API for %s: %v\n", location, err)
				return
			}

			revisionID, err := apis.CreateRevision(apiID, orgID, authToken)
			if err != nil {
				fmt.Printf("Failed to create revision for API %s: %v\n", apiID, err)
				return
			}

			// Collect API ID and revision ID together.
			entry := fmt.Sprintf("%s,%s", apiID, revisionID)

			mu.Lock()
			apiRevisions = append(apiRevisions, entry)
			mu.Unlock()
		}(i, orgID)
	}

	wg.Wait()

	// Save API IDs to file for future use if needed.
	if err := utils.SaveLinesToFile(apiIDsFile, apiRevisions); err != nil {
		fmt.Println("Error saving API IDs:", err)
	}
	fmt.Println("Finished creating APIs and their revisions.")
}
