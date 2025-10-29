package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"mainPackage/model"
	"mainPackage/utils"
	"os"
	"strconv"
	"time"

	"github.com/IBM/sarama"
)

func ESB_USER_STATUS() error {
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
		topic         = os.Getenv("ESB_USER_STATUS")
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
			log.Printf("Error closing ESB_USER_STATUS() consumer: %v", err)
		}
	}()

	log.Println("Kafka ESB_USER_STATUS() started. Listening for messages...", os.Getenv("ESB_USER_STATUS"))
	username_ := os.Getenv("INTEGRATION_USR")
	orgId_ := os.Getenv("INTEGRATION_ORG_ID")
	for msg := range partitionConsumer.Messages() {
		var status model.KafkaUnitStatus
		if err := json.Unmarshal(msg.Value, &status); err != nil {
			log.Printf("❌ Failed to parse Kafka message: %v", err)
			continue
		}

		log.Printf("📦 Received unit status: %+v", status)

		statusId := "000"
		if status.CheckedIn {
			statusId = "001"
		}
		// Example: convert to your DB update fields
		err = UpdateUnitStatus(
			orgId_,           // orgId
			status.UserCode,  // username
			status.CheckedIn, // isLogin
			statusId,         // sttId (if not in message)
			"",               // locLat
			"",               // locLon
			username_,
		)
		if err != nil {
			log.Printf("❌ Failed to update unit: %v", err)
		}
	}

	return nil
}

// UpdateUnitStatus updates login state, location, and status of a unit by username + orgId
func UpdateUnitStatus(orgId, username string, isLogin bool, sttId, locLat, locLon string, username_ string) error {
	var lat, lon sql.NullFloat64

	if locLat != "" {
		if f, err := strconv.ParseFloat(locLat, 64); err == nil {
			lat = sql.NullFloat64{Float64: f, Valid: true}
		}
	}
	if locLon != "" {
		if f, err := strconv.ParseFloat(locLon, 64); err == nil {
			lon = sql.NullFloat64{Float64: f, Valid: true}
		}
	}

	log.Printf("🧩 [DEBUG] UpdateUnitStatus: orgId=%s, username=%s, isLogin=%t, sttId=%s, lat=%v, lon=%v, updatedBy=%s",
		orgId, username, isLogin, sttId, lat, lon, username_)

	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return fmt.Errorf("database connection is nil")
	}
	defer cancel()
	defer conn.Close(ctx)

	query := `
		UPDATE mdm_units
		SET 
			"isLogin" = $1,
			"sttId" = $2,
			"locLat" = $3,
			"locLon" = $4,
			"updatedAt" = NOW(),
			"updatedBy" = $5
		WHERE "orgId" = $6 AND "username" = $7
		RETURNING "unitId";
	`

	var updatedUnit string
	err := conn.QueryRow(ctx, query,
		isLogin,
		sttId,
		lat,
		lon,
		username_,
		orgId,
		username,
	).Scan(&updatedUnit)

	if err != nil {
		log.Printf("❌ [ERROR] Failed to update unit: %v", err)
		return fmt.Errorf("failed to update unit (orgId=%s, username=%s): %w", orgId, username, err)
	}

	log.Printf("✅ [SUCCESS] Unit %s updated (orgId=%s, username=%s)", updatedUnit, orgId, username)
	return nil
}
