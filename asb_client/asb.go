package asb_client

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus/admin"
	"log"
	"math/rand"
	"sync"
	"time"
)

// Message struct to store topic and message body information.
type Message struct {
	Topic   string
	Content string
}

// Generates a random subscription name.
func generateRandomSubscriptionName() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("sub-%d", rand.Intn(100000))
}

// Creates a subscription for a given topic with a random name.
func createSubscription(ctx context.Context, adminClient *admin.Client, topicName string) string {
	subscriptionName := generateRandomSubscriptionName()

	_, err := adminClient.CreateSubscription(ctx, topicName, subscriptionName, nil)
	if err != nil {
		log.Fatalf("Failed to create subscription: %v", err)
	}

	log.Printf("Created subscription: %s for topic: %s", subscriptionName, topicName)
	return subscriptionName
}

// CreateASBListener function that creates a Service Bus receiver and listens to messages.
func CreateASBListener(ctx context.Context, connStr, topicName string, messageChan chan<- Message, wg *sync.WaitGroup) {
	defer wg.Done()

	// Create an admin client to manage topics and subscriptions.
	adminClient, err := admin.NewClientFromConnectionString(connStr, nil)
	if err != nil {
		log.Fatalf("Failed to create admin client: %v", err)
	}

	// Create a subscription with a random name.
	subscriptionName := createSubscription(ctx, adminClient, topicName)

	// Create a Service Bus client.
	client, err := azservicebus.NewClientFromConnectionString(connStr, nil)
	if err != nil {
		log.Fatalf("Failed to create Service Bus client: %v", err)
	}
	defer client.Close(ctx)

	// Create a receiver for the topic and the new subscription.
	receiver, err := client.NewReceiverForSubscription(topicName, subscriptionName, nil)
	if err != nil {
		log.Fatalf("Failed to create receiver: %v", err)
	}
	defer receiver.Close(ctx)

	log.Printf("Listening on topic: %s, subscription: %s", topicName, subscriptionName)

	// Continuously receive messages.
	for {
		msgs, err := receiver.ReceiveMessages(ctx, 1, nil)
		if err != nil {
			log.Printf("Error receiving message: %v", err)
			return
		}

		for _, msg := range msgs {
			// Push the received message to the channel.
			messageChan <- Message{Topic: topicName, Content: string(msg.Body)}

			// Complete the message to remove it from the queue.
			if err := receiver.CompleteMessage(ctx, msg, nil); err != nil {
				log.Printf("Failed to complete message: %v", err)
			}
		}
	}
}
