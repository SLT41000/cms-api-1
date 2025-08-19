package handler

import (
	"mainPackage/config"
	"mainPackage/model"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
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
	logger := config.GetLog()

	conn, ctx, cancel := config.ConnectDB()
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
	query := `SELECT id, "orgId", "caseId", username, type, "fullMsg", "jsonData", "createdAt", "createdBy"
	FROM public.tix_case_history_events WHERE "orgId"=$1 LIMIT $2 OFFSET $3`

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query))
	rows, err = conn.Query(ctx, query, orgId, length, start)
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
		c.JSON(http.StatusInternalServerError, response)
	} else {
		response := model.Response{
			Status: "0",
			Msg:    "Success",
			Data:   CaseHistoryList,
			Desc:   "",
		}
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
	logger := config.GetLog()

	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)

	caseId := c.Param("caseId")
	orgId := GetVariableFromToken(c, "orgId")

	query := `SELECT id, "orgId", "caseId", username, type, "fullMsg", "jsonData", "createdAt", "createdBy"
	FROM public.tix_case_history_events WHERE "orgId"=$1 AND "caseId"=$2`

	logger.Debug("Query", zap.String("query", query))

	rows, err := conn.Query(ctx, query, orgId, caseId)
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

	var CaseHistory model.CaseHistory
	var CaseHistoryList []model.CaseHistory

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
		c.JSON(http.StatusOK, model.Response{
			Status: "-1",
			Msg:    "NoData",
			Desc:   "",
		})
		return
	}

	// ✅ Otherwise return success
	c.JSON(http.StatusOK, model.Response{
		Status: "0",
		Msg:    "Success",
		Data:   CaseHistoryList,
		Desc:   "",
	})
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
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)

	var req model.CaseHistoryInsert
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
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	// Continue logic...
	c.JSON(http.StatusOK, model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Create successfully",
	})

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
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)

	id := c.Param("id")

	var req model.CaseHistoryUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Update failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
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
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Update failed", zap.Error(err))
		return
	}

	// Continue logic...
	c.JSON(http.StatusOK, model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Update successfully",
	})
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

	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	orgId := GetVariableFromToken(c, "orgId")
	id := c.Param("id")
	query := `DELETE FROM public.tix_case_history_events WHERE id = $1 AND "orgId"=$2`
	logger.Debug("Query", zap.String("query", query), zap.Any("id", id))
	_, err := conn.Exec(ctx, query, id, orgId)
	if err != nil {
		// log.Printf("Insert failed: %v", err)
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Update failed", zap.Error(err))
		return
	}

	// Continue logic...
	c.JSON(http.StatusOK, model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Delete successfully",
	})
}
