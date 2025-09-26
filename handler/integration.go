package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mainPackage/config"
	"mainPackage/model"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

func IntegrateCreateCaseFromWorkOrder(ctx context.Context, conn *pgx.Conn, workOrder model.WorkOrder, username, orgId string) error {
	now := time.Now()
	caseId := workOrder.WorkOrderNumber
	var id int
	caseData, err := GetCaseByID(ctx, conn, orgId, caseId)
	if err != nil {
		log.Printf("Error getting case: %v", err)
		return nil
	}

	if caseData != nil {
		log.Printf("Case Duplicate")
		return nil
	}

	query := `
	INSERT INTO public."tix_cases"(
	"orgId", "caseId", "caseVersion" , "caseTypeId", "caseSTypeId", priority, "wfId", "versions",source, "deviceId",
	"caseDetail", "statusId", "caseLat", "caseLon", "countryId", "provId", "distId", "caseDuration", "createdDate", "startedDate",
	  usercreate,  "createdAt", "updatedAt", "createdBy", "updatedBy", "integration_ref_number", "caseSla", "phoneNoHide", "deviceMetaData")
	VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
		$11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
		$21, $22, $23, $24, $25, $26, $27, $28, $29
	) RETURNING id ;
	`

	//Mapping get Subtype and Workflow SOP
	IntegrationRefNumber := workOrder.WorkOrderRefNumber
	DeviceType := workOrder.DeviceMetadata.DeviceType
	WorlkOrderType := workOrder.WorkOrderType
	sType, err := GetSubTypeByID(ctx, conn, orgId, DeviceType, WorlkOrderType)
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
	area, err := GetAreaByNamespace(ctx, conn, orgId, workOrder.Namespace)
	if err != nil {
		log.Fatalf("lookup failed: %v", err)
	}
	if area == nil {
		log.Println("area not found")
	} else {
		fmt.Printf("Country: %s, Province: %s, District: %s\n", area.CountryID, area.ProvID, area.DistID)
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
	err = conn.QueryRow(ctx, query,
		orgId, caseId, "publish", caseTypeId, caseSTypeId, Priority, wfId, wfVersion,
		source, workOrder.DeviceMetadata.DeviceID, workOrder.WorkOrderMetadata.Description, statusId,
		lat, lon, countryId, provId, distId, // countryId, provId, distId (อาจ map จาก device location)
		caseDuration, now, now, username, now, now, username, username, IntegrationRefNumber, caseSla, true, string(deviceJSON)).Scan(&id)

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

	// TODO: upsert device info

	// TODO: send notification
	statuses, err := GetCaseStatusList(&gin.Context{}, conn, orgId)
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
	err_ = genNotiCustom(ctx, conn, orgId, "System", Provider, "", *statusName.Th, data_, msg_alert, recipients, "", "User")

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

	return nil
}

func IntegrateCaseCurrentStageInsert(username string, orgId string, conn *pgx.Conn, ctx context.Context, req model.CustomCaseCurrentStage) error {
	logger := config.GetLog()

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
		"caseId", "integration_ref_number"
	FROM public."tix_cases"
	WHERE "orgId" = $1 AND "caseId" = $2
	LIMIT 1;
	`

	var c model.Case
	var deviceMetaJSON []byte

	err := conn.QueryRow(ctx, query, orgId, caseId).Scan(
		&c.CaseID,
		&c.IntegrationRefNumber,
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

func IntegrateUpdateCaseFromWorkOrder(ctx context.Context, conn *pgx.Conn, workOrder model.WorkOrder, username, orgId string) error {
	now := time.Now()
	caseId := workOrder.WorkOrderNumber

	// หา case เดิมก่อน
	caseData, err := GetCaseByID(ctx, conn, orgId, caseId)
	if err != nil {
		return fmt.Errorf("error getting case: %w", err)
	}
	if caseData == nil {
		return fmt.Errorf("case not found for update: %s", caseId)
	}

	// Mapping Subtype / Workflow
	//DeviceType := workOrder.DeviceMetadata.DeviceType
	//WorkOrderType := workOrder.WorkOrderType
	// sType, err := GetSubTypeByID(ctx, conn, orgId, DeviceType, WorkOrderType)
	// if err != nil {
	// 	return fmt.Errorf("error getting subtype: %w", err)
	// }
	// if sType == nil {
	// 	return fmt.Errorf("failed to get subtype for DeviceType: %s, WorkOrderType: %s", DeviceType, WorkOrderType)
	// }

	// caseTypeId := sType.TypeID
	// caseSTypeId := sType.STypeID
	// wfId := sType.WFID
	// wfVersion := sType.WfVersions
	// caseSla := sType.CaseSLA
	// wfNodeId := sType.WfNodeId
	// source := "05"
	// statusId := "S001"

	statusMap := GetCaseStatusMap()
	statusId := statusMap[workOrder.Status]

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
	"deviceMetaData" = $9
WHERE 
	"orgId" = $10 AND 
	"caseId" = $11;
`

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
	)

	if err != nil {
		return fmt.Errorf("update case failed: %w", err)
	}

	// === Stage update ===
	var data = model.UpdateStageRequest{
		CaseId:   caseId,
		Status:   statusId,
		UnitId:   workOrder.UserMetadata.AssignedEmployeeCode,
		UnitUser: workOrder.UserMetadata.AssignedEmployeeCode,
		NodeId:   "",
	}

	results, err := UpdateCurrentStageCore(&gin.Context{}, conn, data)
	if err != nil {
		return fmt.Errorf("UpdateCurrentStageCore failed: %w", err)
	}
	log.Print(results)

	return nil
}
