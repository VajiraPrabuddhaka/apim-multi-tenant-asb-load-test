package messaging

import (
	"apim-multi-tenant-asb-load-test/asb_client"
	"apim-multi-tenant-asb-load-test/utils"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type EventPayload struct {
	Event struct {
		PayloadData PayloadData `json:"payloadData"`
	} `json:"event"`
}

type PayloadData struct {
	EventType string `json:"eventType"`
	Timestamp int64  `json:"timestamp"`
	Event     string `json:"event"`
}

type APIEvent struct {
	ApiID int    `json:"apiId"`
	UUID  string `json:"uuid"`
	Name  string `json:"name"`
}

var SentTimes = sync.Map{}

// ListenToChannel function for the common channel to print received messages.
func ListenToChannel(messageChan <-chan asb_client.Message, outputFileFaulty, outputFile *os.File) {
	for msg := range messageChan {
		// unmarshal the message into a struct
		var eventPayload EventPayload
		if err := json.Unmarshal([]byte(msg.Content), &eventPayload); err == nil {
			if eventPayload.Event.PayloadData.EventType == "DEPLOY_API_IN_GATEWAY" {
				decodedBytes, err := base64.StdEncoding.DecodeString(eventPayload.Event.PayloadData.Event)
				if err != nil {
					fmt.Printf("failed to decode base64: %s\n", err.Error())
				}
				// Unmarshal the JSON into the APIEvent struct
				var apiEvent APIEvent
				if err := json.Unmarshal(decodedBytes, &apiEvent); err != nil {
					fmt.Printf("failed to unmarshal JSON: %s\n", err.Error())
				}
				if t, ok := SentTimes.Load(apiEvent.UUID); ok {
					timestamp := t.(time.Time)
					if timestamp.Before(time.Now().Add(-1 * time.Minute)) {
						_, err := outputFileFaulty.WriteString(fmt.Sprintf("API UUID: %s, diff:%s\n", apiEvent.UUID, time.Now().Sub(timestamp).String()))
						if err != nil {
							fmt.Printf("failed to write to file: %s\n", err.Error())
						}
					} else {
						_, err := outputFile.WriteString(fmt.Sprintf("API UUID: %s, diff:%s\n", apiEvent.UUID, time.Now().Sub(timestamp).String()))
						if err != nil {
							fmt.Printf("failed to write to file: %s\n", err.Error())
						}
					}
				}
			}
		}
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
