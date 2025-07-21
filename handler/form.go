package handler

import (
	"encoding/json"
	"mainPackage/config"
	"mainPackage/model"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// @summary Get Form
// @tags Form and Workflow
// @security ApiKeyAuth
// @id Get Form
// @accept json
// @produce json
// @Param id query string true "id"
// @Param version query string true "version"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/forms [get]
func GetForm(c *gin.Context) {
	logger := config.GetLog()
	id := c.Query("id")
	version := c.Query("version")
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	orgId := GetVariableFromToken(c, "orgId")
	query := `SELECT form_builder."formId",form_builder."formName",form_builder."formColSpan",form_elements."eleData" 
	FROM public.form_builder INNER JOIN public.form_elements ON form_builder."formId"=form_elements."formId" 
	WHERE form_builder."formId" = $1 AND form_builder."orgId"=$2 AND form_elements."versions"=$3`

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query))
	logger.Debug(`id :` + id)
	rows, err := conn.Query(ctx, query, id, orgId, version)
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
	var form model.Form
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

// @summary Create Form
// @tags Form and Workflow
// @security ApiKeyAuth
// @id Create Form
// @accept json
// @produce json
// @param Case body model.FormInsert true "Created Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/forms [post]
func FormInsert(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)

	var req model.FormInsert
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	uuid := uuid.New()
	now := time.Now()
	var id int
	var exists bool
	checkQuery := `
	SELECT EXISTS (
		SELECT 1 FROM public."form_builder" 
		WHERE "orgId" = $1 AND "formName" = $2
	);
`
	err := conn.QueryRow(ctx, checkQuery, orgId, req.FormName).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Insert failed", zap.Error(err))
		return
	}
	if exists {
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   "form name already exists",
		})
		logger.Warn("Insert failed form name already exists")
		return
	}
	query := `
	INSERT INTO public."form_builder"(
	"orgId", "formId", "formName", "formColSpan", active, publish, versions, locks, "createdAt", "updatedAt", "createdBy", "updatedBy")
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	RETURNING id ;
	`
	err = conn.QueryRow(ctx, query,
		orgId, uuid, req.FormName, req.FormColSpan, req.Active, req.Publish, "draft", req.Locks, now,
		now, username, username).Scan(&id)

	if err != nil {
		// log.Printf("Insert failed: %v", err)
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	for i, item := range req.FormFieldJson {
		logger.Debug("eleNumber", zap.Int("i", i+1))
		logger.Debug("JsonArray", zap.Any("Json", item))
		query := `
				INSERT INTO public.form_elements(
				"orgId", "formId", versions, "eleNumber", "eleData", "createdAt", "updatedAt", "createdBy", "updatedBy")
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
				RETURNING id ;
				`
		err = conn.QueryRow(ctx, query,
			orgId, uuid, "draft", i+1, item, now, now, username, username).Scan(&id)

		if err != nil {
			// log.Printf("Insert failed: %v", err)
			c.JSON(http.StatusInternalServerError, model.Response{
				Status: "-1",
				Msg:    "Failure",
				Desc:   err.Error(),
			})
			logger.Warn("Insert failed", zap.Error(err))
			return
		}
	}

	c.JSON(http.StatusOK, model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Create successfully",
	})
}

// @summary Update Form
// @tags Form and Workflow
// @security ApiKeyAuth
// @id Update Form
// @accept json
// @produce json
// @Param uuid path string true "uuid"
// @param Case body model.FormUpdate true "Update Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/forms/{uuid} [patch]
func FormUpdate(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	uuid := c.Param("uuid")
	var req model.FormUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	now := time.Now()
	var id int

	query := `
	UPDATE public.form_builder
	SET "formName"=$2, "formColSpan"=$3, active=$4, publish=$5,
	 versions=$6, locks=$7, "updatedAt"=$8,"updatedBy"=$9
	WHERE "formId"=$1 AND "orgId"=$10;
	`
	_, err := conn.Exec(ctx, query, uuid,
		req.FormName, req.FormColSpan, req.Active, req.Publish, "draft", req.Locks,
		now, username, orgId)

	if err != nil {
		// log.Printf("Insert failed: %v", err)
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	query = `
	DELETE FROM public.form_elements
	WHERE "formId"=$1;
	`
	_, err = conn.Exec(ctx, query, uuid)

	if err != nil {
		// log.Printf("Insert failed: %v", err)
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	for i, item := range req.FormFieldJson {
		now := time.Now()
		logger.Debug("eleNumber", zap.Int("i", i+1))
		logger.Debug("JsonArray", zap.Any("Json", item))
		query := `
				INSERT INTO public.form_elements(
				"orgId", "formId", versions, "eleNumber", "eleData", "createdAt", "updatedAt", "createdBy", "updatedBy")
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
				RETURNING id ;
				`
		err = conn.QueryRow(ctx, query,
			orgId, uuid, "draft", i+1, item, now, now, username, username).Scan(&id)

		if err != nil {
			// log.Printf("Insert failed: %v", err)
			c.JSON(http.StatusInternalServerError, model.Response{
				Status: "-1",
				Msg:    "Failure",
				Desc:   err.Error(),
			})
			logger.Warn("Insert failed", zap.Error(err))
			return
		}
	}

	c.JSON(http.StatusOK, model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Update successfully",
	})
}

// @summary Update Form Publish
// @tags Form and Workflow
// @security ApiKeyAuth
// @id Update Form Publish
// @accept json
// @produce json
// @param Case body model.FormPublish true "Update Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/forms/publish [patch]
func FormPublish(c *gin.Context) {

	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	var req model.FormPublish
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	now := time.Now()

	query := `
	UPDATE public.form_builder
	SET publish=$2, "updatedAt"=$3,"updatedBy"=$4
	WHERE "formId"=$1 AND "orgId"=$5;
	`
	_, err := conn.Exec(ctx, query, req.FormID, req.Publish, now, username, orgId)
	logger.Debug("Update Case SQL Args",
		zap.String("query", query),
		zap.Any("Input", []any{
			req.FormID, req.Publish, now, username, orgId,
		}))
	if err != nil {
		// log.Printf("Insert failed: %v", err)
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Insert failed", zap.Error(err))
		return
	}
	c.JSON(http.StatusOK, model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Update successfully",
	})
}

// @summary Update Form Lock
// @tags Form and Workflow
// @security ApiKeyAuth
// @id Update Form Lock
// @accept json
// @produce json
// @param Case body model.FormLock true "Update Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/forms/lock [patch]
func FormLock(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	var req model.FormLock
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	now := time.Now()

	query := `
	UPDATE public.form_builder
	SET locks=$2, "updatedAt"=$3,"updatedBy"=$4
	WHERE "formId"=$1 AND "orgId"=$5;
	`
	_, err := conn.Exec(ctx, query, req.FormID, req.Locks, now, username, orgId)
	logger.Debug("Update Case SQL Args",
		zap.String("query", query),
		zap.Any("Input", []any{
			req.FormID, req.Locks, now, username, orgId,
		}))

	if err != nil {
		// log.Printf("Insert failed: %v", err)
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Insert failed", zap.Error(err))
		return
	}
	c.JSON(http.StatusOK, model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Update successfully",
	})
}

// @summary Update Form Status
// @tags Form and Workflow
// @security ApiKeyAuth
// @id Update Form Status
// @accept json
// @produce json
// @param Case body model.FormActive true "Update Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/forms/active [patch]
func FormActive(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	var req model.FormActive
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	now := time.Now()

	query := `
	UPDATE public.form_builder
	SET active=$2, "updatedAt"=$3,"updatedBy"=$4
	WHERE "formId"=$1 AND "orgId"=$5;
	`
	_, err := conn.Exec(ctx, query, req.FormID, req.Active, now, username, orgId)
	logger.Debug("Update Case SQL Args",
		zap.String("query", query),
		zap.Any("Input", []any{
			req.FormID, req.Active, now, username, orgId,
		}))
	if err != nil {
		// log.Printf("Insert failed: %v", err)
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Insert failed", zap.Error(err))
		return
	}
	c.JSON(http.StatusOK, model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Update successfully",
	})
}

// @summary Delete Form
// @id Delete Form
// @security ApiKeyAuth
// @accept json
// @tags Form and Workflow
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/forms/{id} [delete]
func DeleteForm(c *gin.Context) {}

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
	var workflow model.WorkFlow
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
