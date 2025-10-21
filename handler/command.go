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

// ListCommands godoc
// @summary Get Commands
// @tags Organization
// @security ApiKeyAuth
// @id Get Commands
// @accept json
// @produce json
// @Param start query int false "start" default(0)
// @Param length query int false "length" default(10)
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/commands [get]
func GetCommand(c *gin.Context) {
	logger := utils.GetLog()

	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	startStr := c.DefaultQuery("start", "0")
	start, err := strconv.Atoi(startStr)
	if err != nil {
		start = 0
	}
	lengthStr := c.DefaultQuery("length", "1000")
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		length = 1000
	}
	orgId := GetVariableFromToken(c, "orgId")
	query := `SELECT  id,"deptId", "orgId", "commId", en, th, active, "createdAt", "updatedAt", "createdBy", "updatedBy" 
	FROM public.sec_commands WHERE "orgId"=$1 LIMIT $2 OFFSET $3`

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query))
	rows, err = conn.Query(ctx, query, orgId, length, start)
	id := c.Param("id")
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	txtId := uuid.New().String()

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
			txtId, id, "Command", "GetCommand", "",
			"search", -1, start_time, GetQueryParams(c), response, "Query failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer rows.Close()
	var errorMsg string
	var Department model.Command
	var DepartmentList []model.Command
	rowIndex := 0
	for rows.Next() {
		rowIndex++
		err := rows.Scan(&Department.ID, &Department.DeptID, &Department.OrgID, &Department.CommID, &Department.En, &Department.Th,
			&Department.Active, &Department.CreatedAt, &Department.UpdatedAt, &Department.CreatedBy, &Department.UpdatedBy)
		if err != nil {
			logger.Warn("Scan failed", zap.Error(err))
			response := model.Response{
				Status: "-1",
				Msg:    "Failed",
				Desc:   errorMsg,
			}
			//=======AUDIT_START=====//
			_ = utils.InsertAuditLogs(
				c, conn, orgId.(string), username.(string),
				txtId, id, "Command", "GetCommand", "",
				"search", -1, start_time, GetQueryParams(c), response, "Scan failed : "+err.Error(),
			)
			//=======AUDIT_END=====//
			c.JSON(http.StatusInternalServerError, response)
		}
		DepartmentList = append(DepartmentList, Department)
	}
	if errorMsg != "" {
		response := model.Response{
			Status: "-1",
			Msg:    "Failed",
			Desc:   errorMsg,
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Command", "GetCommand", "",
			"search", -1, start_time, GetQueryParams(c), response, "Failed : "+errorMsg,
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
			txtId, id, "Command", "GetCommand", "",
			"search", 0, start_time, GetQueryParams(c), response, "GetCommand Success",
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusOK, response)
	}
}

// ListCommands godoc
// @summary Get Commands by id
// @tags Organization
// @security ApiKeyAuth
// @id Get Commands by id
// @accept json
// @produce json
// @Param id path string true "id" "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/commands/{id} [get]
func GetCommandById(c *gin.Context) {
	logger := utils.GetLog()
	id := c.Param("id")
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	orgId := GetVariableFromToken(c, "orgId")
	query := `SELECT id, "deptId", "orgId", "commId", en, th, active, "createdAt", "updatedAt", "createdBy", "updatedBy" 
	FROM public.sec_commands WHERE "commId"=$1 AND "orgId"=$2`

	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	txtId := uuid.New().String()

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query))
	rows, err := conn.Query(ctx, query, id, orgId)
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
			txtId, "", "Command", "GetCommandById", "",
			"search", -1, start_time, GetQueryParams(c), response, "Query failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer rows.Close()
	var errorMsg string
	var Department model.Command
	err = rows.Scan(&Department.ID, &Department.DeptID, &Department.OrgID, &Department.CommID, &Department.En, &Department.Th,
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
			txtId, "", "Command", "GetCommandById", "",
			"search", -1, start_time, GetQueryParams(c), response, "Scan failed : "+err.Error(),
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
		txtId, "", "Command", "GetCommandById", "",
		"search", 0, start_time, GetQueryParams(c), response, "GetCommandById Success",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)

}

// @summary Create Commands
// @id Create Commands
// @security ApiKeyAuth
// @tags Organization
// @accept json
// @produce json
// @param Body body model.CommandInsert true "Create Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/commands/add [post]
func InsertCommand(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()
	start_time := time.Now()
	txtId := uuid.New().String()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")

	var req model.CommandInsert
	if err := c.ShouldBindJSON(&req); err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, "", "Command", "InsertCommand", "",
			"create", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusBadRequest, response)
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	uuid := uuid.New()
	now := time.Now()
	var id int
	query := `
	INSERT INTO public."sec_commands"(
	"deptId", "orgId", "commId", en, th, active, "createdAt", "updatedAt", "createdBy", "updatedBy")
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	RETURNING id ;
	`

	err := conn.QueryRow(ctx, query,
		req.DeptID, orgId, uuid, req.En, req.Th, req.Active, now,
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
			txtId, "", "Command", "InsertCommand", "",
			"create", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	// Continue logic...
	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Create successfully",
	}
	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, "", "Command", "InsertCommand", "",
		"create", 0, start_time, GetQueryParams(c), response, "InsertCommand Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)

}

// @summary Update Commands
// @id Update Commands
// @security ApiKeyAuth
// @accept json
// @produce json
// @tags Organization
// @Param id path string true "id"
// @param Body body model.CommandUpdate true "Update data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/commands/{id} [patch]
func UpdateCommand(c *gin.Context) {
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

	var req model.CommandUpdate
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
			txtId, id, "Command", "UpdateCommand", "",
			"update", -1, start_time, GetQueryParams(c), response, "Update failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusBadRequest, response)
		return
	}
	now := time.Now()
	query := `UPDATE public."sec_commands"
	SET "deptId"=$2, "orgId"=$3, en=$4, th=$5, active=$6,
	 "updatedAt"=$7, "updatedBy"=$8
	WHERE "commId" = $1 AND "orgId"=$9`
	_, err := conn.Exec(ctx, query,
		id, req.DeptID, orgId, req.En, req.Th, req.Active,
		now, username, orgId,
	)
	logger.Debug("Update Case SQL Args",
		zap.String("query", query),
		zap.Any("Input", []any{
			id, req.DeptID, orgId, req.En, req.Th, req.Active,
			now, username, orgId,
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
			txtId, id, "Command", "UpdateCommand", "",
			"update", -1, start_time, GetQueryParams(c), response, "Failure : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		logger.Warn("Update failed", zap.Error(err))
		return
	}

	// Continue logic...
	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Update successfully",
	}
	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, id, "Command", "UpdateCommand", "",
		"update", 0, start_time, GetQueryParams(c), response, "UpdateCommand Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)
}

// @summary Delete Commands
// @id Delete Commands
// @security ApiKeyAuth
// @accept json
// @tags Organization
// @produce json
// @Param id path string true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/commands/{id} [delete]
func DeleteCommand(c *gin.Context) {

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

	query := `DELETE FROM public."sec_commands" WHERE "commId" = $1 AND "orgId"=$2`
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
			txtId, id, "Command", "DeleteCommand", "",
			"delete", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		logger.Warn("Update failed", zap.Error(err))
		return
	}

	// Continue logic...
	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Delete successfully",
	}
	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, id, "Command", "DeleteCommand", "",
		"delete", 0, start_time, GetQueryParams(c), response, "DeleteCommand Success",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)
}
