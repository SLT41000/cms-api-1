package handler

import (
	"encoding/json"
	"fmt"
	"mainPackage/config"
	"mainPackage/model"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func Process(function string, key string, status string, input interface{}, output interface{}) string {
	// Marshal input to JSON
	inputJSON, err := json.Marshal(input)
	if err != nil {
		inputJSON = []byte(`"marshal input error"`)
	}

	// Marshal output to JSON
	outputJSON, err := json.Marshal(output)
	if err != nil {
		outputJSON = []byte(`"marshal output error"`)
	}

	// Format final log string
	return fmt.Sprintf("[%s][%s][%s][%s][%s]", function, key, status, string(inputJSON), string(outputJSON))
}

// func genCaseID() string {
// 	currentTime := time.Now()
// 	year := currentTime.Format("06")                                 // "25" for 2025
// 	month := fmt.Sprintf("%02d", int(currentTime.Month()))           // "06" for June
// 	day := fmt.Sprintf("%02d", currentTime.Day())                    // "10" for 10th
// 	hour := fmt.Sprintf("%02d", currentTime.Hour())                  // "15" for 3 PM
// 	minute := fmt.Sprintf("%02d", currentTime.Minute())              // "04"
// 	second := fmt.Sprintf("%02d", currentTime.Second())              // "05"
// 	millisecond := fmt.Sprintf("%07d", currentTime.Nanosecond()/1e3) // "1234567" (nanoseconds → microseconds)

// 	// Combine into DYYMMDDHHMMSSNNNNNNN format
// 	timestamp := fmt.Sprintf("D%s%s%s%s%s%s%s",
// 		year,
// 		month,
// 		day,
// 		hour,
// 		minute,
// 		second,
// 		millisecond,
// 	)
// 	return timestamp
// }

// ListCase godoc
// @summary List Cases
// @tags Cases
// @security ApiKeyAuth
// @id ListCaseTypes
// @accept json
// @produce json
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/casetypes [get]
func ListCaseType(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	orgId := GetVariableFromToken(c, "orgId")

	query := `SELECT id,"typeId", "orgId", en, th, active, "createdAt", "updatedAt", "createdBy", "updatedBy" 
	FROM public.case_types WHERE "orgId"=$1 LIMIT 1000`
	logger.Debug(`Query`, zap.String("query", query))

	rows, err := conn.Query(ctx, query, orgId)
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

	var caseLists []model.CaseType
	var errorMsg string
	for rows.Next() {
		var cusCase model.CaseType
		err := rows.Scan(&cusCase.Id, &cusCase.TypeId, &cusCase.OrgId, &cusCase.En, &cusCase.Th, &cusCase.Active, &cusCase.CreatedAt,
			&cusCase.UpdatedAt, &cusCase.CreatedBy, &cusCase.UpdatedBy)
		if err != nil {
			logger.Warn("Query failed", zap.Error(err))
			errorMsg = err.Error()
			continue
		}

		caseLists = append(caseLists, cusCase)
	}

	// Final JSON
	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Data:   caseLists,
		Desc:   errorMsg,
	}
	c.JSON(http.StatusOK, response)

	paramQuery := c.Request.URL.RawQuery
	logStr := Process("ListCase", paramQuery, response.Status, paramQuery, response)
	logger.Info(logStr)
}

// @summary Create CaseType
// @id Create CaseType
// @security ApiKeyAuth
// @tags Cases
// @accept json
// @produce json
// @param Body body model.CaseTypeInsert true "Create Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/casetypes/add [post]
func InsertCaseType(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	var req model.CaseTypeInsert
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
	now := time.Now()
	var id int
	orgId := GetVariableFromToken(c, "orgId")
	uuid := uuid.New()
	query := `
	INSERT INTO public."case_types"(
	"typeId", "orgId", en, th, active, "createdAt", "updatedAt", "createdBy", "updatedBy")
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING id ;
	`

	err := conn.QueryRow(ctx, query,
		uuid, orgId, req.En, req.Th, req.Active, now,
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

	// Continue logic...
	c.JSON(http.StatusOK, model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Create successfully",
	})

}

// @summary Update CaseType
// @id Update CaseType
// @security ApiKeyAuth
// @accept json
// @produce json
// @tags Cases
// @Param id path int true "id"
// @param Body body model.CaseTypeUpdate true "Update data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/casetypes/{id} [patch]
func UpdateCaseType(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	id := c.Param("id")

	var req model.CaseTypeUpdate
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
	query := `UPDATE public."case_types"
	SET en=$2, th=$3, active=$4,
	 "updatedAt"=$5, "updatedBy"=$6
	WHERE id = $1 AND "orgId"=$7`
	_, err := conn.Exec(ctx, query,
		id, req.En, req.Th, req.Active,
		now, username, orgId,
	)
	logger.Debug("Update Case SQL Args",
		zap.String("query", query),
		zap.Any("Input", []any{
			id, req.En, req.Th, req.Active,
			now, username, orgId,
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

// @summary Delete CaseType
// @id Delete CaseType
// @security ApiKeyAuth
// @accept json
// @tags Cases
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/casetypes/{id} [delete]
func DeleteCaseType(c *gin.Context) {

	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()
	orgId := GetVariableFromToken(c, "orgId")
	id := c.Param("id")
	query := `DELETE FROM public."case_types" WHERE id = $1 AND "orgId"=$2`
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

// ListCase godoc
// @summary List CasesSubType
// @tags Cases
// @security ApiKeyAuth
// @id ListCaseSubTypes
// @accept json
// @produce json
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/casesubtypes [get]
func ListCaseSubType(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	orgId := GetVariableFromToken(c, "orgId")
	query := `SELECT id, "typeId", "sTypeId", "sTypeCode", "orgId", en, th, "wfId", "caseSla", priority, "userSkillList", "unitPropLists",
	 active, "createdAt", "updatedAt", "createdBy", "updatedBy" FROM public.case_sub_types WHERE "orgId"=$1`
	logger.Debug(`Query`, zap.String("query", query))

	rows, err := conn.Query(ctx, query, orgId)
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

	var caseLists []model.CaseSubType
	var errorMsg string
	for rows.Next() {
		var cusCase model.CaseSubType
		err := rows.Scan(&cusCase.Id, &cusCase.TypeID, &cusCase.STypeID, &cusCase.STypeCode, &cusCase.OrgID, &cusCase.EN, &cusCase.TH, &cusCase.WFID,
			&cusCase.CaseSLA, &cusCase.Priority, &cusCase.UserSkillList, &cusCase.UnitPropLists, &cusCase.Active,
			&cusCase.CreatedAt, &cusCase.UpdatedAt, &cusCase.CreatedBy, &cusCase.UpdatedBy)
		if err != nil {
			logger.Warn("Query failed", zap.Error(err))
			errorMsg = err.Error()
			continue
		}

		caseLists = append(caseLists, cusCase)
	}

	// Final JSON
	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Data:   caseLists,
		Desc:   errorMsg,
	}
	c.JSON(http.StatusOK, response)

	paramQuery := c.Request.URL.RawQuery
	logStr := Process("ListCase", paramQuery, response.Status, paramQuery, response)
	logger.Info(logStr)
}

// @summary Create CaseSubType
// @id Create CaseSubType
// @security ApiKeyAuth
// @tags Cases
// @accept json
// @produce json
// @param Body body model.CaseSubTypeInsert true "Create Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/casesubtypes/add [post]
func InsertCaseSubType(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	var req model.CaseSubTypeInsert
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
	now := time.Now()
	var id int
	orgId := GetVariableFromToken(c, "orgId")
	uuid := uuid.New()
	query := `
	INSERT INTO public."case_sub_types"(
	"typeId", "sTypeId", "sTypeCode", "orgId", en, th, "wfId", "caseSla", priority,
	 "userSkillList", "unitPropLists", active, "createdAt", "updatedAt", "createdBy", "updatedBy")
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING id ;
	`

	err := conn.QueryRow(ctx, query, req.TypeID, uuid, req.STypeCode, orgId, req.EN, req.TH,
		req.WFID, req.CaseSLA, req.Priority, req.UserSkillList, req.UnitPropLists, req.Active, now,
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

	// Continue logic...
	c.JSON(http.StatusOK, model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Create successfully",
	})

}

// @summary Update CaseSubType
// @id Update CaseSubType
// @security ApiKeyAuth
// @accept json
// @produce json
// @tags Cases
// @Param id path int true "id"
// @param Body body model.CaseSubTypeUpdate true "Update data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/casesubtypes/{id} [patch]
func UpdateCaseSubType(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	id := c.Param("id")

	var req model.CaseSubTypeUpdate
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
	query := `UPDATE public."case_sub_types"
	SET "sTypeCode"=$3, en=$4, th=$5, "wfId"=$6, "caseSla"=$7,
	 priority=$8, "userSkillList"=$9, "unitPropLists"=$10, active=$11, "updatedAt"=$12,
	  "updatedBy"=$13
	WHERE id = $1 AND "orgId"=$2`
	_, err := conn.Exec(ctx, query,
		id, orgId, req.STypeCode, req.EN, req.TH, req.WFID, req.CaseSLA, req.Priority, req.UserSkillList, req.UnitPropLists, req.Active,
		now, username,
	)
	logger.Debug("Update Case SQL Args",
		zap.String("query", query),
		zap.Any("Input", []any{
			id, orgId, req.STypeCode, req.EN, req.TH, req.WFID, req.CaseSLA, req.Priority, req.UserSkillList, req.UnitPropLists, req.Active,
			now, username,
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

// @summary Delete CaseSubType
// @id Delete CaseSubType
// @security ApiKeyAuth
// @accept json
// @tags Cases
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/casesubtypes/{id} [delete]
func DeleteCaseSubType(c *gin.Context) {

	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()
	orgId := GetVariableFromToken(c, "orgId")
	id := c.Param("id")
	query := `DELETE FROM public."case_sub_types" WHERE id = $1 AND "orgId"=$2`
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
