package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mainPackage/model"
	"mainPackage/utils"
	"os"
	"strconv"
	"time"

	"github.com/IBM/sarama"
)

func ESB_USER_CREATE(type_ string) error {
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
	TOPIC := os.Getenv("ESB_USER_STAFF_CREATE")
	if type_ == "ADMIN" {
		TOPIC = os.Getenv("ESB_USER_ADMIN_CREATE")
	}
	var (
		brokers       = []string{os.Getenv("ESB_SERVER")}
		topic         = TOPIC
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
			log.Printf("Error closing %s() consumer: %v", type_, err)
		}
	}()

	log.Printf("Kafka %s() started. Listening for messages...%v", type_, TOPIC)
	// username_ := os.Getenv("INTEGRATION_USR")
	// orgId_ := os.Getenv("INTEGRATION_ORG_ID")
	for msg := range partitionConsumer.Messages() {
		//---> Call funtion for insert or update by user_code
		log.Print(string(msg.Value))
		var payload model.ESBUserStaffPayload
		if err := json.Unmarshal(msg.Value, &payload); err != nil {
			log.Println("❌ JSON Unmarshal error:", err)
			continue
		}

		if err := UpsertUserFromESB(type_, payload); err != nil {
			log.Println("❌ Upsert error:", err)
		} else {
			log.Println("✅ Create User processed:", payload.Username)
		}
	}

	return nil
}

func ESB_USER_UPDATE(type_ string) error {
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

	TOPIC := os.Getenv("ESB_USER_STAFF_UPDATE")
	if type_ == "ADMIN" {
		TOPIC = os.Getenv("ESB_USER_ADMIN_UPDATE")
	}

	var (
		brokers       = []string{os.Getenv("ESB_SERVER")}
		topic         = TOPIC
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
			log.Printf("Error closing %s() consumer: %v", type_, err)
		}
	}()

	log.Printf("Kafka %s() started. Listening for messages...%v", type_, TOPIC)
	// username_ := os.Getenv("INTEGRATION_USR")
	// orgId_ := os.Getenv("INTEGRATION_ORG_ID")
	for msg := range partitionConsumer.Messages() {
		//---> Call funtion for insert or update by user_code
		log.Print(string(msg.Value))
		var payload model.ESBUserStaffPayload
		if err := json.Unmarshal(msg.Value, &payload); err != nil {
			log.Println("❌ JSON Unmarshal error:", err)
			continue
		}

		if err := UpsertUserFromESB(type_, payload); err != nil {
			log.Println("❌ Upsert error:", err)
		} else {
			log.Println("✅ Update User processed:", payload.Username)
		}
	}

	return nil
}

func ESB_USER_DELETE(type_ string) error {
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

	TOPIC := os.Getenv("ESB_USER_STAFF_DELETE")
	if type_ == "ADMIN" {
		TOPIC = os.Getenv("ESB_USER_ADMIN_DELETE")
	}

	var (
		brokers       = []string{os.Getenv("ESB_SERVER")}
		topic         = TOPIC
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
			log.Printf("Error closing %s() consumer: %v", type_, err)
		}
	}()

	log.Printf("Kafka %s() started. Listening for messages...%v", type_, os.Getenv("ESB_USER_STAFF_DELETE"))
	// username_ := os.Getenv("INTEGRATION_USR")
	// orgId_ := os.Getenv("INTEGRATION_ORG_ID")
	for msg := range partitionConsumer.Messages() {
		//---> Call funtion for insert or update by user_code
		log.Print(string(msg.Value))
		var payload model.ESBUserStaffPayload
		if err := json.Unmarshal(msg.Value, &payload); err != nil {
			log.Println("❌ JSON Unmarshal error:", err)
			continue
		}

		// if err := DeleteUserFromESB(payload); err != nil {
		// 	log.Println("❌ Upsert error:", err)
		// } else {
		// 	log.Println("✅ User processed:", payload.Username)
		// }

		if err := DeActivateUserFromESB(payload); err != nil {
			log.Println("❌ Upsert error:", err)
		} else {
			log.Println("✅ Disable User processed:", payload.Username)
		}
	}

	return nil
}

func UpsertUserFromESB(type_ string, p model.ESBUserStaffPayload) error {

	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return fmt.Errorf("database connection is nil")
	}
	defer cancel()
	defer conn.Close(ctx)

	username := os.Getenv("INTEGRATION_USR")
	orgId := os.Getenv("INTEGRATION_ORG_ID")

	deptId := os.Getenv("USER_DEPT_ID")
	commId := os.Getenv("USER_COMM_ID")
	stnId := os.Getenv("USER_STN_ID")

	roleId := os.Getenv("USER_ROLE")
	if type_ == "ADMIN" {
		roleId = os.Getenv("ADMIN_ROLE")
	}

	// 1) เช็คว่ามีอยู่หรือยัง
	var exists bool
	err := conn.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM um_users WHERE  "orgId"=$1 AND (username=$2 OR "empId"=$3)  )`,
		orgId, p.Username, p.UserCode,
	).Scan(&exists)
	log.Print(err)
	if err != nil {
		return err
	}

	fmt.Println("DEBUG PARAM TYPES:")
	fmt.Printf("roleId      (%T): %v\n", roleId, roleId)
	fmt.Printf("deptId      (%T): %v\n", deptId, deptId)
	fmt.Printf("commId      (%T): %v\n", commId, commId)
	fmt.Printf("stnId       (%T): %v\n", stnId, stnId)
	fmt.Printf("username    (%T): %v\n", username, username)

	if !exists {
		// ---------- INSERT ----------
		enc, _ := encrypt(p.Password)
		_, err := conn.Exec(ctx, `
				INSERT INTO um_users(
					"orgId","displayName","empId","firstName","lastName","mobileNo",
					"email","username","password","roleId","userType",
					"deptId","commId","stnId","active",
					"createdAt","updatedAt","createdBy","updatedBy"
				)
				VALUES (
					$1::uuid, $2, $3, $4, $5, $6,
					$7, $8, $9, $10::uuid, 1,
					$11::uuid, $12::uuid, $13::uuid, true,
					NOW(), NOW(), $14, $15
				)
			`,
			orgId,
			p.FirstName+" "+p.LastName,
			p.UserCode,
			p.FirstName,
			p.LastName,
			p.Phone,
			p.Email,
			p.Username,
			enc,
			os.Getenv("USER_ROLE"),
			os.Getenv("USER_DEPT_ID"),
			os.Getenv("USER_COMM_ID"),
			os.Getenv("USER_STN_ID"),
			username,
			username,
		)

		log.Print(err)

		if err != nil {
			return err
		}

		return nil
	}

	// ---------- UPDATE ----------
	displayName := p.FirstName + " " + p.LastName
	enc, _ := encrypt(p.Password)
	_, err = conn.Exec(ctx, `
		UPDATE um_users
		SET
			"displayName" = $1,
			"firstName" = $2,
			"lastName" = $3,
			"mobileNo" = $4,
			"email" = $5,
			"password" = $6,
			"roleId" = $7::uuid,
			"deptId" = $8::uuid,
			"commId" = $9::uuid,
			"stnId" = $10::uuid,
			"updatedAt" = NOW(),
			"updatedBy" = $11
		WHERE "orgId" = $12::uuid AND "empId" = $13
	`,
		displayName, p.FirstName, p.LastName, p.Phone, p.Email, enc,
		roleId, deptId, commId, stnId, username,
		orgId, p.UserCode,
	)

	return err
}

func DeleteUserFromESB(p model.ESBUserStaffPayload) error {

	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return fmt.Errorf("database connection is nil")
	}
	defer cancel()
	defer conn.Close(ctx)

	orgId := os.Getenv("INTEGRATION_ORG_ID")
	username := os.Getenv("INTEGRATION_USR")
	empId := p.UserCode
	// ---- DELETE by orgId + empId ----
	_, err := conn.Exec(ctx, `
		DELETE FROM um_users
		WHERE "orgId" = $1::uuid AND "empId" = $2
	`,
		orgId,
		empId,
	)

	if err != nil {
		return fmt.Errorf("delete user failed: %w", err)
	}

	log.Printf("✔ User deleted orgId=%s empId=%s by %s", orgId, empId, username)
	return nil
}

func DeActivateUserFromESB(p model.ESBUserStaffPayload) error {
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return errors.New("cannot connect database")
	}
	defer cancel()
	defer conn.Close(ctx)

	orgId := os.Getenv("INTEGRATION_ORG_ID")
	username := os.Getenv("INTEGRATION_USR")
	empId := p.UserCode

	query := `
        UPDATE um_users
        SET active = false,
            "updatedAt" = NOW(),
			"updatedBy" = $3
        WHERE "orgId" = $1 AND "empId" = $2
    `

	cmd, err := conn.Exec(ctx, query, orgId, empId, username)
	if err != nil {
		return fmt.Errorf("update error: %v", err)
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("no user found with orgId=%s empId=%s", orgId, empId)
	}

	return nil
}
