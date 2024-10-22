package messaging

import (
	"apim-multi-tenant-asb-load-test/asb_client"
	"apim-multi-tenant-asb-load-test/utils"
	"context"
	"fmt"
	"log"
	"sync"
)

// ListenToChannel function for the common channel to print received messages.
func ListenToChannel(messageChan <-chan asb_client.Message) {
	for msg := range messageChan {
		fmt.Printf("Received message from topic '%s': %s\n", msg.Topic, msg.Content)
	}
}

// CreateTopicListeners function to create listeners for each topic.
func CreateTopicListeners(ctx context.Context, topicsFilePath string, messageChan chan<- asb_client.Message, wg *sync.WaitGroup) {
	configs, err := utils.ReadAsbTopicAndConnectionStringsFromFile(topicsFilePath)
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	for _, config := range configs {
		topicName := config[0]
		connStr := config[1]

		wg.Add(1)
		go asb_client.CreateASBListener(ctx, connStr, topicName, messageChan, wg)
	}
}
