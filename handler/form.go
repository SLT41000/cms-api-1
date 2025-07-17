package handler

import (
	"encoding/json"
	"mainPackage/config"
	"mainPackage/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// ListCase godoc
// @summary Get Form
// @tags Form and Workflow
// @security ApiKeyAuth
// @id Get Form
// @accept json
// @produce json
// @Param id path string true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/forms/{id} [get]
func GetForm(c *gin.Context) {
	logger := config.GetLog()
	id := c.Param("id")

	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	orgId := GetVariableFromToken(c, "orgId")
	query := `SELECT form_builder."formId",form_builder."formName",form_builder."formColSpan",form_elements."eleData" 
	FROM public.form_builder INNER JOIN public.form_elements ON form_builder."formId"=form_elements."formId" 
	WHERE form_builder."formId" = $1 AND form_builder."orgId"=$2`

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query))
	logger.Debug(`id :` + id)
	rows, err := conn.Query(ctx, query, id, orgId)
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
	var errorMsg string
	var formFields []map[string]interface{}
	var form model.FormGetOptModel
	for rows.Next() {
		var rawJSON []byte
		err := rows.Scan(&form.FormId, &form.FormName, &form.FormColSpan, &rawJSON)
		if err != nil {
			logger.Warn("Scan failed", zap.Error(err))
			continue
		}

		var field map[string]interface{}
		if err := json.Unmarshal(rawJSON, &field); err != nil {
			logger.Warn("unmarshal field failed", zap.Error(err))
			continue
		}

		formFields = append(formFields, field)
		form.FormFieldJson = formFields
	}
	if errorMsg != "" {
		response := model.Response{
			Status: "-1",
			Msg:    "Failed",
			Data:   form,
			Desc:   errorMsg,
		}
		c.JSON(http.StatusInternalServerError, response)
	} else {
		response := model.Response{
			Status: "0",
			Msg:    "Success",
			Data:   form,
			Desc:   "",
		}
		c.JSON(http.StatusOK, response)
	}
}

// ListCase godoc
// @summary Get Workflow
// @tags Form and Workflow
// @security ApiKeyAuth
// @id Get Workflow
// @accept json
// @produce json
// @Param id path string true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/workflows/{id} [get]
func GetWorkFlow(c *gin.Context) {
	logger := config.GetLog()
	id := c.Param("id")

	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	orgId := GetVariableFromToken(c, "orgId")
	query := `SELECT "type","data",title,"desc",wf_definitions."versions",wf_definitions."createdAt",wf_definitions."updatedAt" 
	FROM public.wf_definitions Inner join public.wf_nodes
	ON wf_definitions."wfId" = wf_nodes."wfId" WHERE wf_definitions."wfId" = $1 AND wf_nodes."orgId"=$2`

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query))
	logger.Debug(`id :` + id)
	rows, err := conn.Query(ctx, query, id, orgId)
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
	var errorMsg string
	var NodesArray []map[string]interface{}
	var ConnectionArray []map[string]interface{}
	var workflow model.WorkFlowGetOptModel
	var workflowMetaData model.WorkFlowMetadata
	rowIndex := 0
	for rows.Next() {
		rowIndex++
		var rawJSON []byte
		var rowsType string
		err := rows.Scan(&rowsType, &rawJSON, &workflowMetaData.Title, &workflowMetaData.Desc,
			&workflowMetaData.Status, &workflowMetaData.CreatedAt, &workflowMetaData.UpdatedAt)
		if err != nil {
			logger.Warn("Scan failed", zap.Error(err))
			response := model.Response{
				Status: "-1",
				Msg:    "Failed",
				Data:   workflow,
				Desc:   errorMsg,
			}
			c.JSON(http.StatusInternalServerError, response)
		}

		if rowsType == "nodes" {
			var field map[string]interface{}
			if err := json.Unmarshal(rawJSON, &field); err != nil {
				logger.Debug(string(rawJSON))
				logger.Warn("unmarshal field failed", zap.Error(err))
				continue
			}
			NodesArray = append(NodesArray, field)
			workflow.Nodes = NodesArray
		} else {
			var field []map[string]interface{}
			if err := json.Unmarshal(rawJSON, &field); err != nil {
				logger.Debug(string(rawJSON))
				logger.Warn("unmarshal field failed", zap.Error(err))
				continue
			}
			ConnectionArray = append(ConnectionArray, field...)
			workflow.Connections = ConnectionArray
		}
		workflow.MetaData = workflowMetaData
	}
	if errorMsg != "" {
		response := model.Response{
			Status: "-1",
			Msg:    "Failed",
			Data:   workflow,
			Desc:   errorMsg,
		}
		c.JSON(http.StatusInternalServerError, response)
	} else {
		response := model.Response{
			Status: "0",
			Msg:    "Success",
			Data:   workflow,
			Desc:   "",
		}
		c.JSON(http.StatusOK, response)
	}
}
