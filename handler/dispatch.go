package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"mainPackage/model"
	"mainPackage/utils"
	"net/http"
	"os"
	"time"

	"log"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// @summary Get SOP
// @tags Dispatch
// @security ApiKeyAuth
// @id Case By CaseId
// @accept json
// @produce json
// @Param caseId path string true "caseId"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/dispatch/{caseId}/SOP [get]
func GetSOP(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}

	defer cancel()
	defer conn.Close(ctx)

	if conn == nil {
		return
	}

	fmt.Println("=GetSOP==xx=x=x=x=x=x")
	log.Println("===")

	orgId := GetVariableFromToken(c, "orgId")
	caseId := c.Param("caseId")

	query := `SELECT id, "orgId", "caseId", "caseVersion", "referCaseId", "caseTypeId", "caseSTypeId", priority, "wfId", "versions", source, "deviceId", "phoneNo", "phoneNoHide", "caseDetail", "extReceive", "statusId", "caseLat", "caseLon", "caselocAddr", "caselocAddrDecs", "countryId", "provId", "distId", "caseDuration", "createdDate", "startedDate", "commandedDate", "receivedDate", "arrivedDate", "closedDate", usercreate, usercommand, userreceive, userarrive, userclose, "resId", "resDetail", "createdAt", "updatedAt", "createdBy", "updatedBy", "caseSla", "deviceMetaData"
	FROM public.tix_cases WHERE "orgId"=$1 AND "caseId"=$2`
	logger.Debug(`Query`, zap.String("query", query))
	var cusCase model.Case
	err := conn.QueryRow(ctx, query, orgId, caseId).Scan(
		&cusCase.ID,
		&cusCase.OrgID,
		&cusCase.CaseID,
		&cusCase.CaseVersion,
		&cusCase.ReferCaseID,
		&cusCase.CaseTypeID,
		&cusCase.CaseSTypeID,
		&cusCase.Priority,
		&cusCase.WfID,
		&cusCase.WfVersions,
		&cusCase.Source,
		&cusCase.DeviceID,
		&cusCase.PhoneNo,
		&cusCase.PhoneNoHide,
		&cusCase.CaseDetail,
		&cusCase.ExtReceive,
		&cusCase.StatusID,
		&cusCase.CaseLat,
		&cusCase.CaseLon,
		&cusCase.CaseLocAddr,
		&cusCase.CaseLocAddrDecs,
		&cusCase.CountryID,
		&cusCase.ProvID,
		&cusCase.DistID,
		&cusCase.CaseDuration,
		&cusCase.CreatedDate,
		&cusCase.StartedDate,
		&cusCase.CommandedDate,
		&cusCase.ReceivedDate,
		&cusCase.ArrivedDate,
		&cusCase.ClosedDate,
		&cusCase.UserCreate,
		&cusCase.UserCommand,
		&cusCase.UserReceive,
		&cusCase.UserArrive,
		&cusCase.UserClose,
		&cusCase.ResID,
		&cusCase.ResDetail,
		&cusCase.CreatedAt,
		&cusCase.UpdatedAt,
		&cusCase.CreatedBy,
		&cusCase.UpdatedBy,
		&cusCase.CaseSLA,
		&cusCase.DeviceMetaData,
	)

	if err != nil {
		logger.Warn("Query failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		return
	}
	log.Println("=orgId--")
	log.Println(orgId)
	allNodes, currentNode, nextStage, dispatchNode, err := GetWorkflowAndCurrentNode(c, orgId.(string), caseId, "")
	if err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failed",
			Desc:   err.Error(),
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	cusCase.SOP = allNodes
	cusCase.CurrentStage = currentNode
	cusCase.NextStage = nextStage
	cusCase.DispatchStage = dispatchNode
	log.Println(cusCase)
	log.Println("=GetSOP==allNodes=x=x=x=x=x")
	log.Println(allNodes)
	log.Println(currentNode)

	//Get Reference Case
	referCaseLists, err := GetReferCaseList(ctx, conn, orgId.(string), caseId)
	if err != nil {
		panic(err)
	}
	cusCase.ReferCaseLists = referCaseLists

	//Get Units
	unitLists, count, err := GetUnits(ctx, conn, orgId.(string), caseId, "", "")
	if err != nil {
		panic(err)
	}
	log.Print(unitLists, count)
	cusCase.UnitLists = unitLists

	//Get Cuurent dynamic form
	formId := currentNode.FormId // จาก JSON
	log.Print("====formId==")
	log.Print(formId)
	answers, err := GetFormAnswers(conn, ctx, orgId.(string), caseId, formId, false)
	if err != nil {
		log.Fatal("GetFormAnswers error:", err)
	}
	cusCase.FormAnswer = answers
	//Get SLA
	slaTimelines, err := GetSLA(c, conn, orgId.(string), caseId, "case")
	if err != nil {
		log.Fatal("GetSLA error:", err)
	}
	cusCase.SlaTimelines = slaTimelines

	//Get Attachment
	attachments, err := GetCaseAttachments(ctx, conn, orgId.(string), caseId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	cusCase.Attachments = attachments

	// Final JSON
	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Data:   cusCase,
		Desc:   "",
	}
	c.JSON(http.StatusOK, response)

	paramQuery := c.Request.URL.RawQuery
	logStr := Process("ListCase", paramQuery, response.Status, paramQuery, response)
	logger.Info(logStr)
}

func GetWorkflowAndCurrentNode(c *gin.Context, orgId, caseId string, unitId string) ([]model.WorkflowNode, *model.CurrentStage, *model.WorkflowNode, *model.WorkflowNode, error) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return nil, nil, nil, nil, nil
	}
	defer cancel()
	defer conn.Close(ctx)

	// 🔹 Step 1: Get current node and wfId
	currentQuery := `
		SELECT "wfId", "caseId", "nodeId", "versions", "type", "section", "data", "pic", "group", "formId"
		FROM tix_case_current_stage
		WHERE "orgId"=$1 AND "caseId"=$2 AND "stageType" = 'case'  AND "unitId" =$3
	`
	if unitId != "" {
		currentQuery = `
		SELECT "wfId", "caseId", "nodeId", "versions", "type", "section", "data", "pic", "group", "formId"
		FROM tix_case_current_stage
		WHERE "orgId"=$1 AND "caseId"=$2 AND "stageType" = 'unit' AND "unitId" =$3
	`
	}

	var current model.CurrentStage
	var wfId string
	log.Print(currentQuery)
	err := conn.QueryRow(ctx, currentQuery, orgId, caseId, unitId).
		Scan(&wfId, &current.CaseId, &current.NodeId, &current.Versions, &current.Type, &current.Section, &current.Data, &current.Pic, &current.Group, &current.FormId)
	if err != nil {
		logger.Error("Failed to fetch current stage", zap.Error(err))
		return nil, nil, nil, nil, fmt.Errorf("current node not found for caseId=%s", caseId)
	}

	log.Println("===== current stage =====")
	log.Printf("wfId: %s, current: %+v\n", wfId, current)

	// 🔹 Step 2: Get all workflow nodes using wfId
	nodesQuery := `
		SELECT "nodeId", "type", "section", "data"
		FROM wf_nodes
		WHERE "orgId"=$1 AND "wfId"=$2 AND "versions"=$3
		ORDER BY 
			CASE 
				WHEN "section" = 'nodes' THEN 1 
				WHEN "section" = 'connections' THEN 2 
				ELSE 3 
			END
	`

	rows, err := conn.Query(ctx, nodesQuery, orgId, wfId, current.Versions)
	if err != nil {
		logger.Error("Failed to fetch workflow nodes", zap.Error(err))
		return nil, nil, nil, nil, err
	}
	defer rows.Close()

	var dispatchNode model.WorkflowNode
	var allNodes []model.WorkflowNode
	var nodeConn []model.WorkFlowConnection
	allNodesId := make(map[string]model.WorkflowNode)
	for rows.Next() {
		var node model.WorkflowNode
		if err := rows.Scan(&node.NodeId, &node.Type, &node.Section, &node.Data); err != nil {
			logger.Error("Row scan failed", zap.Error(err))
			return nil, nil, nil, nil, err
		}
		allNodesId[node.NodeId] = node
		if node.Type != "sla" {
			allNodes = append(allNodes, node)
		}

		log.Print("TYPE: ", node.Section)
		if node.Type == "dispatch" {
			dispatchNode = node
		}
		if node.Section == "connections" {
			log.Print("----CONNECTION")
			dataBytes, err := json.Marshal(node.Data)
			if err != nil {
				logger.Error("Failed to marshal connection data", zap.Error(err))
				continue
			}

			var conns []model.WorkFlowConnection
			if err := json.Unmarshal(dataBytes, &conns); err != nil {
				logger.Error("Unmarshal connection failed", zap.Error(err))
				continue
			}

			nodeConn = append(nodeConn, conns...)
		}
	}

	log.Println("===== all workflow nodes =====")
	//log.Println(allNodes)

	var nextNode model.WorkflowNode
	for _, wfConn := range nodeConn {
		log.Print(wfConn)
		if wfConn.Source == current.NodeId {
			candidate := allNodesId[wfConn.Target]

			// ถ้า node type เป็น sla ให้ข้ามไปยัง target ต่อไป
			for candidate.Type == "sla" {
				found := false
				for _, c := range nodeConn {
					if c.Source == candidate.NodeId && c.Label == "yes" {
						candidate = allNodesId[c.Target]
						log.Print("---candidate---")
						log.Print(candidate.Type)
						found = true
						break

					}
				}
				if !found {
					// ไม่มี target "yes" ต่อไปแล้ว ออกจาก loop
					break
				}
			}

			nextNode = candidate
			log.Printf("Next node (non-SLA): %+v\n", nextNode)
			break
		}
	}

	log.Println("===== next nodes =====")
	log.Print(nextNode)
	log.Println("===== END =====")

	//allNodes_ := AddPreviousSLA(allNodes)

	return allNodes, &current, &nextNode, &dispatchNode, nil
}

// @summary Get Unit
// @tags Dispatch
// @security ApiKeyAuth
// @id CaseByCaseId
// @accept json
// @produce json
// @Param caseId path string true "caseId"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/dispatch/{caseId}/units [get]
func GetUnit(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}

	defer cancel()
	defer conn.Close(ctx)

	if conn == nil {
		return
	}

	fmt.Println("=xcxxxx==xx=x=x=x=x=x")
	log.Println("===")

	orgId := GetVariableFromToken(c, "orgId")
	caseId := c.Param("caseId")

	//--Get Skill All
	Skills, err_ := GetUserSkills(ctx, conn, orgId.(string))
	log.Print("---Skills---")
	if err_ != nil {
		panic(err_)
	}

	log.Print(Skills)

	//--Get Property All
	Props, err_ := GetUnitProp(ctx, conn, orgId.(string))
	log.Print("---Props---")
	if err_ != nil {
		panic(err_)
	}

	log.Print(Props)

	query := `
  WITH case_info AS (
  SELECT 
    c."caseSTypeId", 
    c."countryId",
    c."provId",
    c."distId",
    s."unitPropLists", 
    s."userSkillList"
  FROM "tix_cases" c
  JOIN "case_sub_types" s ON c."caseSTypeId" = s."sTypeId"
  WHERE c."caseId" = $1
    AND s."active" = TRUE
),
unit_with_props AS (
  SELECT 
    "unitId", 
    array_agg("propId") AS props
  FROM "mdm_unit_with_properties"
  WHERE "active" = TRUE
  GROUP BY "unitId"
),
units_matched AS (
  SELECT u."unitId", u."unitName", p.props
  FROM "mdm_units" u
  JOIN unit_with_props p ON u."unitId" = p."unitId"
  CROSS JOIN case_info c
  WHERE u."active" = TRUE
    AND (
      SELECT COUNT(DISTINCT prop_uuid)
      FROM (
        SELECT (jsonb_array_elements_text(c."unitPropLists"::jsonb))::uuid AS prop_uuid
      ) AS required_props
      WHERE prop_uuid = ANY(p.props)
    ) = (SELECT jsonb_array_length(c."unitPropLists"::jsonb))
),
users_on_units AS (
  SELECT u."unitId", mdm."username"
  FROM units_matched u
  JOIN "mdm_units" mdm ON mdm."unitId" = u."unitId" AND mdm."active" = TRUE
  JOIN "um_users" um ON um."username" = mdm."username" AND um."active" = TRUE
),
users_with_skill AS (
  SELECT DISTINCT "userName"
  FROM "um_user_with_skills"
  WHERE "skillId" IN (
    SELECT (jsonb_array_elements_text(ci."userSkillList"::jsonb))::uuid
    FROM case_info ci
  )
  AND "active" = TRUE
),
users_in_area AS (
  SELECT "username"
  FROM "um_user_with_area_response" ua
  CROSS JOIN case_info c
  WHERE ua."orgId" = $2
    AND EXISTS (
      SELECT 1
      FROM jsonb_array_elements_text(ua."distIdLists") AS distId
      WHERE distId.value = c."distId"
    )
)
SELECT mu."orgId",
       mu."unitId",
       mu."unitName",
       mu."unitSourceId",
       mu."unitTypeId",
       mu."priority",
       mu."compId",
       mu."deptId",
       mu."commId",
       mu."stnId",
       mu."plateNo",
       mu."provinceCode",
       mu."active",
       mu."username",
       mu."isLogin",
       mu."isFreeze",
       mu."isOutArea",
       mu."locLat",
       mu."locLon",
       mu."locAlt",
       mu."locBearing",
       mu."locSpeed",
       mu."locProvider",
       mu."locGpsTime",
       mu."locSatellites",
       mu."locAccuracy",
       mu."locLastUpdateTime",
       mu."breakDuration",
       mu."healthChk",
       mu."healthChkTime",
       mu."sttId",
       mu."createdBy",
       mu."updatedBy",
       ci."unitPropLists",
       ci."userSkillList"
FROM users_on_units u
JOIN users_with_skill us ON u."username" = us."userName"
JOIN users_in_area ua ON u."username" = ua."username"
JOIN "mdm_units" mu ON mu."unitId" = u."unitId"
CROSS JOIN case_info ci;
`
	logger.Debug(`Query`, zap.String("query", query))
	rows, err := conn.Query(context.Background(), query, caseId, orgId)
	if err != nil {
		logger.Warn("Query failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		return
	}
	defer rows.Close()

	var results []model.UnitUser

	for rows.Next() {
		var u model.UnitUser
		if err := rows.Scan(
			&u.OrgID, &u.UnitID, &u.UnitName, &u.UnitSourceID, &u.UnitTypeID, &u.Priority,
			&u.CompID, &u.DeptID, &u.CommID, &u.StnID, &u.PlateNo, &u.ProvinceCode,
			&u.Active, &u.Username, &u.IsLogin, &u.IsFreeze, &u.IsOutArea,
			&u.LocLat, &u.LocLon, &u.LocAlt, &u.LocBearing, &u.LocSpeed, &u.LocProvider,
			&u.LocGpsTime, &u.LocSatellites, &u.LocAccuracy, &u.LocLastUpdateTime,
			&u.BreakDuration, &u.HealthChk, &u.HealthChkTime, &u.SttID,
			&u.CreatedBy, &u.UpdatedBy, &u.UnitPropLists, &u.UserSkillList,
		); err != nil {
			logger.Warn("Row scan failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.Response{
				Status: "-1",
				Msg:    "Failure",
				Desc:   err.Error(),
			})
			return
		}

		u.SkillLists = ConvertSkills(Skills, *u.UserSkillList)
		u.ProplLists = ConvertProps(Props, *u.UnitPropLists)

		results = append(results, u)
	}

	c.JSON(http.StatusOK, model.Response{
		Status: "0",
		Msg:    "OK",
		Data:   results,
	})
}

// @summary Dispatch unit follow SOP
// @tags Dispatch
// @security ApiKeyAuth
// @id updateUnit
// @accept json
// @produce json
// @param Body body model.UpdateStageRequest true "Update unit event"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/dispatch/event [post]
func UpdateCurrentStage(c *gin.Context) {
	logger := utils.GetLog()

	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)

	var req model.UpdateStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Insert failed", zap.Error(err))
		return
	}
	log.Print(req)

	// username := GetVariableFromToken(c, "username")
	// orgId := GetVariableFromToken(c, "orgId")

	results, err := UpdateCurrentStageCore(c, conn, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

// @summary Get SOP - UnitId
// @tags Dispatch
// @security ApiKeyAuth
// @id Case By UnitId
// @accept json
// @produce json
// @Param caseId path string true "caseId"
// @Param unitId path string true "unitId"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/dispatch/{caseId}/SOP/unit/{unitId} [get]
func GetUnitSOP(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}

	defer cancel()
	defer conn.Close(ctx)

	fmt.Println("=xcxxxx==xx=x=x=x=x=x")
	log.Println("===")

	orgId := GetVariableFromToken(c, "orgId")
	caseId := c.Param("caseId")
	unitId := c.Param("unitId")

	query := `SELECT id, "orgId", "caseId", "caseVersion", "referCaseId", "caseTypeId", "caseSTypeId", priority, "wfId", "versions", source, "deviceId", "phoneNo", "phoneNoHide", "caseDetail", "extReceive", "statusId", "caseLat", "caseLon", "caselocAddr", "caselocAddrDecs", "countryId", "provId", "distId", "caseDuration", "createdDate", "startedDate", "commandedDate", "receivedDate", "arrivedDate", "closedDate", usercreate, usercommand, userreceive, userarrive, userclose, "resId", "resDetail", "createdAt", "updatedAt", "createdBy", "updatedBy"
	FROM public.tix_cases WHERE "orgId"=$1 AND "caseId"=$2`
	logger.Debug(`Query`, zap.String("query", query))
	var cusCase model.Case
	err := conn.QueryRow(ctx, query, orgId, caseId).Scan(
		&cusCase.ID,
		&cusCase.OrgID,
		&cusCase.CaseID,
		&cusCase.CaseVersion,
		&cusCase.ReferCaseID,
		&cusCase.CaseTypeID,
		&cusCase.CaseSTypeID,
		&cusCase.Priority,
		&cusCase.WfID,
		&cusCase.WfVersions,
		&cusCase.Source,
		&cusCase.DeviceID,
		&cusCase.PhoneNo,
		&cusCase.PhoneNoHide,
		&cusCase.CaseDetail,
		&cusCase.ExtReceive,
		&cusCase.StatusID,
		&cusCase.CaseLat,
		&cusCase.CaseLon,
		&cusCase.CaseLocAddr,
		&cusCase.CaseLocAddrDecs,
		&cusCase.CountryID,
		&cusCase.ProvID,
		&cusCase.DistID,
		&cusCase.CaseDuration,
		&cusCase.CreatedDate,
		&cusCase.StartedDate,
		&cusCase.CommandedDate,
		&cusCase.ReceivedDate,
		&cusCase.ArrivedDate,
		&cusCase.ClosedDate,
		&cusCase.UserCreate,
		&cusCase.UserCommand,
		&cusCase.UserReceive,
		&cusCase.UserArrive,
		&cusCase.UserClose,
		&cusCase.ResID,
		&cusCase.ResDetail,
		&cusCase.CreatedAt,
		&cusCase.UpdatedAt,
		&cusCase.CreatedBy,
		&cusCase.UpdatedBy,
	)

	if err != nil {
		logger.Warn("Query failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		return
	}
	log.Println("=orgId--")
	log.Println(unitId)
	allNodes, currentNode, nextStage, dispatchNode, err := GetWorkflowAndCurrentNode(c, orgId.(string), caseId, unitId)
	if err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failed",
			Desc:   err.Error(),
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	cusCase.SOP = allNodes
	cusCase.CurrentStage = currentNode
	st := []string{"S007", "S016", "S017", "S018"}

	// If statusId is in the skip list, return empty slice
	raw, _ := json.Marshal(currentNode.Data)
	var stageData struct {
		Data struct {
			Config struct {
				Action string `json:"action"`
			} `json:"config"`
		} `json:"data"`
	}
	_ = json.Unmarshal(raw, &stageData)
	action := stageData.Data.Config.Action
	fmt.Println("action =", action)

	cusCase.NextStage = nextStage
	if contains(st, action) {
		cusCase.NextStage = nil
	}

	//Get SLA
	slaTimelines, err := GetSLA(c, conn, orgId.(string), caseId, unitId)
	if err != nil {
		log.Fatal("query error:", err)
	}
	cusCase.SlaTimelines = slaTimelines

	// cusCase.DispatchStage = dispatchNode
	log.Println(dispatchNode)
	log.Println("=GetUnitSOP==allNodes=x=x=x=x=x")
	log.Println(allNodes)
	log.Println(currentNode)
	// Final JSON
	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Data:   cusCase,
		Desc:   "",
	}
	c.JSON(http.StatusOK, response)

	paramQuery := c.Request.URL.RawQuery
	logStr := Process("ListCase", paramQuery, response.Status, paramQuery, response)
	logger.Info(logStr)
}

// GetReferCaseList calls the PostgreSQL function GetReferCaseId and returns the caseList
func GetReferCaseList(ctx context.Context, conn *pgx.Conn, orgID string, caseID string) ([]string, error) {
	var referCaseList []string

	query := `
	SELECT ARRAY(
		SELECT "caseId"
		FROM public.tix_cases
		WHERE "orgId" = $1 AND ( "referCaseId" = $2)
	) AS caseList;
	`

	err := conn.QueryRow(ctx, query, orgID, caseID).Scan(&referCaseList)
	if err != nil {
		return []string{}, nil
	}

	return referCaseList, nil
}

// GetUnits returns a list of units (unitId, username)
func GetUnits(ctx context.Context, conn *pgx.Conn, orgID string, caseID string, statusId string, unitID string) ([]model.UnitDispatch, int, error) {
	var unitLists []model.UnitDispatch
	//st := []string{"S007", "S016", "S017", "S018"}

	// If statusId is in the skip list, return empty slice
	// if contains(st, statusId) {
	// 	//return unitLists, nil
	// }

	query := `
		SELECT 
    cs."unitId",
    cs."username",
    u."firstName",
    u."lastName"
FROM public.tix_case_current_stage AS cs
JOIN public.um_users AS u
    ON cs."username" = u."username"
    AND cs."orgId" = u."orgId"
WHERE 
    cs."orgId" = $1
    AND cs."caseId" = $2
    AND cs."stageType" = 'unit';
	`

	args := []interface{}{orgID, caseID}
	argIndex := 3

	// ถ้ามี statusId ให้เพิ่ม filter จาก JSONB
	if statusId != "" {
		query += fmt.Sprintf(` AND data->'data'->'config'->>'action' = $%d`, argIndex)
		args = append(args, statusId)
		argIndex++
	}

	// ถ้ามี unitID ให้เพิ่ม filter ด้วย
	if unitID != "" {
		query += fmt.Sprintf(` AND "unitId" = $%d`, argIndex)
		args = append(args, unitID)
		argIndex++
	}
	log.Print(query)
	log.Print(args)
	rows, err := conn.Query(ctx, query, args...)

	//rows, err := conn.Query(ctx, query, orgID, caseID)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var u model.UnitDispatch
		if err := rows.Scan(&u.UnitID, &u.Username, &u.FirstName, &u.LastName); err != nil {
			return nil, 0, err
		}
		unitLists = append(unitLists, u)
	}

	count := len(unitLists)

	return unitLists, count, nil
}

func GetFormAnswers(conn *pgx.Conn, ctx context.Context, orgId, caseId, formId string, returnUid bool) (*model.FormAnswerRequest, error) {
	// Query both form metadata and answers
	query := `
		SELECT 
			fb."formName",
			fb."formColSpan",
			fb."versions" as formVersion,
			fa."eleData",
			fa."id" as uid
		FROM form_builder fb
		LEFT JOIN form_answers fa
			ON fb."orgId" = fa."orgId"::uuid
			AND fb."formId" = fa."formId"::uuid
			AND fa."caseId" = $1
		WHERE fb."orgId" = $2
			AND fb."formId" = $3
		ORDER BY fa."eleNumber" ASC NULLS LAST
	`

	rows, err := conn.Query(ctx, query, caseId, orgId, formId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var formName string
	var formColSpan int
	var formVersion string
	var formFieldJson []model.IndividualFormField

	for rows.Next() {
		var eleDataBytes []byte
		var uid sql.NullString

		if err := rows.Scan(&formName, &formColSpan, &formVersion, &eleDataBytes, &uid); err != nil {
			return nil, err
		}

		if len(eleDataBytes) > 0 {
			var field model.IndividualFormField
			if err := json.Unmarshal(eleDataBytes, &field); err != nil {
				return nil, err
			}
			if returnUid && uid.Valid {
				field.UID = &uid.String
			}
			formFieldJson = append(formFieldJson, field)
		}
	}

	response := model.FormAnswerRequest{
		Versions: formVersion,
		// WfId:          "", // optionally fill from your workflow
		FormId:        formId,
		FormName:      formName,
		FormColSpan:   formColSpan,
		FormFieldJson: formFieldJson,
	}
	return &response, nil
}

func GetSLA(ctx *gin.Context, conn *pgx.Conn, orgID string, caseID string, unitId string) ([]model.CaseResponderCustom, error) {
	// 1. Get master status list
	statuses, err := utils.GetCaseStatusList(ctx, conn, orgID)
	if err != nil {
		return nil, err
	}

	// Build lookup map for quick access
	statusMap := make(map[string]model.CaseStatus)
	for _, s := range statuses {
		if s.StatusID != nil {
			statusMap[*s.StatusID] = s
		}
	}

	// 2. Query responders
	query := `
        SELECT "orgId", "caseId", "unitId", "userOwner", "statusId" , "createdAt"
        FROM public.tix_case_responders
        WHERE "orgId" = $1 AND "caseId" = $2 AND "unitId" = $3 ORDER BY "createdAt" asc
    `

	rows, err := conn.Query(ctx, query, orgID, caseID, unitId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var responders []model.CaseResponderCustom

	for rows.Next() {
		var r model.CaseResponderCustom
		if err := rows.Scan(&r.OrgID, &r.CaseID, &r.UnitID, &r.UserOwner, &r.StatusID, &r.CreatedAt); err != nil {
			return nil, err
		}

		// 3. Enrich with status.th & status.en
		if status, ok := statusMap[r.StatusID]; ok {
			r.StatusTh = status.Th
			r.StatusEn = status.En
		}

		responders = append(responders, r)
	}

	respondersNew := CalSLA(responders)
	//log.Print(respondersNew)
	return respondersNew, nil
}

// @summary Cancel unit assigned to a case
// @Description Cancel the current unit assignment for a case. This operation can only be performed if the current stage status is **S003 (ASSIGNED)**.
// If the case has only one assigned unit, the case status will be reset to **S001 (NEW)**.
// @tags Dispatch
// @security ApiKeyAuth
// @id CancelUnit
// @Accept json
// @Produce json
// @param Body body model.CancelUnitRequest true "Update unit event"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/dispatch/cancel/unit [post]
func DispatchCancelUnit(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}

	defer cancel()
	defer conn.Close(ctx)

	var req model.CancelUnitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		return
	}

	orgId := GetVariableFromToken(c, "orgId")
	username := GetVariableFromToken(c, "username")
	caseId := req.CaseId
	new_ := os.Getenv("NEW")
	assign_ := os.Getenv("ASSIGNED")
	cancel_ := os.Getenv("CANCEL")

	// ✅ Console log all parameters
	logger.Info("DispatchCancelUnit parameters",
		zap.String("orgId", fmt.Sprintf("%v", orgId)),
		zap.String("username", fmt.Sprintf("%v", username)),
		zap.String("caseId", caseId),
		zap.Any("request_body", req),
		zap.String("NEW", new_),
		zap.String("ASSIGNED", assign_),
		zap.String("CANCEL", cancel_),
	)
	//[1] => Check Unit Status S003
	unitLists, count, err := GetUnits(ctx, conn, orgId.(string), caseId, assign_, req.UnitId)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Status: "-1",
			Msg:    "Get unitLists Fail ",
			Desc:   err.Error(),
		})
		return
	}
	logger.Debug("unitLists-1",
		zap.Int("count", count),
		zap.Any("units", unitLists),
	)
	if count == 0 {
		c.JSON(http.StatusBadRequest, model.Response{
			Status: "-1",
			Msg:    "Invalid Unit " + req.UnitId + " on CaseId " + req.CaseId,
			Desc:   "Unit not found or not in assigned state (S003)",
		})
		return
	}

	//[2] Insert responder
	log.Print("--> 1. Insert responder")
	_, err = conn.Exec(ctx, `
		    INSERT INTO tix_case_responders ("orgId","caseId","unitId","userOwner","statusId","createdAt","createdBy")
		    VALUES ($1,$2,$3,$4,$5,NOW(),$6)
		`, orgId, req.CaseId, req.UnitId, req.UnitUser, cancel_, username)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Status: "-1",
			Msg:    "Insert Responder Fail ",
			Desc:   err.Error(),
		})
		return
	}

	//[3] => Delete Unit
	deletedCount, err := DeleteUnit(ctx, conn, orgId.(string), caseId, "", req.UnitId)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Status: "-1",
			Msg:    "Delete Unit Fail ",
			Desc:   err.Error(),
		})
		return
	}
	if deletedCount > 0 {
		log.Printf("Successfully deleted %d unit(s)", deletedCount)
	} else {
		log.Printf("No units deleted (maybe not found or status skipped)")
	}

	//[4] => Check Unit Count && Update Case status S001
	unitLists, count, err = GetUnits(ctx, conn, orgId.(string), caseId, "", "")
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Status: "-2",
			Msg:    "Get unitLists Fail ",
			Desc:   err.Error(),
		})
		return
	}
	logger.Debug("unitLists-2",
		zap.Int("count", count),
		zap.Any("units", unitLists),
	)
	if count == 0 {
		//Change Case Status S001
		err := UpdateCancelCaseForUnit(ctx, conn, orgId.(string), caseId, req.ResId, req.ResDetail, new_, username.(string))
		if err != nil {
			logger.Error("UpdateCancelCase failed", zap.Error(err))
		} else {
			logger.Info("Case status updated successfully",
				zap.String("caseId", caseId),
				zap.String("newStatus", new_),
			)
			//Change Current Case Status =  S001
			err := UpdateStageByAction(ctx, conn, orgId.(string), caseId, username.(string))
			if err != nil {
				c.JSON(http.StatusBadRequest, model.Response{
					Status: "-1",
					Msg:    "Update stage failed",
					Desc:   err.Error(),
				})
				return
			}

		}
	}

	//[5] => Alert & Event S013
	req_ := model.UpdateStageRequest{
		CaseId:   caseId,
		Status:   cancel_,
		UnitUser: req.UnitUser, // หรือ set ค่า default
	}
	log.Print(req)
	GenerateNotiAndComment(c, conn, req_, orgId.(string), "0")

	// req_.UnitUser = ""
	// UpdateBusKafka_WO(c, conn, req_)

	c.JSON(http.StatusOK, model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Unit cancelled successfully",
	})
}

func DeleteUnit(ctx context.Context, conn *pgx.Conn, orgID, caseID, statusID, unitID string) (int64, error) {

	query := `
		DELETE FROM public.tix_case_current_stage
		WHERE "orgId" = $1 
		  AND "caseId" = $2 
		  AND "stageType" = 'unit'
	`
	args := []interface{}{orgID, caseID}
	argIndex := 3

	// ถ้ามี unitID ให้เพิ่ม filter
	if unitID != "" {
		query += fmt.Sprintf(` AND "unitId" = $%d`, argIndex)
		args = append(args, unitID)
	}

	// Execute delete query
	cmdTag, err := conn.Exec(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("DeleteUnit failed: %w", err)
	}

	// cmdTag.RowsAffected() คืนค่าจำนวนแถวที่ถูกลบ
	deletedCount := cmdTag.RowsAffected()
	log.Printf("Deleted %d unit(s) from case %s", deletedCount, caseID)

	return deletedCount, nil
}

func UpdateCancelCaseForUnit(ctx context.Context, conn *pgx.Conn, orgID string, caseID string, resId string, resDetail string, newStatus string, updatedBy string) error {

	if resDetail == "" {
		query := `
		UPDATE public.tix_cases
		SET "statusId" = $3,  
			"updatedAt" = NOW(),
			"updatedBy" = $4
		WHERE "orgId" = $1 AND "caseId" = $2
	`

		cmdTag, err := conn.Exec(ctx, query, orgID, caseID, newStatus, updatedBy)
		if err != nil {
			return fmt.Errorf("failed to update case status: %w", err)
		}

		if cmdTag.RowsAffected() == 0 {
			return fmt.Errorf("no case found with caseId=%s", caseID)
		}
	} else {
		query := `
		UPDATE public.tix_cases
		SET "statusId" = $3,
			"resId" = $4,
			"resDetail" = $5,
			"updatedAt" = NOW(),
			"updatedBy" = $6
		WHERE "orgId" = $1 AND "caseId" = $2
	`

		cmdTag, err := conn.Exec(ctx, query, orgID, caseID, newStatus, resId, resDetail, updatedBy)
		if err != nil {
			return fmt.Errorf("failed to update case status: %w", err)
		}

		if cmdTag.RowsAffected() == 0 {
			return fmt.Errorf("no case found with caseId=%s", caseID)
		}
	}

	return nil
}

// UpdateStageByAction combines GetNodeByAction + UpdateCurrentStage
func UpdateStageByAction(ctx context.Context, conn *pgx.Conn, orgId, caseId, username string) error {

	provID, wfId, versions, err := GetInfoFromCase(ctx, conn, orgId, caseId)
	if err != nil {
		log.Printf("error getting provId: %v", err)
	} else {
		log.Printf("provId = %s", provID)
		log.Print(provID, wfId, versions)
	}

	// 1️⃣ Get Node by action (e.g., S001)
	node, err := GetNodeByAction(ctx, conn, orgId, wfId, versions, "S001")
	if err != nil {
		return fmt.Errorf("failed to get node by action: %v", err)
	}

	log.Print("====>> GetNodeByAction")
	log.Print(node)
	// 2️⃣ Update current stage with node data
	err = UpdateStage(ctx, conn, orgId, caseId, node.NodeId, node.Data, username)
	if err != nil {
		return fmt.Errorf("failed to update current stage: %v", err)
	}

	return nil
}

// GetNodeByAction retrieves one node where data->'data'->'config'->>'action' = given action
func GetNodeByAction(ctx context.Context, conn *pgx.Conn, orgId, wfId string, version string, action string) (*model.WorkflowNode, error) {
	query := `
		SELECT "nodeId", "type", "section", "data"
		FROM wf_nodes
		WHERE "orgId" = $1
		  AND "wfId" = $2
		  AND "versions" = $3
		  AND "data"->'data'->'config'->>'action' = $4
		  AND "section" = 'nodes'
		LIMIT 1;
	`

	log.Print("--GetNodeByAction--")
	log.Print(query)
	row := conn.QueryRow(ctx, query, orgId, wfId, version, action)

	var node model.WorkflowNode
	var dataRaw []byte

	err := row.Scan(&node.NodeId, &node.Type, &node.Section, &dataRaw)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // no result found
		}
		return nil, fmt.Errorf("GetNodeByAction query failed: %v", err)
	}

	if err := json.Unmarshal(dataRaw, &node.Data); err != nil {
		return nil, fmt.Errorf("error unmarshalling node data: %v", err)
	}

	return &node, nil
}

func UpdateStage(ctx context.Context, conn *pgx.Conn, orgId, caseId, nodeId string, newData interface{}, username string) error {
	jsonData, err := json.Marshal(newData)
	if err != nil {
		return fmt.Errorf("marshal error: %v", err)
	}

	// ดึง formId จาก newData
	var parsed struct {
		Data struct {
			Config struct {
				FormID string `json:"formId"`
			} `json:"config"`
		} `json:"data"`
	}

	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		return fmt.Errorf("unmarshal error: %v", err)
	}

	formId := parsed.Data.Config.FormID
	if formId == "" {
		formId = "unknown" // fallback
	}

	log.Print("====>> UpdateStage")
	log.Printf("nodeId=%s formId=%s", nodeId, formId)

	log.Print("====>> UpdateStage")
	log.Print(string(jsonData))

	query := `
		UPDATE public.tix_case_current_stage
		SET 
		    "nodeId" = $3,
			"type" = $4,
			"data" = $5,
			"formId" = $6,
			"updatedAt" = $7,
			"updatedBy" = $8
		WHERE "orgId" = $1 AND "caseId" = $2 AND "stageType" = 'case'
	`

	_, err = conn.Exec(ctx, query, orgId, caseId, nodeId, "process", jsonData, formId, time.Now(), username)
	if err != nil {
		log.Print("==UpdateStage Error -=-")
		log.Print(err)
		return fmt.Errorf("update current stage failed: %v", err)
	}
	return nil
}

// @summary Cancel Case and all units
// @Cancel Case and All Unit
// @tags Dispatch
// @security ApiKeyAuth
// @id CancelCase
// @Accept json
// @Produce json
// @param Body body model.CancelCaseRequest true "Update unit event"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/dispatch/cancel/case [post]
func DispatchCancelCase(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}

	defer cancel()
	defer conn.Close(ctx)

	var req model.CancelCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		return
	}
	orgId := GetVariableFromToken(c, "orgId")
	username := GetVariableFromToken(c, "username")
	caseId := req.CaseId
	new_ := os.Getenv("NEW")
	//assign_ := os.Getenv("ASSIGNED")
	cancel_ := os.Getenv("CANCEL_CASE")

	//[1] Insert responder
	// log.Print("--> 1. Insert Case responder ")
	// _, err := conn.Exec(ctx, `
	// 	    INSERT INTO tix_case_responders ("orgId","caseId","unitId","userOwner","statusId","createdAt","createdBy")
	// 	    VALUES ($1,$2,$3,$4,$5,NOW(),$6)
	// 	`, orgId, req.CaseId, "case", username, cancel_, username)
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, model.Response{
	// 		Status: "-1",
	// 		Msg:    "Insert Responder Fail ",
	// 		Desc:   err.Error(),
	// 	})
	// 	return
	// }

	//[2] => Delete all Unit
	deletedCount, err := DeleteUnit(ctx, conn, orgId.(string), caseId, "", "")
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Status: "-1",
			Msg:    "Delete Unit Fail ",
			Desc:   err.Error(),
		})
		return
	}
	if deletedCount > 0 {
		log.Printf("Successfully deleted %d unit(s)", deletedCount)
	} else {
		log.Printf("No units deleted (maybe not found or status skipped)")
	}

	err = UpdateCancelCaseForUnit(ctx, conn, orgId.(string), caseId, req.ResId, req.ResDetail, cancel_, username.(string))
	if err != nil {
		logger.Error("UpdateCancelCaseForUnit failed", zap.Error(err))
	} else {
		logger.Info("Case status updated successfully",
			zap.String("caseId", caseId),
			zap.String("newStatus", new_),
		)
	}

	//[3] => Alert & Event S013
	req_ := model.UpdateStageRequest{
		CaseId:   caseId,
		Status:   cancel_,
		UnitUser: username.(string), // หรือ set ค่า default
	}
	log.Print(req)
	GenerateNotiAndComment(c, conn, req_, orgId.(string), "0")

	c.JSON(http.StatusOK, model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Case cancelled successfully",
	})
}
