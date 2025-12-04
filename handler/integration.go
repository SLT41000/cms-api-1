package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mainPackage/model"
	"mainPackage/utils"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

func IntegrateCreateCaseFromWorkOrder(ctx *gin.Context, conn *pgx.Conn, workOrder model.WorkOrder, username, orgId string) error {
	log.Printf("====IntegrateCreateCaseFromWorkOrder===")
	loc, _ := time.LoadLocation("Asia/Bangkok")
	now := time.Now().In(loc)
	caseId := workOrder.WorkOrderNumber
	var id int
	caseData, err := GetCaseByID(ctx, conn, orgId, caseId)
	if err != nil {
		log.Printf("Error getting case: %v", err)
		return nil
	}

	if caseData != nil {
		log.Printf("Case Duplicate :  %s", string(caseData.CaseID))
		return nil
	}

	// default = CURRENT
	useCurrent := true
	var scheduleTime time.Time
	scheduleFlag := false
	wd := strings.TrimSpace(workOrder.WorkDate)
	if wd != "" {
		// parse work_date at 00:00:00
		t, err := time.ParseInLocation("2006-01-02", wd, loc)
		if err == nil {
			// ถ้า work_date > ปัจจุบัน → ใช้ t
			if t.After(now) {
				scheduleTime = t
				useCurrent = false
				scheduleFlag = true
			}
		}
	}

	if useCurrent {
		scheduleTime = now
	}

	// convert ไป UTC (ดีที่สุดสำหรับเก็บลง DB)
	scheduleUTC := scheduleTime.UTC()

	query := `
		INSERT INTO public."tix_cases"(
		"orgId", "caseId", "caseVersion", "caseTypeId", "caseSTypeId", priority, "wfId", "versions", source, "deviceId",
		"caseDetail", "statusId", "caseLat", "caseLon", "countryId", "provId", "distId", "caseDuration", "createdDate", "startedDate",
		usercreate, "createdAt", "updatedAt", "createdBy", "updatedBy", "integration_ref_number", "caseSla", "phoneNoHide", "deviceMetaData", "scheduleFlag"
	)
	VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
		$11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
		$21, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, $22, $23, $24, $25, $26, $27, $28
	)
	RETURNING id;
	`

	//Mapping get Subtype and Workflow SOP
	IntegrationRefNumber := workOrder.WorkOrderRefNumber
	DeviceType := workOrder.DeviceMetadata.DeviceType
	WorlkOrderType := workOrder.WorkOrderType
	sType, err := utils.GetSubTypeByID(ctx, conn, orgId, DeviceType, WorlkOrderType)
	if err != nil {
		log.Printf("Error: %v", err)
	}
	if sType == nil {
		return fmt.Errorf("failed to get subtype for DeviceType: %s, WorkOrderType: %s", DeviceType, WorlkOrderType)
	}

	log.Print("=====sType==")
	log.Print(sType)

	caseTypeId := sType.TypeID
	caseSTypeId := sType.STypeID
	wfId := sType.WFID
	wfVersion := sType.WfVersions
	caseSla := sType.CaseSLA
	wfNodeId := sType.WfNodeId
	source := "05"
	statusId := "S001"

	//Priority = fixed หรือ derive จาก WorkOrder ถ้ามี field severity
	//CRITICAL, HIGH, MEDIUM, LOW
	// Priority := 10
	// Severity := workOrder.WorkOrderMetadata.Severity
	// switch Severity {
	// case "CRITICAL":
	// 	Priority = 1
	// case "HIGH":
	// 	Priority = 3
	// case "MEDIUM":
	// 	Priority = 6
	// case "LOW":
	// 	Priority = 9
	// }

	priorityMap := GetPriorityMap()
	Priority := priorityMap[strings.ToUpper(workOrder.WorkOrderMetadata.Severity)]

	// ถ้า severity ไม่ตรง จะ fallback เป็น default (เช่น LOW)
	if Priority == 0 {
		Priority = priorityMap["LOW"]
	}

	// Convert struct to JSON
	deviceJSON, err := json.Marshal(workOrder.DeviceMetadata)
	if err != nil {
		log.Fatal("Failed to marshal device:", err)
	}

	//======== Waiting Mapping Master data
	area, err := utils.GetAreaByNamespace(ctx, conn, orgId, workOrder.Namespace)
	if err != nil {
		log.Fatalf("lookup failed: %v", err)
	}
	if area == nil {
		log.Println("area not found")
	} else {
		fmt.Printf("Country: %v, Province: %v, District: %v\n", area.CountryID, area.ProvID, area.DistID)
	}
	countryId := area.CountryID
	provId := area.ProvID
	distId := area.DistID

	caseDuration := 0

	lat := workOrder.WorkOrderMetadata.Location.Latitude
	if lat == "" {
		lat = "0"
	}
	lon := workOrder.WorkOrderMetadata.Location.Longitude
	if lon == "" {
		lon = "0"
	}
	//workOrder.WorkOrderType
	createBy := workOrder.CreatedBy
	if workOrder.CreatedBy == "" {
		createBy = username
	}
	err = conn.QueryRow(ctx, query,
		orgId, caseId, "publish", caseTypeId, caseSTypeId, Priority, wfId, wfVersion,
		source, workOrder.DeviceMetadata.DeviceID, workOrder.WorkOrderMetadata.Description, statusId,
		lat, lon, countryId, provId, distId, caseDuration,
		scheduleUTC, scheduleUTC, // createdDate, startedDate
		username, createBy, username, IntegrationRefNumber, caseSla, true, string(deviceJSON), scheduleFlag,
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("insert case failed: %w", err)
	}

	var data = model.CustomCaseCurrentStage{
		CaseID:   caseId,
		WfID:     &wfId,
		NodeID:   wfNodeId,
		StatusID: statusId,
	}

	// TODO: insert stage if needed
	Provider := username
	err_ := IntegrateCaseCurrentStageInsert(Provider, orgId, conn, ctx, data)
	if err_ != nil {
		log.Fatalf("IntegrateCaseCurrentStageInsert : %v", err_)
	}

	// X. Insert responder
	_, err = conn.Exec(ctx, `
	    INSERT INTO tix_case_responders ("orgId","caseId","unitId","userOwner","statusId","createdAt","createdBy")
	    VALUES ($1,$2,$3,$4,$5,$6,$7)
	`, orgId, caseId, "case", username, statusId, scheduleUTC, username)
	if err != nil {
		return err
	}

	// === Attachments from WorkOrder images ===
	for _, imgUrl := range workOrder.WorkOrderMetadata.Images {
		if strings.TrimSpace(imgUrl) == "" {
			continue
		}

		// extract name
		parts := strings.Split(imgUrl, "/")
		attName := parts[len(parts)-1]

		err := InsertCaseAttachment(ctx, conn,
			orgId, caseId, "case", attName, imgUrl, username)

		if err != nil {
			log.Printf("InsertCaseAttachment failed: %v", err)
		}
	}

	// TODO: upsert device info

	if workOrder.UserMetadata.AssignedEmployeeCode.UserEmployeeCode != "" {
		unit := workOrder.UserMetadata.AssignedEmployeeCode.UserEmployeeCode
		// === Stage update ===
		var data = model.UpdateStageRequest{
			CaseId:    caseId,
			Status:    "S003",
			UnitId:    unit,
			UnitUser:  unit,
			NodeId:    "",
			ResID:     "",
			ResDetail: "",
		}

		log.Print("====data1===")
		b, _ := json.MarshalIndent(data, "", "  ")
		fmt.Println(string(b))

		log.Print("====8===")

		_, _, _, dispatchNode, err := GetWorkflowAndCurrentNode(ctx, orgId, caseId, "")
		if err != nil {
			response := model.Response{
				Status: "-1",
				Msg:    "Failed",
				Desc:   err.Error(),
			}
			ctx.JSON(http.StatusInternalServerError, response)

		}

		log.Print("====dispatch===")
		b, _ = json.MarshalIndent(dispatchNode, "", "  ")
		fmt.Println(string(b))
		data.NodeId = dispatchNode.NodeId

		log.Print("====data2===")
		b, _ = json.MarshalIndent(data, "", "  ")
		fmt.Println(string(b))

		log.Print("====Dispatch on Create===")
		results, err := UpdateCurrentStageCore(ctx, conn, data, false)
		if err != nil {
			return fmt.Errorf("UpdateCurrentStageCore failed: %w", err)
		}
		log.Print(results)

	}

	// TODO: send notification
	statuses, err := utils.GetCaseStatusList(ctx, conn, orgId)
	if err != nil {

	}
	statusMap := make(map[string]model.CaseStatus)
	for _, s := range statuses {
		statusMap[*s.StatusID] = s
	}
	statusName := statusMap["S001"]

	msg := *statusName.Th

	msg_alert := msg + " :: " + caseId

	data_ := []model.Data{
		{Key: "delay", Value: "0"}, //0=white, 1=yellow , 2=red
	}
	recipients := []model.Recipient{
		{Type: "provId", Value: *provId},
	}
	err_ = genNotiCustom(ctx, conn, orgId, "System", Provider, "", *statusName.Th, data_, msg_alert, recipients, "", "User", "")

	//err_ = genNotiCustom(ctx, conn, orgId, "System", Provider, "", "Create", dataNoti, msg+" : "+caseId, recipients, "", "User")
	if err_ != nil {
		log.Fatalf("genNotiCustom : %v", err_)
	}
	//Add Comment
	evt := model.CaseHistoryEvent{
		OrgID:     orgId,
		CaseID:    caseId,
		Username:  Provider,
		Type:      "event",
		FullMsg:   Provider + " :: " + msg + " :: " + caseId,
		JsonData:  "",
		CreatedBy: Provider,
	}
	err = InsertCaseHistoryEvent(ctx, conn, evt)
	if err != nil {
		log.Fatalf("Insert failed: %v", err)
	}

	// For Dashboard
	err = CalDashboardCaseSummary(ctx, conn, orgId, recipients, username, caseTypeId, *countryId, *provId, *distId)
	if err != nil {
		log.Print("AddOrUpdateCaseSummary failed", zap.Error(err))
	}

	return nil
}

func IntegrateCaseCurrentStageInsert(username string, orgId string, conn *pgx.Conn, ctx context.Context, req model.CustomCaseCurrentStage) error {
	logger := utils.GetLog()

	now := time.Now()

	// // 1. Insert responder
	// _, err := conn.Exec(ctx, `
	//     INSERT INTO tix_case_responders ("orgId","caseId","unitId","userOwner","statusId","createdAt","createdBy")
	//     VALUES ($1,$2,$3,$4,$5,NOW(),$6)
	// `, orgId, req.CaseID, "case", username, req.StatusID, username)
	// if err != nil {
	// 	return err
	// }

	// Step 2: Load workflow node from DB
	query := `
	SELECT t1.id, t1."orgId", t1."wfId", t1."nodeId", t1.versions, t1.type, t1.section, t1.data,
	       t1.pic, t1."group", t1."formId", t1."createdAt", t1."updatedAt", t1."createdBy", t1."updatedBy"
	FROM public.wf_nodes t1
	JOIN public.wf_definitions t2
	  ON t1."versions" = t2."versions" AND t1."wfId" = t2."wfId"
	WHERE t2."wfId" = $1 AND t1."nodeId" = $2 AND t2."orgId" = $3
	`

	logger.Debug("Loading workflow node",
		zap.String("query", query),
		zap.Any("params", []any{req.WfID, req.NodeID, orgId}),
	)

	var workflow model.WfNode
	err := conn.QueryRow(ctx, query, req.WfID, req.NodeID, orgId).Scan(
		&workflow.ID, &workflow.OrgID, &workflow.WfID, &workflow.NodeID,
		&workflow.Versions, &workflow.Type, &workflow.Section,
		&workflow.Data, &workflow.Pic, &workflow.Group, &workflow.FormID,
		&workflow.CreatedAt, &workflow.UpdatedAt, &workflow.CreatedBy, &workflow.UpdatedBy,
	)

	log.Print("===workflow")
	log.Print(workflow)

	if err != nil {
		if err == pgx.ErrNoRows {
			logger.Warn("No workflow node found")
			return fmt.Errorf("workflow node not found")
		}
		logger.Error("Failed to load workflow node", zap.Error(err))
		return err
	}

	// Step 3: Insert into tix_case_current_stage
	insertQuery := `
	INSERT INTO public.tix_case_current_stage(
		"orgId", "caseId", "wfId", "nodeId", "stageType", "unitId", "username", versions, type, section, data, pic, "group", "formId",
		"createdAt", "updatedAt", "createdBy", "updatedBy"
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11,
		$12, $13, $14, $15, $16, $17, $18
	)
	`

	args := []interface{}{
		workflow.OrgID, req.CaseID, workflow.WfID, req.NodeID, "case", "", "", workflow.Versions,
		workflow.Type, workflow.Section, workflow.Data, workflow.Pic,
		workflow.Group, workflow.FormID, now, now, username, username,
	}

	logger.Debug("Inserting current stage",
		zap.String("query", insertQuery),
		zap.Any("args", args),
	)

	_, err = conn.Exec(ctx, insertQuery, args...)
	if err != nil {
		logger.Error("Insert failed", zap.Error(err))
		return err
	}

	logger.Info("Insert success", zap.String("caseId", req.CaseID))
	return nil
}

func GetCaseByID(ctx context.Context, conn *pgx.Conn, orgId string, caseId string) (*model.Case, error) {
	query := `
	SELECT 
		"caseId", "integration_ref_number", "distId", "statusId", "caseTypeId", "caseSTypeId", "priority", "caseLat", "caseLon", "caseDetail", "deviceMetaData", "wfId",
		"countryId", "provId", "distId", "createdDate", "scheduleFlag", "scheduleDate"
	FROM public."tix_cases"
	WHERE "orgId" = $1 AND "caseId" = $2
	LIMIT 1;
	`

	var c model.Case
	var deviceMetaJSON []byte

	err := conn.QueryRow(ctx, query, orgId, caseId).Scan(
		&c.CaseID,
		&c.IntegrationRefNumber,
		&c.DistID,
		&c.StatusID,
		&c.CaseTypeID,
		&c.CaseSTypeID,
		&c.Priority,
		&c.CaseLat,
		&c.CaseLon,
		&c.CaseDetail,
		&c.DeviceMetaData,
		&c.WfID,
		&c.CountryID,
		&c.ProvID,
		&c.DistID,
		&c.CreatedDate,
		&c.ScheduleFlag,
		&c.ScheduleDate,
	)

	// ✅ case not found
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query case failed: %w", err)
	}

	// ✅ handle JSON column (TEXT / JSONB)
	if len(deviceMetaJSON) > 0 {
		if err := json.Unmarshal(deviceMetaJSON, &c.DeviceMetaData); err != nil {
			log.Printf("Warning: cannot unmarshal deviceMetaData: %v", err)
		}
	}

	return &c, nil
}

func IntegrateUpdateCaseFromWorkOrder(ctx *gin.Context, conn *pgx.Conn, workOrder model.WorkOrder, username, orgId string) error {
	now := time.Now()
	caseId := workOrder.WorkOrderNumber
	log.Print("====3===")
	// หา case เดิมก่อน
	caseData, err := GetCaseByID(ctx, conn, orgId, caseId)
	if err != nil {
		return fmt.Errorf("error getting case: %w", err)
	}
	if caseData == nil {
		return fmt.Errorf("case not found for update: %s", caseId)
	}
	log.Print("====4===")
	statusMap := GetCaseStatusMap()
	statusId := statusMap[workOrder.Status]
	if workOrder.State == "CLOSED" {
		statusId = statusMap[workOrder.State]
	}
	// Priority จาก ENV mapping
	priorityMap := GetPriorityMap()
	Priority := priorityMap[strings.ToUpper(workOrder.WorkOrderMetadata.Severity)]
	if Priority == 0 {
		Priority = priorityMap["LOW"]
	}

	// Convert struct to JSON
	deviceJSON, err := json.Marshal(workOrder.DeviceMetadata)
	if err != nil {
		return fmt.Errorf("failed to marshal device: %w", err)
	}
	log.Print("====5===")
	// Master data mapping
	//countryId := "TH"
	//provId := "10"
	//distId := "101"

	// lat := workOrder.WorkOrderMetadata.Location.Latitude
	// if lat == "" {
	// 	lat = "0"
	// }
	// lon := workOrder.WorkOrderMetadata.Location.Longitude
	// if lon == "" {
	// 	lon = "0"
	// }

	// === UPDATED SQL QUERY ===
	query := `
UPDATE public."tix_cases"
SET 
	"caseVersion" = $1,
	priority = $2,
	"deviceId" = $3,
	"caseDetail" = $4,
	"statusId" = $5, 
	"updatedAt" = $6,
	"updatedBy" = $7,
	"integration_ref_number" = $8, 
	"deviceMetaData" = $9,
	"overSlaFlag" = $12,
	"overSlaDate" = $13,
	"overSlaCount" = 0
WHERE 
	"orgId" = $10 AND 
	"caseId" = $11;
`
	log.Print("====6===")
	_, err = conn.Exec(ctx, query,
		"publish",                               // $1  caseVersion
		Priority,                                // $2  priority
		workOrder.DeviceMetadata.DeviceID,       // $3  deviceId
		workOrder.WorkOrderMetadata.Description, // $4  caseDetail
		statusId,                                // $5  statusId
		now,                                     // $6  updatedAt
		username,                                // $7  updatedBy
		workOrder.WorkOrderRefNumber,            // $8  integration_ref_number
		string(deviceJSON),                      // $9  deviceMetaData
		orgId,                                   // $10 orgId
		caseId,                                  // $11 caseId
		false,
		nil,
	)

	if err != nil {
		return fmt.Errorf("update case failed: %w", err)
	}
	log.Print("====7===")
	// === Stage update ===
	var data = model.UpdateStageRequest{
		CaseId:    caseId,
		Status:    statusId,
		UnitId:    workOrder.UserMetadata.AssignedEmployeeCode.UserEmployeeCode,
		UnitUser:  workOrder.UserMetadata.AssignedEmployeeCode.UserEmployeeCode,
		NodeId:    "",
		ResID:     "",
		ResDetail: "",
	}
	log.Print("====data1===")
	b, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(b))

	log.Print("====8===")

	if strings.TrimSpace(data.NodeId) == "" {
		log.Print("====8.11111===")
		_, _, _, dispatchNode, err := GetWorkflowAndCurrentNode(ctx, orgId, caseId, "")
		if err != nil {
			response := model.Response{
				Status: "-1",
				Msg:    "Failed",
				Desc:   err.Error(),
			}
			ctx.JSON(http.StatusInternalServerError, response)

		}

		log.Print("====dispatch===")
		b, _ = json.MarshalIndent(dispatchNode, "", "  ")
		fmt.Println(string(b))
		data.NodeId = dispatchNode.NodeId
	}

	log.Print("====data2===")
	b, _ = json.MarshalIndent(data, "", "  ")
	fmt.Println(string(b))

	results, err := UpdateCurrentStageCore(ctx, conn, data, false)
	if err != nil {
		return fmt.Errorf("UpdateCurrentStageCore failed: %w", err)
	}
	log.Print(results)

	return nil
}

func CreateBusKafka_WO(ctx *gin.Context, conn *pgx.Conn, req model.CaseInsert, sType *model.CaseSubType, integration_ref_number string, source string, username string) error {
	log.Print("=====CreateBusKafka_WO===")

	if os.Getenv("INTEGRATION_SOURCE") == source {
		log.Printf("Skip Original Source : %s\n", source)
	}
	//username := GetVariableFromToken(ctx, "username")
	orgId := GetVariableFromToken(ctx, "orgId")
	log.Print("=====orgId===", orgId)
	log.Print("=====CaseSTypeID===", req.CaseSTypeID)
	// sType, err := GetCaseSubTypeByCode(ctx, conn, orgId.(string), req.CaseSTypeID)
	// if err != nil {
	// 	log.Printf("sType Error: %v", err)
	// }
	// if sType == nil {
	// 	return fmt.Errorf("failed for CaseSTypeID: %s", req.CaseSTypeID)
	// }
	log.Print("=====sType===", sType.TH)
	log.Print("=====DistID===", req.DistID)
	areaDist, err := utils.GetAreaById(ctx, conn, orgId.(string), req.DistID)
	if err != nil {
		log.Printf("areaDist Error: %v", err)
	}
	if areaDist == nil {
		return fmt.Errorf("failed for areaDist: %s", req.DistID)
	}
	log.Print("=====areaDist===", areaDist.Th)
	currentDate := time.Now().Format("2006-01-02")
	if *req.ScheduleFlag {
		log.Print("=====ScheduleDate===", req.ScheduleDate.String())
		currentDate = ConvertDateSafe(req.ScheduleDate.String())
		if req.StatusID == os.Getenv("SCHEDULE") {
			req.StatusID = os.Getenv("NEW")
		}
	}

	num, err := strconv.Atoi(sType.Priority)
	if err != nil {
		fmt.Println("Error:", err)
	}

	attachments, err := GetCaseAttachments_(ctx, conn, orgId.(string), *req.CaseId)
	if err != nil {

	}

	data := map[string]interface{}{
		"work_order_number":     req.CaseId,
		"work_order_ref_number": integration_ref_number,
		"work_order_type":       sType.MWorkOrderType,
		"work_order_metadata": map[string]interface{}{
			"title":       sType.TH,
			"description": req.CaseDetail,
			"severity":    GetPriorityName_TXT(num), // CRITICAL, HIGH, MEDIUM, LOW
			"location": map[string]interface{}{
				"latitude":  req.CaseLat,
				"longitude": req.CaseLon,
			},
			"images": attachments,
		},
		"user_metadata": map[string]interface{}{
			"assigned_employee_code":  "",
			"associate_employee_code": []string{},
		},
		"device_metadata": map[string]interface{}{}, // ตอนนี้ว่าง
		"sop_metadata":    map[string]interface{}{},
		"state":           "OPEN",
		"status":          "NEW",
		"work_date":       currentDate,
		"workspace":       os.Getenv("INTEGRATION_WORKSPACE"),
		"namespace":       *areaDist.NameSpace,
		"source":          os.Getenv("INTEGRATION_SOURCE"),
		"created_by":      username,
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
	}

	jsonStr := string(jsonBytes)
	fmt.Println(jsonStr)
	log.Print("===END===CreateBusKafka_WO=")
	res, err := callAPI(os.Getenv("METTTER_SERVER")+"/mettriq/v1/work_order/create", "POST", data)
	if err != nil {
		return err
	}

	log.Print(res)

	return nil
}

func UpdateBusKafka_WO(ctx *gin.Context, conn *pgx.Conn, req model.UpdateStageRequest) error {
	log.Print("=====UpdateBusKafka_WO===")
	currentDate := time.Now().Format("2006-01-02")
	orgId := GetVariableFromToken(ctx, "orgId")

	caseData, err := GetCaseByID(ctx, conn, orgId.(string), req.CaseId)
	if err != nil {
		log.Printf("Error getting case: %v", err)
	}

	areaDist, err := utils.GetAreaById(ctx, conn, orgId.(string), caseData.DistID)
	if err != nil {
		log.Printf("areaDist Error: %v", err)
	}
	if areaDist == nil {
		return fmt.Errorf("failed for areaDist: %s", caseData.DistID)
	}
	log.Print("=====areaDist===", areaDist.Th)

	stName := mapStatus(caseData.StatusID)
	//---> REF Number
	//---> user profile
	// --- User assignment info
	var state = "OPEN"
	if stName == "CLOSED" {
		state = stName
		stName = "DONE"
	}

	var uAssign interface{} = "" // default empty string
	if req.UnitUser != "" {
		user, err := utils.GetUserByUsername(ctx, conn, orgId.(string), req.UnitUser)
		if err != nil {
			log.Printf("Error getting user: %v", err)
		} else if user != nil {
			uAssign = map[string]interface{}{
				"user_employee_code": user.EmpID,
				"user_firstname":     user.FirstName,
				"user_lastname":      user.LastName,
				"user_avatar":        user.Photo,
				"user_phone":         user.MobileNo,
			}
		}
	}

	if stName == "CANCEL" {
		state = "CLOSED"
		uAssign = ""
		assign_ := os.Getenv("ASSIGNED")
		unitLists, count, err := GetUnits(ctx, conn, orgId.(string), req.CaseId, assign_, "")
		if err != nil {
			return fmt.Errorf("get unitLists failed: %w", err)
		}
		if count > 0 {
			firstUnitId := unitLists[0].UnitID
			user, err := utils.GetUserByUsername(ctx, conn, orgId.(string), firstUnitId)
			if err != nil {
				log.Printf("Error getting user: %v", err)
			} else if user != nil {
				uAssign = map[string]interface{}{
					"user_employee_code": user.EmpID,
					"user_firstname":     user.FirstName,
					"user_lastname":      user.LastName,
					"user_avatar":        user.Photo,
					"user_phone":         user.MobileNo,
				}
			}
		}

	}

	sType, err := utils.GetCaseSubTypeByCode(ctx, conn, orgId.(string), caseData.CaseSTypeID)
	if err != nil {
		log.Printf("sType Error: %v", err)
	}

	attachments, err := GetCaseAttachments_(ctx, conn, orgId.(string), req.CaseId)
	if err != nil {

	}
	data := map[string]interface{}{
		"work_order_number":     req.CaseId,
		"work_order_ref_number": caseData.IntegrationRefNumber,
		"work_order_metadata": map[string]interface{}{
			"title":       sType.TH,
			"description": caseData.CaseDetail,
			"severity":    GetPriorityName_TXT(caseData.Priority), // CRITICAL, HIGH, MEDIUM, LOW
			"location": map[string]interface{}{
				"latitude":  caseData.CaseLat,
				"longitude": caseData.CaseLon,
			},
			"images": attachments,
		},
		"user_metadata": map[string]interface{}{
			"assigned_employee_code":  uAssign,
			"associate_employee_code": []string{},
		},
		"device_metadata": caseData.DeviceMetaData,
		//"sop_metadata": caseData.DeviceMetaData,
		"status":    stName, //NEW, ASSIGNED, ACKNOWLEDGE, INPROGRESS, DONE, ONHOLD, CANCEL
		"state":     state,
		"work_date": currentDate,
		"workspace": os.Getenv("INTEGRATION_WORKSPACE"),
		"namespace": *areaDist.NameSpace,
		"source":    os.Getenv("INTEGRATION_SOURCE"),
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
	}

	jsonStr := string(jsonBytes)
	fmt.Println(jsonStr)
	res, err := callAPI(os.Getenv("METTTER_SERVER")+"/mettriq/v1/work_order/update", "POST", data)
	if err != nil {
		return err
	}
	log.Print("/mettriq/v1/work_order/update")
	log.Print(data)
	log.Print(res)

	return nil
}

func InsertCaseAttachment(ctx context.Context, conn *pgx.Conn,
	orgId, caseId, attachmentType, attName, attUrl, createdBy string) error {

	query := `
		INSERT INTO public.tix_case_attachments
		("orgId","caseId","type","attId","attName","attUrl","createdAt","updatedAt","createdBy","updatedBy")
		VALUES ($1,$2,$3,$4,$5,$6,now(),now(),$7,$8)
	`

	attId := uuid.New().String()

	_, err := conn.Exec(ctx, query,
		orgId, caseId, attachmentType, attId, attName, attUrl, createdBy, createdBy)

	return err
}
