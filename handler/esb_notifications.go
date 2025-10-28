package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"mainPackage/model"
	"os"
	"strconv"
	"time"

	"github.com/IBM/sarama"
)

func ESB_NOTIFICATIONS() error {
	maxRetryStr := os.Getenv("KAFKA_RETRY")
	intervalStr := os.Getenv("KAFKA_INTERVAL")
	maxRetryInt, err_ := strconv.Atoi(maxRetryStr)
	if err_ != nil {
		fmt.Println("Invalid KAFKA_INTERVAL, using default of 10 seconds")
		maxRetryInt = 10 // default fallback
	}
	intervalInt, err_ := strconv.Atoi(intervalStr)
	if err_ != nil {
		fmt.Println("Invalid KAFKA_INTERVAL, using default of 10 seconds")
		intervalInt = 10 // default fallback
	}
	var (
		brokers       = []string{os.Getenv("ESB_SERVER")}
		topic         = os.Getenv("ESB_NOTIFICATIONS")
		maxRetry      = maxRetryInt
		retryInterval = time.Duration(intervalInt) * time.Second
	)

	var consumer sarama.Consumer
	var err error

	for attempt := 1; attempt <= maxRetry; attempt++ {
		log.Printf("Attempt %d to connect to "+topic+" brokers: %v", attempt, brokers)
		consumer, err = sarama.NewConsumer(brokers, nil)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to "+topic+": %v", err)
		time.Sleep(retryInterval)
	}

	if err != nil {
		return fmt.Errorf("could not connect to "+topic+" after %d attempts: %w", maxRetry, err)
	}

	defer func() {
		if err := consumer.Close(); err != nil {
			log.Printf("Error closing "+topic+" consumer: %v", err)
		}
	}()

	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		log.Printf("Error starting partition consumer: %v", err)
		return err
	}
	defer func() {
		if err := partitionConsumer.Close(); err != nil {
			log.Printf("Error closing ESB_NOTIFICATIONS() consumer: %v", err)
		}
	}()

	log.Println("Kafka ESB_NOTIFICATIONS() started. Listening for messages...")

	for msg := range partitionConsumer.Messages() {
		raw := string(msg.Value)
		log.Printf("📩 Received message: %s", raw)

		var data map[string]interface{}
		if err := json.Unmarshal(msg.Value, &data); err != nil {
			log.Printf("❌ Failed to parse JSON: %v", err)
			continue
		}

		// Example: Access fields safely
		if event, ok := data["EVENT"]; ok {
			log.Printf("EVENT Type: %v", event)
		}

		var notification model.Notification
		if err := json.Unmarshal(msg.Value, &notification); err != nil {
			log.Printf("❌ Failed to unmarshal JSON to Notification: %v", err)
			continue
		}

		// Example: Access fields from struct
		if notification.Event != nil {
			log.Printf("✅ Event: %s | Sender: %s | Message: %s",
				*notification.Event,
				notification.Sender,
				raw)
		}

		go BroadcastNotification(notification)
	}

	return nil
}
