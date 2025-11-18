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
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
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
		log.Print(string(msg.Value))
		go handleMessage_WO_Create(&gin.Context{}, msg.Value)
	}

	return nil
}

func handleMessage_WO_Create(c *gin.Context, message []byte) {
	source := os.Getenv("INTEGRATION_SOURCE")
	log.Printf("Create :: Raw message: %s", string(message))
	var wo model.WorkOrder
	if err := json.Unmarshal(message, &wo); err != nil {
		log.Printf("Error unmarshalling message: %v", err)
		return
	}

	// Check owner -- Start
	hostname, err := os.Hostname()
	if err != nil {
		log.Println("Hostname error:", err)
		hostname = "unknown"
	}

	log.Printf("Recheck ESB Create WO on host: %s\n", hostname)

	val, err := utils.EsbCreateGet(wo.IncidentNumber)
	if err != nil {
		log.Println("Redis GET error:", err)
		return
	}

	if val == "" || val == hostname {
		err = utils.EsbCreateSet(wo.WorkOrderNumber, hostname)
		if err != nil {
			log.Println("Redis SET error:", err)
			return
		}
	} else {
		log.Printf("Skip Message Not allow Owner : %s\n", hostname)
		return
	}
	// Check owner -- End

	if wo.Source == source {
		log.Printf("Skip Message From Original Source : %s\n", source)
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

	c.Set("username", username)
	c.Set("orgId", orgId)

	if err := IntegrateCreateCaseFromWorkOrder(c, conn, wo, username, orgId); err != nil {
		log.Printf("Error creating case from WorkOrder: %v", err)
	}
}

func ESB_WORK_ORDER_UPDATE() error {

	// Check owner -- Start
	// hostname, err := os.Hostname()
	// if err != nil {
	// 	log.Println("Hostname error:", err)
	// 	hostname = "unknown"
	// }
	// log.Printf("Recheck ESB Update WO on host: %s\n", hostname)
	// if val == "" || val == hostname {
	// 	err = utils.EsbCreateSet(wo.WorkOrderNumber, hostname)
	// 	if err != nil {
	// 		log.Println("Redis SET error:", err)
	// 		return
	// 	}
	// } else {
	// 	log.Printf("Skip Connection Not allow Owner : %s\n", hostname)
	// 	return
	// }
	// Check owner -- End

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
		log.Print(string(msg.Value))
		go handleMessage_WO_Update(&gin.Context{}, msg.Value)
	}

	return nil
}

func handleMessage_WO_Update(c *gin.Context, message []byte) {
	source := os.Getenv("INTEGRATION_SOURCE")
	log.Printf("Update :: Raw message: %s", string(message))

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

	username := os.Getenv("INTEGRATION_USR")
	orgId := os.Getenv("INTEGRATION_ORG_ID")

	c.Set("username", username)
	c.Set("orgId", orgId)

	if wo.Status == "NEW" {
		return
	}
	if wo.Status == "CANCEL" {
		log.Printf("WorkOrder %s cancelled, processing...", wo.WorkOrderNumber)

		resId := os.Getenv("RESULT_CLOSE")

		err := CancelCaseCore(c, conn, orgId, username, wo.WorkOrderNumber, resId, "cancel from workorder integration")
		if err != nil {
			log.Printf("cancel failed: %v", err)
		}
		return
	}

	if wo.Status == "ASSIGNED" {

		UserEmployeeCode := wo.UserMetadata.AssignedEmployeeCode.UserEmployeeCode
		if UserEmployeeCode != "" {

			unitLists, count, err := GetUnitsWithDispatch(ctx, conn, orgId, wo.WorkOrderNumber, "S003", "")
			if err != nil {
				panic(err)
			}

			log.Printf(" %s - Total Units: %d", wo.WorkOrderNumber, count)
			log.Print(unitLists)

			// ตรวจสอบว่ามีข้อมูลหรือไม่
			if len(unitLists) > 0 {
				firstUnit := unitLists[0]
				log.Printf("First Unit: %+v", firstUnit)

				// ตัวอย่างเช็คค่าเฉพาะ field
				log.Printf("First Unit ID: %s", firstUnit.UnitID)
				log.Printf("First Unit Username: %s", firstUnit.Username)

				req := model.CancelUnitRequest{
					CaseId:    wo.WorkOrderNumber,
					ResDetail: "Cancelled unit by integration",
					ResId:     "",
					UnitId:    firstUnit.UnitID,
					UnitUser:  firstUnit.Username,
				}

				err := DispatchCancelUnitCore(c, conn, req, os.Getenv("INTEGRATION_ORG_ID"), os.Getenv("INTEGRATION_USR"))
				if err != nil {
					log.Printf("❌ Cancel Unit failed: %v", err)
				} else {
					log.Print("✅ Unit cancelled successfully")
				}

			} else {
				log.Print("No units found.")
			}

			// user, err := utils.GetUserByUsername(ctx, conn, orgId, UserEmployeeCode)
			// if err != nil {
			// 	log.Printf("Error getting user: %v", err)
			// } else if user != nil {

			// 	log.Printf("-->ASSIGNED : %s , %s", user.EmpID, UserEmployeeCode)
			// 	if user.EmpID != UserEmployeeCode {

			// 		//Cancel Unit
			// 		req := model.CancelUnitRequest{
			// 			CaseId:    wo.WorkOrderNumber,
			// 			ResDetail: "Cancelled unit by integration",
			// 			ResId:     "",
			// 			UnitId:    UserEmployeeCode,
			// 			UnitUser:  UserEmployeeCode,
			// 		}

			// 		err := DispatchCancelUnitCore(c, conn, req, os.Getenv("INTEGRATION_ORG_ID"), os.Getenv("INTEGRATION_USR"))
			// 		if err != nil {
			// 			log.Printf("❌ Cancel Unit failed: %v", err)
			// 		} else {
			// 			log.Print("✅ Unit cancelled successfully")
			// 		}

			// 	}

			// }
		}

	}

	log.Print("====1===")

	if err := IntegrateUpdateCaseFromWorkOrder(c, conn, wo, username, orgId); err != nil {
		log.Printf("Error creating case from WorkOrder: %v", err)
	}
}

func CancelCaseCore(ctx *gin.Context, conn *pgx.Conn, orgId, username, caseId, resId, resDetail string) error {
	logger := utils.GetLog()

	cancelStatus := os.Getenv("CANCEL_CASE")

	// Delete all unit
	// deletedCount, err := DeleteCurrentUnit(ctx, conn, orgId, caseId, "", "")
	// if err != nil {
	// 	return fmt.Errorf("delete unit failed: %w", err)
	// }
	// if deletedCount > 0 {
	// 	log.Printf("deleted %d current units", deletedCount)
	// }

	// Update cancel
	err := UpdateCancelCaseForUnit(ctx, conn, orgId, caseId, resId, resDetail, cancelStatus, username)
	if err != nil {
		return fmt.Errorf("update cancel case failed: %w", err)
	}

	// Notifications + Kafka Backfeed
	req := model.UpdateStageRequest{
		CaseId: caseId,
		Status: cancelStatus,
	}
	GenerateNotiAndComment(ctx, conn, req, orgId, "0")
	//UpdateBusKafka_WO(ctx, conn, req)

	logger.Info("case cancelled successfully", zap.String("caseId", caseId))
	return nil
}
