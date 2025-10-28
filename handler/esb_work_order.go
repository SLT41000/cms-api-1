package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"mainPackage/model"
	"mainPackage/utils"
	"os"
	"strconv"
	"time"

	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
)

func ESB_WORK_ORDER_CREATE() error {
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
		topic         = os.Getenv("ESB_WO_CREATE")
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
			log.Printf("Error closing ESB_WORK_ORDER_CREATE() consumer: %v", err)
		}
	}()

	log.Println("Kafka ESB_WORK_ORDER_CREATE() started. Listening for messages...")

	for msg := range partitionConsumer.Messages() {
		go handleMessage_WO_Create(msg.Value)
	}

	return nil
}

func handleMessage_WO_Create(message []byte) {
	source := os.Getenv("INTEGRATION_SOURCE")
	log.Printf("Raw message: %s", string(message))
	var wo model.WorkOrder
	if err := json.Unmarshal(message, &wo); err != nil {
		log.Printf("Error unmarshalling message: %v", err)
		return
	}

	if wo.Source == source {
		log.Printf("Skip Message From Original Source : %s", source)
		return
	}
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		log.Printf("DB connection is nil")
		return
	}
	defer cancel()
	defer conn.Close(ctx)

	username := os.Getenv("INTEGRATION_USR")
	orgId := os.Getenv("INTEGRATION_ORG_ID")

	if err := IntegrateCreateCaseFromWorkOrder(ctx, conn, wo, username, orgId); err != nil {
		log.Printf("Error creating case from WorkOrder: %v", err)
	}
}

func ESB_WORK_ORDER_UPDATE() error {
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
		topic         = os.Getenv("ESB_WO_UPDATE")
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
			log.Printf("Error closing ESB_WORK_ORDER_UPDATE() consumer: %v", err)
		}
	}()

	log.Println("Kafka ESB_WORK_ORDER_UPDATE() started. Listening for messages...")

	for msg := range partitionConsumer.Messages() {
		go handleMessage_WO_Update(&gin.Context{}, msg.Value)
	}

	return nil
}

func handleMessage_WO_Update(c *gin.Context, message []byte) {
	source := os.Getenv("INTEGRATION_SOURCE")
	log.Printf("Raw message: %s", string(message))
	var wo model.WorkOrder
	if err := json.Unmarshal(message, &wo); err != nil {
		log.Printf("Error unmarshalling message: %v", err)
		return
	}

	if wo.Source == source {
		log.Printf("Skip Message From Original Source : %s", source)
		return
	}
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		log.Printf("DB connection is nil")
		return
	}
	defer cancel()
	defer conn.Close(ctx)

	log.Print("==handleMessage_WO_Update==")
	log.Print(wo)

	//Workorder Status
	// •	NEW
	// •	ASSIGNED
	// •	ACKNOWLEDGE
	// •	INPROGRESS
	// •	DONE
	// •	ONHOLD
	// •	CANCEL

	if wo.Status == "NEW" || wo.Status == "ASSIGNED" {
		return
	}
	log.Print("====1===")
	username := os.Getenv("INTEGRATION_USR")
	orgId := os.Getenv("INTEGRATION_ORG_ID")
	c.Set("username", username)
	c.Set("orgId", orgId)

	if err := IntegrateUpdateCaseFromWorkOrder(c, conn, wo, username, orgId); err != nil {
		log.Printf("Error creating case from WorkOrder: %v", err)
	}
}

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

	log.Println("Kafka ESB_USER_STATUS() started. Listening for messages...")

	for msg := range partitionConsumer.Messages() {
		go xx(&gin.Context{}, msg.Value)
	}

	return nil
}

func xx(c *gin.Context, message []byte) {
	log.Print(string(message))
}
