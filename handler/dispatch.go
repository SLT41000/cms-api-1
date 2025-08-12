package handler

import (
	"context"
	"fmt"
	"mainPackage/config"
	"mainPackage/model"
	"net/http"

	"log"

	"github.com/gin-gonic/gin"
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
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
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
	log.Println(orgId)
	allNodes, currentNode, err := GetWorkflowAndCurrentNode(c, orgId.(string), caseId)
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
	log.Println(cusCase)
	log.Println("=xcxxxx==allNodes=x=x=x=x=x")
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

func GetWorkflowAndCurrentNode(c *gin.Context, orgId, caseId string) ([]model.WorkflowNode, *model.CurrentStage, error) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return nil, nil, nil
	}
	defer cancel()
	defer conn.Close(ctx)

	// 🔹 Step 1: Get current node and wfId
	currentQuery := `
		SELECT "wfId", "caseId", "nodeId", "versions", "type", "section", "data", "pic", "group", "formId"
		FROM tix_case_current_stage
		WHERE "orgId"=$1 AND "caseId"=$2
	`

	var current model.CurrentStage
	var wfId string

	err := conn.QueryRow(ctx, currentQuery, orgId, caseId).
		Scan(&wfId, &current.CaseId, &current.NodeId, &current.Versions, &current.Type, &current.Section, &current.Data, &current.Pic, &current.Group, &current.FormId)
	if err != nil {
		logger.Error("Failed to fetch current stage", zap.Error(err))
		return nil, nil, fmt.Errorf("current node not found for caseId=%s", caseId)
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
		return nil, nil, err
	}
	defer rows.Close()

	var allNodes []model.WorkflowNode
	for rows.Next() {
		var node model.WorkflowNode
		if err := rows.Scan(&node.NodeId, &node.Type, &node.Section, &node.Data); err != nil {
			logger.Error("Row scan failed", zap.Error(err))
			return nil, nil, err
		}
		allNodes = append(allNodes, node)
	}

	log.Println("===== all workflow nodes =====")
	log.Println(allNodes)

	return allNodes, &current, nil
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
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
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
    mu."updatedBy"
FROM users_on_units u
JOIN users_with_skill us ON u."username" = us."userName"
JOIN users_in_area ua ON u."username" = ua."username"
JOIN "mdm_units" mu ON mu."unitId" = u."unitId";
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
			&u.CreatedBy, &u.UpdatedBy,
		); err != nil {
			logger.Warn("Row scan failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.Response{
				Status: "-1",
				Msg:    "Failure",
				Desc:   err.Error(),
			})
			return
		}
		results = append(results, u)
	}

	c.JSON(http.StatusOK, model.Response{
		Status: "0",
		Msg:    "OK",
		Data:   results,
	})
}
