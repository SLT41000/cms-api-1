package handler

import (
	"mainPackage/model"
	"mainPackage/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// ListCase Status godoc
// @summary Get Case Status
// @tags Cases
// @security ApiKeyAuth
// @id Get Case Status
// @accept json
// @produce json
// @Param start query int false "start" default(0)
// @Param length query int false "length" default(10)
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/case_status [get]
func GetCaseStatus(c *gin.Context) {
	logger := utils.GetLog()

	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	startStr := c.DefaultQuery("start", "0")
	start, err := strconv.Atoi(startStr)
	id := c.Param("id")
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()

	if err != nil {
		start = 0
	}
	lengthStr := c.DefaultQuery("length", "1000")
	length, err := strconv.Atoi(lengthStr)
	if err != nil || length > 100 {
		length = 1000
	}

	// orgId := GetVariableFromToken(c, "orgId")
	query := `SELECT id, "statusId", th, en, color, active, "createdAt", "updatedAt", "createdBy", "updatedBy"
	FROM public.case_status WHERE active = TRUE ORDER BY "statusId" ASC LIMIT $1 OFFSET $2`

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query), zap.Any("query", []any{length, start}))
	rows, err = conn.Query(ctx, query, length, start)
	if err != nil {
		logger.Warn("Query failed", zap.Error(err))
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Cases", "GetCaseStatus", "",
			"search", -1, start_time, GetQueryParams(c), response, "Query failed = "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer rows.Close()
	var Department model.CaseStatus
	var DepartmentList []model.CaseStatus
	found := false
	for rows.Next() {
		err := rows.Scan(&Department.ID, &Department.StatusID, &Department.Th, &Department.En, &Department.Color,
			&Department.Active, &Department.CreatedAt, &Department.UpdatedAt, &Department.CreatedBy, &Department.UpdatedBy)
		if err != nil {
			logger.Warn("Scan failed", zap.Error(err))
			response := model.Response{
				Status: "-1",
				Msg:    "Failed",
				Desc:   err.Error(),
			}
			//=======AUDIT_START=====//
			_ = utils.InsertAuditLogs(
				c, conn, orgId.(string), username.(string),
				txtId, id, "Cases", "GetCaseStatus", "",
				"search", -1, start_time, GetQueryParams(c), response, "Scan failed = "+err.Error(),
			)
			//=======AUDIT_END=====//
			c.JSON(http.StatusInternalServerError, response)
			return
		}
		DepartmentList = append(DepartmentList, Department)
		found = true
	}
	if !found {
		response := model.Response{
			Status: "-1",
			Msg:    "Failed",
			Desc:   "not found",
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Cases", "GetCaseStatus", "",
			"search", -1, start_time, GetQueryParams(c), response, "Not Found.",
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
	} else {
		response := model.Response{
			Status: "0",
			Msg:    "Success",
			Data:   DepartmentList,
			Desc:   "",
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Cases", "GetCaseStatus", "",
			"search", 0, start_time, GetQueryParams(c), response, "GetCaseStatus Success.",
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusOK, response)
	}
}

// ListCase Status godoc
// @summary Get Case Status by id
// @tags Cases
// @security ApiKeyAuth
// @id Get Case Status by id
// @accept json
// @produce json
// @Param id path string true "id" "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/case_status/{id} [get]
func GetCaseStatusById(c *gin.Context) {
	logger := utils.GetLog()
	id := c.Param("id")
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	// orgId := GetVariableFromToken(c, "orgId")
	query := `SELECT id, "statusId", th, en, color, active, "createdAt", "updatedAt", "createdBy", "updatedBy"
	FROM public.case_status WHERE "id"=$1 `

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query))
	rows, err := conn.Query(ctx, query, id)
	if err != nil {
		logger.Warn("Query failed", zap.Error(err))
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Cases", "GetCaseStatusById", "",
			"search", -1, start_time, GetQueryParams(c), response, "Query failure : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer rows.Close()
	var errorMsg string
	var Department model.CaseStatus
	err = rows.Scan(&Department.ID, &Department.StatusID, &Department.Th, &Department.En, &Department.Color,
		&Department.Active, &Department.CreatedAt, &Department.UpdatedAt, &Department.CreatedBy, &Department.UpdatedBy)
	if err != nil {
		logger.Warn("Scan failed", zap.Error(err))
		errorMsg = err.Error()
		response := model.Response{
			Status: "-1",
			Msg:    "Failed",
			Desc:   errorMsg,
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Cases", "GetCaseStatusById", "",
			"search", -1, start_time, GetQueryParams(c), response, "Scan failure : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Data:   Department,
		Desc:   "",
	}
	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, id, "Cases", "GetCaseStatusById", "",
		"search", -1, start_time, GetQueryParams(c), response, "GetCaseStatusById Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)

}

// @summary Create Case Status
// @id Create Case Status
// @security ApiKeyAuth
// @tags Cases
// @accept json
// @produce json
// @param Body body model.CaseStatusInsert true "Create Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/case_status/add [post]
func InsertCaseStatus(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	var req model.CaseStatusInsert

	if err := c.ShouldBindJSON(&req); err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, "", "",
			"", "", "Cases", "InsertCaseStatus", "",
			"create", -1, time.Now(), GetQueryParams(c), response, "Failure : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusBadRequest, response)
		logger.Warn("Insert failed", zap.Error(err))
		return
	}
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	uuid := uuid.New()
	now := time.Now()
	var id int
	query := `
	INSERT INTO public."case_status"(
	"statusId", th, en, color, active, "createdAt", "updatedAt", "createdBy", "updatedBy")
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	RETURNING id ;
	`
	logger.Debug("Update Case SQL Args",
		zap.String("query", query),
		zap.Any("Input", []any{
			uuid, req.Th, req.En, req.Color, req.Active, now,
			now, username, username,
		}))
	err := conn.QueryRow(ctx, query,
		uuid, req.Th, req.En, req.Color, req.Active, now,
		now, username, username).Scan(&id)

	if err != nil {
		// log.Printf("Insert failed: %v", err)
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			uuid.String(), "", "Cases", "InsertCaseStatus", "",
			"create", -1, now, GetQueryParams(c), response, "Failure : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		logger.Warn("Insert failed", zap.Error(err))
		return
	}
	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Create successfully",
	}
	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		uuid.String(), "", "Cases", "InsertCaseStatus", "",
		"create", 0, now, GetQueryParams(c), response, "InsertCaseStatus Success.",
	)
	//=======AUDIT_END=====//
	// Continue logic...
	c.JSON(http.StatusOK, response)

}

// @summary Update Case Status
// @id Update Case Status
// @security ApiKeyAuth
// @accept json
// @produce json
// @tags Cases
// @Param id path int true "id"
// @param Body body model.CaseStatusUpdate true "Update data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/case_status/{id} [patch]
func UpdateCaseStatus(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	id := c.Param("id")
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()
	var req model.CaseStatusUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Update failed", zap.Error(err))
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Cases", "UpdateCaseStatus", "",
			"update", -1, start_time, GetQueryParams(c), response, "Failure : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusBadRequest, response)
		return
	}
	now := time.Now()

	// orgId := GetVariableFromToken(c, "orgId")
	query := `UPDATE public."case_status"
	SET th=$2, en=$3, "color"=$4,active=$5,
	 "updatedAt"=$6, "updatedBy"=$7
	WHERE id = $1 `
	_, err := conn.Exec(ctx, query,
		id, req.Th, req.En, req.Color, req.Active,
		now, username,
	)
	logger.Debug("Update Case SQL Args",
		zap.String("query", query),
		zap.Any("Input", []any{
			id, req.Th, req.En, req.Color, req.Active,
			now, username,
		}))
	if err != nil {
		// log.Printf("Insert failed: %v", err)
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Cases", "UpdateCaseStatus", "",
			"update", -1, start_time, GetQueryParams(c), response, "Failure : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		logger.Warn("Update failed", zap.Error(err))
		return
	}

	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Update successfully",
	}
	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, id, "Cases", "UpdateCaseStatus", "",
		"update", 0, start_time, GetQueryParams(c), response, "UpdateCaseStatus Success.",
	)
	//=======AUDIT_END=====//
	// Continue logic...
	c.JSON(http.StatusOK, response)
}

// @summary Delete Case Status
// @id Delete Case Status
// @security ApiKeyAuth
// @accept json
// @tags Cases
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/case_status/{id} [delete]
func DeleteCaseStatus(c *gin.Context) {

	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()
	orgId := GetVariableFromToken(c, "orgId")
	id := c.Param("id")
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	txtId := uuid.New().String()

	query := `DELETE FROM public."case_status" WHERE id = $1`
	logger.Debug("Query", zap.String("query", query), zap.Any("id", id))
	_, err := conn.Exec(ctx, query, id, orgId)
	if err != nil {
		// log.Printf("Insert failed: %v", err)
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Cases", "DeleteCaseStatus", "",
			"delete", -1, start_time, GetQueryParams(c), response, "Failure : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		logger.Warn("Update failed", zap.Error(err))
		return
	}
	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Delete successfully",
	}
	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, id, "Cases", "DeleteCaseStatus", "",
		"delete", 0, start_time, GetQueryParams(c), response, "Delete CaseStatus success.",
	)
	//=======AUDIT_END=====//
	// Continue logic...
	c.JSON(http.StatusOK, response)
}
