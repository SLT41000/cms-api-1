package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mainPackage/model"
	"mainPackage/utils"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

	log.Printf("Kafka CREATE %s() started. Listening for messages...%v", type_, TOPIC)
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

	log.Printf("Kafka UPDATE %s() started. Listening for messages...%v", type_, TOPIC)
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

	log.Printf("Kafka DELETE %s() started. Listening for messages...%v", type_, os.Getenv("ESB_USER_STAFF_DELETE"))
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

	// Get environment variables
	username := strings.TrimSpace(os.Getenv("INTEGRATION_USR"))
	orgId := strings.TrimSpace(os.Getenv("INTEGRATION_ORG_ID"))

	deptId := strings.TrimSpace(os.Getenv("USER_DEPT_ID"))
	commId := strings.TrimSpace(os.Getenv("USER_COMM_ID"))
	stnId := strings.TrimSpace(os.Getenv("USER_STN_ID"))

	roleId := strings.TrimSpace(os.Getenv("USER_ROLE"))
	if type_ == "ADMIN" {
		roleId = strings.TrimSpace(os.Getenv("ADMIN_ROLE"))
	}

	// Prepare fields
	displayName := strings.TrimSpace(p.FirstName + " " + p.LastName)
	userCode := strings.TrimSpace(p.UserCode)
	userUsername := strings.TrimSpace(p.Username)
	encPassword, _ := encrypt(p.Password)

	fmt.Println("DEBUG PARAM TYPES:")
	fmt.Printf("orgId      (%T): %v\n", orgId, orgId)
	fmt.Printf("roleId     (%T): %v\n", roleId, roleId)
	fmt.Printf("deptId     (%T): %v\n", deptId, deptId)
	fmt.Printf("commId     (%T): %v\n", commId, commId)
	fmt.Printf("stnId      (%T): %v\n", stnId, stnId)
	fmt.Printf("Username   (%T): %v\n", userUsername, userUsername)
	fmt.Printf("UserCode   (%T): %v\n", userCode, userCode)

	// ---------- 1) Try UPDATE first ----------
	res, err := conn.Exec(ctx, `
        UPDATE um_users
        SET
            "displayName" = $1,
            "firstName" = $2,
            "lastName" = $3,
            "mobileNo" = $4,
            "email" = $5, 
            "roleId" = $6::uuid,
            "deptId" = $7::uuid,
            "commId" = $8::uuid,
            "stnId" = $9::uuid,
            "updatedAt" = NOW(),
            "updatedBy" = $10
        WHERE "orgId" = $11::uuid AND ("empId" = $12 OR "username" = $13)
    `,
		displayName, p.FirstName, p.LastName, p.Phone, p.Email,
		roleId, deptId, commId, stnId, username,
		orgId, userCode, userUsername,
	)
	if err != nil {
		return fmt.Errorf("update user error: %w", err)
	}

	// ---------- 2) If no rows updated → INSERT ----------
	if res.RowsAffected() == 0 {
		log.Print("--INSERT--")
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
			displayName,
			userCode,
			p.FirstName,
			p.LastName,
			p.Phone,
			p.Email,
			userUsername,
			encPassword,
			roleId,
			deptId,
			commId,
			stnId,
			username,
			username,
		)
		if err != nil {
			return fmt.Errorf("insert user error: %w", err)
		}

	} else {
		log.Print("--UPDATE--")
	}
	log.Print("--start-CreateUnit")
	err = CreateUnit(ctx, conn, orgId, p, username)
	if err != nil {
		return fmt.Errorf("failed to create unit: %w", err)
	}

	err = CheckArea(ctx, conn, orgId, p, username)
	if err != nil {
		return fmt.Errorf("failed to create unit: %w", err)
	}

	return nil
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

	DeleteUnit(ctx, conn, orgId, empId)
	DeleteUnitProperty(ctx, conn, orgId, empId)

	return nil
}

func CreateUnit(ctx context.Context, conn *pgx.Conn, orgId string, payload model.ESBUserStaffPayload, createdBy string) error {
	log.Print("===CreateUnit===")

	MDMSourceID := os.Getenv("MDM_SOURCE_ID")
	MDMTypeID := os.Getenv("MDM_TYPE_ID")
	MDMCompany := os.Getenv("MDM_COMPANY")
	UserDeptID := os.Getenv("USER_DEPT_ID")
	UserCommID := os.Getenv("USER_COMM_ID")
	UserStnID := os.Getenv("USER_STN_ID")

	unitId := strings.TrimSpace(payload.UserCode)
	unitName := strings.TrimSpace(payload.FirstName + " " + payload.LastName)

	// ---------- 1) Try UPDATE first ----------
	res, err := conn.Exec(ctx, `
		UPDATE mdm_units
		SET
			"unitName" = $1,
			"unitSourceId" = $2,
			"unitTypeId" = $3,
			"priority" = 1,
			"compId" = $4,
			"deptId" = $5,
			"commId" = $6,
			"stnId" = $7, 
			"username" = $8, 
			"updatedAt" = NOW(),
			"updatedBy" = $9
		WHERE "orgId" = $10 AND "unitId" = $11
	`,
		unitName, MDMSourceID, MDMTypeID,
		MDMCompany, UserDeptID, UserCommID, UserStnID,
		payload.Username,
		createdBy,
		orgId, unitId,
	)
	if err != nil {
		return fmt.Errorf("CreateUnit update error: %w", err)
	}

	// ---------- 2) If no rows updated → INSERT ----------
	if res.RowsAffected() == 0 {
		_, err := conn.Exec(ctx, `
			INSERT INTO mdm_units (
				"orgId","unitId","unitName",
				"unitSourceId","unitTypeId","priority",
				"compId","deptId","commId","stnId",
				"active","sttId","username",
				"isLogin","isFreeze","isOutArea",
				"createdAt","updatedAt","createdBy","updatedBy"
			)
			VALUES (
				$1,$2,$3,
				$4,$5,1,
				$6,$7,$8,$9,
				true,'000', $10,
				false,false,false,
				NOW(),NOW(),$11,$11
			)
		`,
			orgId, unitId, unitName,
			MDMSourceID, MDMTypeID,
			MDMCompany, UserDeptID, UserCommID, UserStnID,
			payload.Username,
			createdBy,
		)
		if err != nil {
			return fmt.Errorf("CreateUnit insert error: %w", err)
		}
	}

	err = CreateUnitProperty(ctx, conn, orgId, unitId, createdBy)
	if err != nil {
		return fmt.Errorf("failed to create unit: %w", err)
	}
	return nil
}

func CreateUnitProperty(
	ctx context.Context,
	conn *pgx.Conn,
	orgId, unitId, createdBy string,
) error {
	log.Print("===CreateUnitProperty===")

	propId := os.Getenv("MDM_PROPERTY") // ต้องเป็น uuid แบบถูกต้อง
	if _, err := uuid.Parse(propId); err != nil {
		return fmt.Errorf("invalid propId uuid: %w", err)
	}

	// ----- 1) UPDATE -----
	res, err := conn.Exec(ctx, `
        UPDATE mdm_unit_with_properties
        SET 
            "active" = true,
            "updatedAt" = NOW(),
            "updatedBy" = $1
        WHERE "orgId" = $2 
            AND "unitId" = $3 
            AND "propId" = $4::uuid;
    `, createdBy, orgId, unitId, propId)
	if err != nil {
		return fmt.Errorf("CreateUnitProperty update error: %w", err)
	}

	// ----- 2) INSERT ถ้า update ไม่เจอ -----
	if res.RowsAffected() == 0 {
		_, err := conn.Exec(ctx, `
            INSERT INTO mdm_unit_with_properties (
                "orgId","unitId","propId",
                "active","createdAt","updatedAt",
                "createdBy","updatedBy"
            )
            VALUES (
                $1,$2,$3::uuid,
                true,NOW(),NOW(),
                $4,$4
            );
        `, orgId, unitId, propId, createdBy)
		if err != nil {
			return fmt.Errorf("CreateUnitProperty insert error: %w", err)
		}
	}

	return nil
}

func DeleteUnit(ctx context.Context, conn *pgx.Conn, orgId, username string) error {
	log.Printf("Deleting unit for username=%s in orgId=%s", username, orgId)

	res, err := conn.Exec(ctx, `
		DELETE FROM mdm_units
		WHERE "orgId" = $1 AND "username" = $2
	`, orgId, username)
	if err != nil {
		return fmt.Errorf("DeleteUnit error: %w", err)
	}

	if res.RowsAffected() == 0 {
		log.Printf("No unit found for username=%s", username)
	} else {
		log.Printf("Deleted %d unit(s) for username=%s", res.RowsAffected(), username)
	}

	return nil
}

func DeleteUnitProperty(ctx context.Context, conn *pgx.Conn, orgId, unitId string) error {
	log.Printf("Deleting unit properties for unitId=%s in orgId=%s", unitId, orgId)

	res, err := conn.Exec(ctx, `
		DELETE FROM mdm_unit_with_properties
		WHERE "orgId" = $1 AND "unitId" = $2
	`, orgId, unitId)
	if err != nil {
		return fmt.Errorf("DeleteUnitProperty error: %w", err)
	}

	if res.RowsAffected() == 0 {
		log.Printf("No unit properties found for unitId=%s", unitId)
	} else {
		log.Printf("Deleted %d property(s) for unitId=%s", res.RowsAffected(), unitId)
	}

	return nil
}

func CheckArea(ctx context.Context, conn *pgx.Conn, orgId string, payload model.ESBUserStaffPayload, username string) error {
	log.Printf("CheckArea employee for empId=%s in orgId=%s", payload.UserCode, orgId)
	empId := payload.UserCode
	//empId = "CDC005"
	res, err := callAPI_2(os.Getenv("METTLINK_SERVER")+"/openapi/v1/work_order/employee/info?workspace=bma&user_employee_code="+empId, "GET", nil, nil)
	if err != nil {
		log.Printf("❌ employee : %v", err)
	} else {
		log.Printf("✅ employee : %v", res)
	}

	SaveUserAreas(ctx, conn, orgId, payload, res, username)

	//SaveUserSkills(ctx, conn, orgId, payload, res, username)
	SaveUserSkills(ctx, conn, orgId, payload.UserCode, res, username)

	return nil
}

func SaveUserAreas(ctx context.Context, conn *pgx.Conn, orgId string, payload model.ESBUserStaffPayload, apiResponse string, createdBy string) error {
	log.Print("=====SaveUserAreas====")
	empId := payload.UserCode
	// 1. Parse API response
	var resp struct {
		Status int `json:"status"`
		Data   struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
			Payload struct {
				Namespace []string `json:"namespace"`
			} `json:"payload"`
		} `json:"data"`
	}

	if err := json.Unmarshal([]byte(apiResponse), &resp); err != nil {
		return fmt.Errorf("failed to parse API response: %w", err)
	}

	namespaceFromAPI := resp.Data.Payload.Namespace
	if len(namespaceFromAPI) == 0 {
		log.Println("No namespace found in API response")
		return nil
	}

	// 2. Load all districts for the org
	districts, err := utils.GetCountryProvinceDistrictsOrLoad(ctx, conn, orgId)
	if err != nil {
		return fmt.Errorf("failed to load districts: %w", err)
	}

	// 3. Build a map of namespace -> distId
	nsToDistId := make(map[string]string)
	for _, d := range districts {
		if d.NameSpace != nil && d.DistID != nil {
			nsToDistId[*d.NameSpace] = *d.DistID
		}
	}
	log.Print("---3. Build a map of namespace -> distId")
	log.Print(nsToDistId)
	// 4. Collect all distIds that match the namespaces from API
	var distIds []string
	for _, ns := range namespaceFromAPI {
		if distId, ok := nsToDistId[ns]; ok {
			distIds = append(distIds, distId)
		}
	}

	if len(distIds) == 0 {
		log.Println("No matching distIds found for user:", empId)
		return nil
	}

	// 5. Insert or update into um_user_with_area_response
	distIdsJSON, _ := json.Marshal(distIds)
	log.Print(string(distIdsJSON))
	_, err = conn.Exec(ctx, `
        INSERT INTO um_user_with_area_response
            ("orgId", "username", "distIdLists", "createdAt", "updatedAt", "createdBy", "updatedBy")
        VALUES ($1, $2, $3, NOW(), NOW(), $4, $4)
        ON CONFLICT ("orgId", "username") 
        DO UPDATE SET "distIdLists" = EXCLUDED."distIdLists", "updatedAt" = NOW(), "updatedBy" = $4
    `, orgId, empId, string(distIdsJSON), createdBy)

	log.Print(err)
	if err != nil {
		return fmt.Errorf("failed to insert/update user area: %w", err)
	}

	log.Printf("Saved distIds for user=%s: %v\n", empId, distIds)
	return nil
}

func SaveUserSkills(ctx context.Context, conn *pgx.Conn, orgId, empId, apiResponse, operator string) error {
	log.Print("=====SaveUserSkills====")
	// 1) Load all skills from DB
	allSkills, err := utils.GetUserSkillsOrLoad(ctx, conn, orgId)
	if err != nil {
		log.Printf("allSkills : %v\n", err)
		return err
	}

	// 2) Parse API response
	var parsed model.APISkillResponse
	if err := json.Unmarshal([]byte(apiResponse), &parsed); err != nil {
		log.Printf("parse apiResponse error: %v", err)
		return err
	}

	apiSkillList := parsed.Data.Payload.Skills // []model.APISkill

	// 3) CALL sync function
	err = SyncUserSkills(ctx, conn, orgId, empId, operator, allSkills, apiSkillList)
	if err != nil {
		log.Printf("SyncUserSkills : %v\n", err)
		return err
	}

	return nil
}

func SyncUserSkills(
	ctx context.Context,
	conn *pgx.Conn,
	orgId, username, operator string,
	skillsDB []model.Skill,
	apiSkills []model.APISkill,
) error {

	log.Print("=====SyncUserSkills====")
	now := time.Now()

	// 1) Build DB master: slug -> skillId
	slugToSkillID := make(map[string]string)
	for _, s := range skillsDB {
		slugToSkillID[strings.ToLower(s.En)] = s.SkillID
	}

	// 2) Build expected skillIds from API
	newSkillIDs := map[string]bool{}

	for _, api := range apiSkills {
		skillID, ok := slugToSkillID[strings.ToLower(api.SkillSlug)]
		if !ok {
			log.Printf("⚠️ skillSlug not found: %s", api.SkillSlug)
			continue
		}
		newSkillIDs[skillID] = true
	}

	// 3) Load existing user skillIds in DB
	querySelect := `
        SELECT "skillId"
        FROM public.um_user_with_skills
        WHERE "orgId"=$1 AND "userName"=$2;
    `
	rows, err := conn.Query(ctx, querySelect, orgId, username)
	if err != nil {
		return err
	}

	existingSkillIDs := map[string]bool{}
	for rows.Next() {
		var skillId string
		err := rows.Scan(&skillId)
		if err != nil {
			return err
		}
		existingSkillIDs[skillId] = true
	}

	// 4) Insert new skills
	for skillId := range newSkillIDs {
		if !existingSkillIDs[skillId] {
			_, err := conn.Exec(ctx, `
                INSERT INTO public.um_user_with_skills
                    ("orgId","userName","skillId","active","createdAt","updatedAt","createdBy","updatedBy")
                VALUES ($1,$2,$3,true,$4,$4,$5,$5)
                ON CONFLICT ("orgId","userName","skillId") DO NOTHING;
            `, orgId, username, skillId, now, operator)
			if err != nil {
				return err
			}
		}
	}

	// 5) Remove skills that are NOT in API list
	for skillId := range existingSkillIDs {
		if !newSkillIDs[skillId] {
			_, err := conn.Exec(ctx, `
                DELETE FROM public.um_user_with_skills
                WHERE "orgId"=$1 AND "userName"=$2 AND "skillId"=$3;
            `, orgId, username, skillId)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func LoadUserSkills(ctx context.Context, conn *pgx.Conn, orgId, username string) ([]string, error) {
	query := `SELECT "skillId" 
              FROM public.um_user_with_skills
              WHERE "orgId"=$1 AND "userName"=$2`

	rows, err := conn.Query(ctx, query, orgId, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userSkills []string
	for rows.Next() {
		var skillId string
		rows.Scan(&skillId)
		userSkills = append(userSkills, skillId)
	}
	return userSkills, nil
}
