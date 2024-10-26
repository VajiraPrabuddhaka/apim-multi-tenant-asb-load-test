package worker

import (
	"apim-multi-tenant-asb-load-test/apis"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// StartRandomDeployments continuously deploys random API revisions using goroutines.
func StartRandomDeployments(data [][]string, authToken string, msgStore *sync.Map, concurrency int) {
	var ops atomic.Int64
	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency) // Semaphore for concurrency control

	rand.Seed(time.Now().UnixNano()) // Seed the random number generator

	fmt.Printf("data length: %d\n", len(data))
	for {
		// Acquire a semaphore slot before starting a goroutine.
		sem <- struct{}{}
		wg.Add(1)

		if int(ops.Load()) >= len(data) {
			ops.Store(0)
		}

		// Start a new goroutine for a random deployment.
		go func() {
			defer wg.Done()
			defer func() {
				<-sem
			}()

			value := ops.Load()
			intValue := int(value)

			randomSet := data[intValue]

			ops.Add(1)

			orgID := randomSet[0]
			dataPlaneID := randomSet[1]
			apiID := randomSet[2]
			revisionID := randomSet[3]

			msgStore.Store(apiID, time.Now())
			// Perform the API revision deployment.
			if err := apis.DeployAPIRevision(apiID, revisionID, orgID, dataPlaneID, authToken); err != nil {
				fmt.Printf("Error deploying API revision:(API_ID: %s, Revision_id: %s, orgID: %s, "+
					"dataPlaneId: %s) err:%v\n", apiID, revisionID, orgID, dataPlaneID, err)
			}
		}()

		// Add a random delay between 10ms and 50ms before spawning the next goroutine.
		randomDelay := time.Duration(rand.Intn(40)+10) * time.Millisecond
		time.Sleep(randomDelay)
	}

	// Ensure all goroutines complete before the program exits (unreachable in this case).
	wg.Wait()
}
