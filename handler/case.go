package handler

import (
	"encoding/json"
	"fmt"
	"mainPackage/config"
	"mainPackage/model"
	"net/http"

	"github.com/gin-gonic/gin"
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

	var caseLists []model.OutputCaseType
	var errorMsg string
	for rows.Next() {
		var cusCase model.OutputCaseType
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

// ListCase godoc
// @summary List Cases Sub Type
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
	query := `SELECT id, "typeId", "sTypeId", "orgId", en, th, "wfId", "caseSla", priority, "userSkillList", "unitPropLists",
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

	var caseLists []model.OutputCaseSubType
	var errorMsg string
	for rows.Next() {
		var cusCase model.OutputCaseSubType
		err := rows.Scan(&cusCase.Id, &cusCase.TypeID, &cusCase.STypeID, &cusCase.OrgID, &cusCase.EN, &cusCase.TH, &cusCase.WFID,
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
