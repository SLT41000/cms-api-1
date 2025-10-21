package handler

import (
	"encoding/json"
	"log"
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

// @summary Get Case History
// @tags Cases
// @security ApiKeyAuth
// @id Get Case History
// @accept json
// @produce json
// @Param start query int false "start" default(0)
// @Param length query int false "length" default(10)
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/case_history [get]
func GetCaseHistory(c *gin.Context) {
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
	id := c.Param("id")
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	txtId := uuid.New().String()

	query := `SELECT id, "orgId", "caseId", username, type, "fullMsg", "jsonData", "createdAt", "createdBy"
	FROM public.tix_case_history_events WHERE "orgId"=$1 LIMIT $2 OFFSET $3`

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query))
	rows, err = conn.Query(ctx, query, orgId, length, start)
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
			txtId, id, "Cases", "GetCaseHistory", "",
			"search", -1, start_time, GetQueryParams(c), response, "Query failed = "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer rows.Close()
	var errorMsg string
	var CaseHistory model.CaseHistory
	var CaseHistoryList []model.CaseHistory
	found := false
	for rows.Next() {
		err := rows.Scan(
			&CaseHistory.ID,
			&CaseHistory.OrgID,
			&CaseHistory.CaseID,
			&CaseHistory.Username,
			&CaseHistory.Type,
			&CaseHistory.FullMsg,
			&CaseHistory.JSONData,
			&CaseHistory.CreatedAt,
			&CaseHistory.CreatedBy,
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
				txtId, id, "Cases", "GetCaseHistory", "",
				"search", -1, start_time, GetQueryParams(c), response, "Scan failed = "+err.Error(),
			)
			//=======AUDIT_END=====//
			c.JSON(http.StatusInternalServerError, response)
			return
		}
		CaseHistoryList = append(CaseHistoryList, CaseHistory)
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
			txtId, id, "Cases", "GetCaseHistory", "",
			"search", -1, start_time, GetQueryParams(c), response, "Not Found.",
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
	} else {
		response := model.Response{
			Status: "0",
			Msg:    "Success",
			Data:   CaseHistoryList,
			Desc:   "",
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Cases", "GetCaseHistory", "",
			"search", 0, start_time, GetQueryParams(c), response, "GetCaseHistory Success.",
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusOK, response)
	}
}

// @summary Get Case History By Case Id
// @tags Cases
// @security ApiKeyAuth
// @id Get Case History By Case Id
// @accept json
// @produce json
// @Param caseId path string true "caseId"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/case_history/{caseId} [get]
func GetCaseHistoryByCaseId(c *gin.Context) {
	logger := utils.GetLog()

	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)

	caseId := c.Param("caseId")
	orgId := GetVariableFromToken(c, "orgId")
	id := c.Param("id")
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	txtId := uuid.New().String()
	query := `SELECT id, type, "fullMsg","createdAt", "createdBy"
	FROM public.tix_case_history_events WHERE "orgId"=$1 AND "caseId"=$2`

	logger.Debug("Query", zap.String("query", query))

	rows, err := conn.Query(ctx, query, orgId, caseId)
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
			txtId, id, "Cases", "GetCaseHistoryCaseId", "",
			"search", -1, start_time, GetQueryParams(c), response, "Query failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer rows.Close()

	var CaseHistory model.CaseHistory
	var CaseHistoryList []model.CaseHistory

	for rows.Next() {
		err := rows.Scan(
			&CaseHistory.ID,
			&CaseHistory.Type,
			&CaseHistory.FullMsg,
			&CaseHistory.CreatedAt,
			&CaseHistory.CreatedBy,
		)
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
				txtId, id, "Cases", "GetCaseHistoryCaseId", "",
				"search", -1, start_time, GetQueryParams(c), response, "Scan failed : "+err.Error(),
			)
			//=======AUDIT_END=====//
			c.JSON(http.StatusInternalServerError, model.Response{
				Status: "-1",
				Msg:    "Failed",
				Desc:   err.Error(),
			})
			return
		}
		CaseHistoryList = append(CaseHistoryList, CaseHistory)
	}

	// ✅ If no rows found, return 200 with "NoData"
	if len(CaseHistoryList) == 0 {
		response := model.Response{
			Status: "-1",
			Msg:    "NoData",
			Desc:   "",
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Cases", "GetCaseHistoryCaseId", "",
			"search", -1, start_time, GetQueryParams(c), response, "Not Found.",
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusOK, response)
		return
	}
	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Data:   CaseHistoryList,
		Desc:   "",
	}
	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, id, "Cases", "GetCaseHistoryCaseId", "",
		"search", 0, start_time, GetQueryParams(c), response, "GetCaseHistoryByCaseId Success.",
	)
	//=======AUDIT_END=====//
	// ✅ Otherwise return success
	c.JSON(http.StatusOK, response)
}

// @summary Create Case History
// @id Create Case History
// @security ApiKeyAuth
// @tags Cases
// @accept json
// @produce json
// @param Body body model.CaseHistoryInsert true "Create Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/case_history/add [post]
func InsertCaseHistory(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	txtId := uuid.New().String()
	start_time := time.Now()
	var req model.CaseHistoryInsert
	if err := c.ShouldBindJSON(&req); err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, "", "",
			txtId, "", "Cases", "InsertCaseHistory", "",
			"create", -1, start_time, GetQueryParams(c), response, "Failure : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusBadRequest, response)
		logger.Warn("Insert failed", zap.Error(err))
		return
	}
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	now := time.Now()
	var id int
	query := `
	INSERT INTO public.tix_case_history_events(
	"orgId", "caseId", username, type, "fullMsg", "jsonData", "createdAt", "createdBy")
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	RETURNING id ;
	`

	err := conn.QueryRow(ctx, query,
		orgId, req.CaseID, username, "comment", req.FullMsg, req.JSONData,
		now, username).Scan(&id)

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
			txtId, strconv.Itoa(id), "Cases", "InsertCaseHistory", "",
			"create", -1, start_time, GetQueryParams(c), response, "Failure : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	provID, wfId, versions, err := GetInfoFromCase(ctx, conn, orgId.(string), req.CaseID)
	if err != nil {
		log.Print("GetInfoFromCase Error :", err.Error())
	}
	log.Print(provID, wfId, versions)

	recipients := []model.Recipient{
		{Type: "provId", Value: provID},
	}

	msg := map[string]interface{}{
		"caseId":    req.CaseID,
		"fullMsg":   req.FullMsg,
		"jsonData":  req.JSONData,
		"orgId":     orgId.(string),
		"username":  req.Username,
		"createdAt": now,
		"type":      req.Type,
	}
	additionalJsonMap := msg
	additionalJSON, err := json.Marshal(additionalJsonMap)
	if err != nil {
		log.Printf("covent additionalData Error :", err)
	}
	additionalData := json.RawMessage(additionalJSON)
	event := "CASE-HISTORY"
	genNotiCustom(c, conn, orgId.(string), username.(string), username.(string), "/case/"+req.CaseID, "hidden", nil, "", recipients, "/case/"+req.CaseID, "User", event, &additionalData)
	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Create successfully",
	}
	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, strconv.Itoa(id), "Cases", "InsertCaseHistory", "",
		"create", 0, start_time, GetQueryParams(c), response, "InsertCaseHistory Success.",
	)
	//=======AUDIT_END=====//
	// Continue logic...
	c.JSON(http.StatusOK, response)

}

// @summary Update Case History
// @id Update Case History
// @security ApiKeyAuth
// @accept json
// @produce json
// @tags Cases
// @Param id path int true "id"
// @param Body body model.CaseHistoryUpdate true "Update data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/case_history/{id} [patch]
func UpdateCaseHistory(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)

	id := c.Param("id")
	start_time := time.Now()
	txtId := uuid.New().String()

	var req model.CaseHistoryUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Update failed", zap.Error(err))
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, "", "",
			txtId, "", "Cases", "UpdateCaseHistory", "",
			"update", -1, start_time, GetQueryParams(c), response, "Failure = "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusBadRequest, response)
		return
	}
	now := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	query := `UPDATE public."tix_case_history_events"
	SET "username"=$3, type=$4, "fullMsg"=$5, "jsonData"=$6,
	WHERE id = $1 AND "orgId"=$2`
	_, err := conn.Exec(ctx, query,
		id, orgId, username, req.Type, req.FullMsg, req.JSONData, now, username,
	)
	logger.Debug("Update Case SQL Args",
		zap.String("query", query),
		zap.Any("Input", []any{
			id, orgId, username, req.Type, req.FullMsg, req.JSONData, now, username,
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
			txtId, "", "Cases", "UpdateCaseHistory", "",
			"update", -1, start_time, GetQueryParams(c), response, "Failure = "+err.Error(),
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
		txtId, "", "Cases", "UpdateCaseHistory", "",
		"update", 0, start_time, GetQueryParams(c), response, "UpdateCaseHistory Success.",
	)
	//=======AUDIT_END=====//
	// Continue logic...
	c.JSON(http.StatusOK, response)
}

// @summary Delete Case History
// @id Delete Case History
// @security ApiKeyAuth
// @accept json
// @tags Cases
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/case_history/{id} [delete]
func DeleteCaseHistory(c *gin.Context) {

	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	orgId := GetVariableFromToken(c, "orgId")
	id := c.Param("id")
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	txtId := uuid.New().String()
	query := `DELETE FROM public.tix_case_history_events WHERE id = $1 AND "orgId"=$2`
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
			txtId, id, "Cases", "DeleteCaseHistory", "",
			"delete", -1, start_time, GetQueryParams(c), response, "Query Fail : "+err.Error(),
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
		txtId, id, "Cases", "DeleteCaseHistory", "",
		"delete", -1, start_time, GetQueryParams(c), response, "DeleteCaseHistory Success.",
	)
	//=======AUDIT_END=====//
	// Continue logic...
	c.JSON(http.StatusOK, response)
}
