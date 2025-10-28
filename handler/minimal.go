package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mainPackage/model"
	"mainPackage/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// @summary Trigger Create Case
// @id Trigger Create Case
// @tags Minimal API Integration
// @accept json
// @produce json
// @param Body body model.MinimalCaseInsert true "Create Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/minimal/case/create [post]
func MinimalCreateCase(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	var req model.MinimalCaseInsert
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Insert failed", zap.Error(err))
		return
	}
	username := "apiwat.rod"
	orgId := "434c0f16-b7ea-4a7b-a74b-e2e0f859f549"
	txtId := uuid.New().String()
	now := time.Now()
	caseId := req.CaseId
	var id int

	query := `
	INSERT INTO public."tix_cases"(
	"orgId", "caseId", "caseVersion" , "caseTypeId", "caseSTypeId", priority, "wfId", "versions",source, "deviceId",
	"phoneNo", "phoneNoHide", "caseDetail", "statusId", "caseLat", "caseLon", "countryId", "provId", "distId", "caseDuration", "createdDate", "startedDate",
	  usercreate,  "createdAt", "updatedAt", "createdBy", "updatedBy")
	VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
		$11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
		$21, $22, $23, $24, $25, $26, $27
	) RETURNING id ;
	`

	//Priority = Get from subtype
	Priority := 0
	caseDuration := 0

	lat := req.IotInfo.Latitude
	if lat != nil && *lat == "" {
		lat = nil
	}
	lon := req.IotInfo.Longitude
	if lon != nil && *lon == "" {
		lon = nil
	}

	logger.Debug(`Query`, zap.String("query", query), zap.Any("req", req))
	err := conn.QueryRow(ctx, query,
		orgId, caseId, "publish", req.CaseTypeID, req.CaseSTypeID, Priority, req.WfID, "draft",
		req.Source, req.IotInfo.DeviceID, req.PhoneNo, true, req.CaseDetail, req.StatusID,
		lat, lon, req.IotInfo.CountryID, req.IotInfo.ProvID, req.IotInfo.DistID,
		caseDuration, now, now, username, now, now, username, username).Scan(&id)

	if err != nil {
		// log.Printf("Insert failed: %v", err)
		response := model.Response{
			Status: "-1",
			Msg:    "Failure.1",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId, username,
			txtId, "", "Cases", "MinimalCreateCase", "",
			"search", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		logger.Warn("Insert failed", zap.Error(err))
		return
	}
	fmt.Printf("=======xxxx========")
	if req.NodeID != "" {
		var data = model.CustomCaseCurrentStage{
			CaseID:   caseId,
			WfID:     &req.WfID,
			NodeID:   req.NodeID,
			StatusID: req.StatusID,
		}
		fmt.Printf("=======yyy========")
		err = MinCaseCurrentStageInsert(username, orgId, conn, ctx, c, data)
		if err != nil {
			response := model.Response{
				Status: "-1",
				Msg:    "Failure.2",
				Desc:   err.Error(),
			}
			//=======AUDIT_START=====//
			_ = utils.InsertAuditLogs(
				c, conn, orgId, username,
				txtId, "", "Cases", "MinimalCreateCase", "",
				"search", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
			)
			//=======AUDIT_END=====//
			c.JSON(http.StatusInternalServerError, response)
			return
		}
	}

	if req.IotInfo != nil {
		err := UpsertDeviceIot(conn, ctx, orgId, req.IotInfo, username)
		if err != nil {
			logger.Error("Failed to upsert device_iot", zap.Error(err))
			// handle error
		}
	}

	//Noti Custom
	data := []model.Data{
		{Key: "Create", Value: "2"},
	}

	recipients := []model.Recipient{
		{Type: "provId", Value: req.IotInfo.ProvID},
	}
	event := "CASE-CREATE"
	additionalJsonMap := map[string]interface{}{
		"caseId": req.CaseId,
	}
	additionalJSON, err := json.Marshal(additionalJsonMap)
	if err != nil {
		log.Printf("covent additionalData Error :", err)
	}
	additionalData := json.RawMessage(additionalJSON)
	genNotiCustom(c, conn, orgId, "System", "MEETRIQ", "", "Create", data, "เปิด Case สำเร็จ : "+caseId, recipients, "", "User", event, &additionalData)

	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Create successfully",
	}
	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId, username,
		txtId, "", "Cases", "MinimalCreateCase", "",
		"search", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)

}

func MinCaseCurrentStageInsert(username string, orgId string, conn *pgx.Conn, ctx context.Context, c *gin.Context, req model.CustomCaseCurrentStage) error {
	logger := utils.GetLog()

	now := time.Now()

	// 1. Insert responder
	_, err := conn.Exec(ctx, `
        INSERT INTO tix_case_responders ("orgId","caseId","unitId","userOwner","statusId","createdAt","createdBy")
        VALUES ($1,$2,$3,$4,$5,NOW(),$6)
    `, orgId, req.CaseID, "case", username, req.StatusID, username)
	if err != nil {
		return err
	}

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
	err = conn.QueryRow(ctx, query, req.WfID, req.NodeID, orgId).Scan(
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

func UpsertDeviceIot(conn *pgx.Conn, ctx context.Context, orgId string, iot *model.IotInfo, username string) error {
	now := time.Now()

	query := `
    INSERT INTO public.device_iot(
        "orgId", "deviceId", "deviceType", "model", "firmwareVer", "latitude", "longitude",
        "ipAddress", "macAddress", "createdAt", "updatedAt", "createdBy", "updatedBy"
    ) VALUES (
        $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
    )
    ON CONFLICT ("deviceId") DO UPDATE SET
        "orgId" = EXCLUDED."orgId",
        "deviceType" = EXCLUDED."deviceType",
        "model" = EXCLUDED."model",
        "firmwareVer" = EXCLUDED."firmwareVer",
        "latitude" = EXCLUDED."latitude",
        "longitude" = EXCLUDED."longitude",
        "ipAddress" = EXCLUDED."ipAddress",
        "macAddress" = EXCLUDED."macAddress",
        "updatedAt" = EXCLUDED."updatedAt",
        "updatedBy" = EXCLUDED."updatedBy";
    `

	_, err := conn.Exec(ctx, query,
		orgId,
		iot.DeviceID,
		iot.DeviceType,
		iot.Model,
		iot.FirmwareVer,
		iot.Latitude,
		iot.Longitude,
		iot.IPAddress,
		iot.MacAddress,
		now,
		now,
		username,
		username,
	)

	return err
}
