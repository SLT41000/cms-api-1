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

// @summary Get Audit Log
// @tags Audit Log
// @security ApiKeyAuth
// @id Get Audit Log
// @accept json
// @produce json
// @Param start query int false "start" default(0)
// @Param length query int false "length" default(10)
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/audit_log [get]
func GetAuditlog(c *gin.Context) {
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
	query := `SELECT id, "orgId", username, "txId", "uniqueId", "mainFunc", "subFunc", "nameFunc", action, status, duration, "newData", "oldData", "resData", message, "createdAt"
	FROM public.audit_logs ORDER BY "createdAt" DESC LIMIT $1 OFFSET $2 FOR SHARE SKIP LOCKED`

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query))
	rows, err = conn.Query(ctx, query, length, start)
	if err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Audit Log", "GetAuditLog", "",
			"search", -1, start_time, GetQueryParams(c), response, "Query failed = "+err.Error(),
		)
		//=======AUDIT_END=====//
		logger.Warn("Query failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer rows.Close()
	var errorMsg string
	var AuditLog model.AuditLog
	var AuditLogList []model.AuditLog
	found := false
	for rows.Next() {

		err := rows.Scan(
			&AuditLog.ID,
			&AuditLog.OrgID,
			&AuditLog.Username,
			&AuditLog.TxID,
			&AuditLog.UniqueId,
			&AuditLog.MainFunc,
			&AuditLog.SubFunc,
			&AuditLog.NameFunc,
			&AuditLog.Action,
			&AuditLog.Status,
			&AuditLog.Duration,
			&AuditLog.NewData,
			&AuditLog.OldData,
			&AuditLog.ResData,
			&AuditLog.Message,
			&AuditLog.CreatedAt,
		)
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
				txtId, id, "Audit Log", "GetAuditLog", "",
				"search", -1, start_time, GetQueryParams(c), response, "Scan failed = "+err.Error(),
			)
			//=======AUDIT_END=====//
			c.JSON(http.StatusInternalServerError, response)
			return
		}
		AuditLogList = append(AuditLogList, AuditLog)
		found = true
	}
	if !found {
		response := model.Response{
			Status: "-1",
			Msg:    "Failed",
			Desc:   errorMsg,
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Audit Log", "GetAuditLog", "",
			"search", -1, start_time, GetQueryParams(c), response, "Not Found",
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
	} else {
		response := model.Response{
			Status: "0",
			Msg:    "Success",
			Data:   AuditLogList,
			Desc:   "",
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Audit Log", "GetAuditLog", "",
			"search", 0, start_time, GetQueryParams(c), response, "Success",
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusOK, response)
	}
}

// @summary Get Audit Log By Username
// @tags Audit Log
// @security ApiKeyAuth
// @id Get Audit Log By Username
// @accept json
// @produce json
// @Param username path string true "username"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/audit_log/{username} [get]
func GetAuditlogByUsername(c *gin.Context) {
	logger := utils.GetLog()

	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	orgId := GetVariableFromToken(c, "orgId")
	username := c.Param("username")
	id := c.Param("id")
	start_time := time.Now()
	txtId := uuid.New().String()
	query := `SELECT id, "orgId", username, "txId", "uniqueId", "mainFunc", "subFunc", "nameFunc", action, status, duration, "newData", "oldData", "resData", message, "createdAt"
	FROM public.audit_logs WHERE "orgId"=$1 AND username=$2`
	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query))
	rows, err := conn.Query(ctx, query, orgId, username)
	if err != nil {
		logger.Warn("Query failed", zap.Error(err))
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username,
			txtId, id, "Audit Log", "GetAuditlogByUsername", "",
			"search", -1, start_time, GetQueryParams(c), response, "Failure : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer rows.Close()
	var errorMsg string
	var AuditLog model.AuditLog
	var AuditLogList []model.AuditLog
	found := false
	for rows.Next() {

		err := rows.Scan(
			&AuditLog.ID,
			&AuditLog.OrgID,
			&AuditLog.Username,
			&AuditLog.TxID,
			&AuditLog.UniqueId,
			&AuditLog.MainFunc,
			&AuditLog.SubFunc,
			&AuditLog.NameFunc,
			&AuditLog.Action,
			&AuditLog.Status,
			&AuditLog.Duration,
			&AuditLog.NewData,
			&AuditLog.OldData,
			&AuditLog.ResData,
			&AuditLog.Message,
			&AuditLog.CreatedAt,
		)
		if err != nil {
			logger.Warn("Scan failed", zap.Error(err))
			response := model.Response{
				Status: "-1",
				Msg:    "Failed",
				Desc:   errorMsg,
			}
			//=======AUDIT_START=====//
			_ = utils.InsertAuditLogs(
				c, conn, orgId.(string), username,
				txtId, id, "Audit Log", "GetAuditlogByUsername", "",
				"search", -1, start_time, GetQueryParams(c), response, "Scan failed : "+err.Error(),
			)
			//=======AUDIT_END=====//
			c.JSON(http.StatusInternalServerError, response)
			return
		}
		AuditLogList = append(AuditLogList, AuditLog)
		found = true
	}
	if !found {
		response := model.Response{
			Status: "-1",
			Msg:    "Failed",
			Desc:   errorMsg,
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username,
			txtId, id, "Audit Log", "GetAuditlogByUsername", "",
			"search", -1, start_time, GetQueryParams(c), response, "Not Found.",
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
	} else {
		response := model.Response{
			Status: "0",
			Msg:    "Success",
			Data:   AuditLogList,
			Desc:   "",
		}
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username,
			txtId, id, "Audit Log", "GetAuditlogByUsername", "",
			"search", 0, start_time, GetQueryParams(c), response, "GetAuditlog By Username Success.",
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusOK, response)
	}
}
