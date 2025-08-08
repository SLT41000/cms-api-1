package handler

import (
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
