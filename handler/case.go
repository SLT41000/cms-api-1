package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mainPackage/config"
	"mainPackage/model"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
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

func genCaseID() string {
	currentTime := time.Now()
	year := currentTime.Format("06")                                 // "25" for 2025
	month := fmt.Sprintf("%02d", int(currentTime.Month()))           // "06" for June
	day := fmt.Sprintf("%02d", currentTime.Day())                    // "10" for 10th
	hour := fmt.Sprintf("%02d", currentTime.Hour())                  // "15" for 3 PM
	minute := fmt.Sprintf("%02d", currentTime.Minute())              // "04"
	second := fmt.Sprintf("%02d", currentTime.Second())              // "05"
	millisecond := fmt.Sprintf("%07d", currentTime.Nanosecond()/1e3) // "1234567" (nanoseconds → microseconds)

	// Combine into DYYMMDDHHMMSSNNNNNNN format
	timestamp := fmt.Sprintf("D%s%s%s%s%s%s%s",
		year,
		month,
		day,
		hour,
		minute,
		second,
		millisecond,
	)
	return timestamp
}

func Convert(list map[string]string, value *string) *string {
	if value != nil {
		if val, ok := list[*value]; ok {
			return &val
		}
	}
	return nil
}

func GetStationMap(ctx context.Context, db *pgx.Conn) (map[string]string, error) {
	stationMap := make(map[string]string)

	rows, err := db.Query(ctx, `SELECT policestationcode, policestationname FROM public.police_station`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var code, name string
		if err := rows.Scan(&code, &name); err != nil {
			log.Printf("Row scan failed: %v", err)
			continue
		}
		stationMap[code] = name
	}

	return stationMap, nil
}

func GetCaseTypeMap(ctx context.Context, db *pgx.Conn) (map[string]string, error) {
	casetypeMap := make(map[string]string)
	casetypes, err := db.Query(ctx, `SELECT casetypecode, casetypename FROM public.case_type`)
	if err != nil {
		return nil, err
	}
	defer casetypes.Close()

	for casetypes.Next() {
		var code, name string
		err := casetypes.Scan(&code, &name) // use =, not :=
		if err != nil {
			log.Printf("Row scan failed: %v", err)
			continue
		}
		casetypeMap[code] = name
	}

	return casetypeMap, nil
}

func GetCaseStatusMap(ctx context.Context, db *pgx.Conn) (map[string]string, error) {
	caseStatusMap := make(map[string]string)

	caseStatus, err := db.Query(ctx, `SELECT casestatuscode, casestatusname FROM public.case_status`)
	if err != nil {
		return nil, err
	}
	defer caseStatus.Close()

	for caseStatus.Next() {
		var code, name string
		err := caseStatus.Scan(&code, &name) // use =, not :=
		if err != nil {
			log.Printf("Row scan failed: %v", err)
			continue
		}
		caseStatusMap[code] = name
	}

	return caseStatusMap, nil
}

// ListCase godoc
// @summary List Cases
// @tags Cases
// @security ApiKeyAuth
// @id ListCases
// @accept json
// @produce json
// @Param start query int false "start" default(0)
// @Param length query int false "length" default(10)
// @Param keyword query string false "keyword"
// @response 200 {object} model.CaseListData "OK - Request successful"
// @Router /api/v1/cases [get]
func ListCase(c *gin.Context) {
	logger := config.GetLog()
	keyword := c.Query("keyword")
	startStr := c.DefaultQuery("start", "0")
	lengthStr := c.DefaultQuery("length", "0")
	userID, exists := c.Get("username")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	logger.Debug(userID.(string))
	start, err := strconv.Atoi(startStr)
	if err != nil {
		start = 1 // fallback default
	}

	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		length = 10 // fallback default
	}
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	//---

	caseStatusMap, err := GetCaseStatusMap(ctx, conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query failed",
			"message": err.Error()})
		logger.Warn("Query failed", zap.Error(err))
		return
	}
	//---
	stationMap, err := GetStationMap(ctx, conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query failed",
			"message": err.Error()})
		logger.Warn("Query failed", zap.Error(err))
		return
	}

	//-------
	casetypeMap, err := GetCaseTypeMap(ctx, conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query failed",
			"message": err.Error()})
		logger.Warn("Query failed", zap.Error(err))
		return
	}
	//----

	query := `SELECT id, caseid,casetypecode,priority,phonenumber,casestatuscode ,
			resultcode, casedetail,policestationcode,caselocationaddress,caselocationdetail,
			specialemergency,urgentamount,createddate,casetypecode,mediatype,vowner,vvin
			FROM public."case" WHERE caseid ILIKE '%' || $3 || '%' LIMIT $1 OFFSET $2`
	logger.Debug(`Query`,
		zap.String("query", query),
		zap.Any("Input", []any{
			length, start, keyword,
		}))

	rows, err := conn.Query(ctx, query, length, start, keyword)
	if err != nil {
		logger.Warn("Query failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.CaseListResponse{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		return
	}
	defer rows.Close()

	var caseLists []model.OutputCase
	var errorMsg string
	for rows.Next() {
		var cusCase model.OutputCase
		err := rows.Scan(&cusCase.Id, &cusCase.CaseId, &cusCase.Casetype_code, &cusCase.Priority, &cusCase.Phone_number,
			&cusCase.Case_status_code, &cusCase.Case_status_name, &cusCase.Case_detail, &cusCase.Police_station_name,
			&cusCase.Case_location_address, &cusCase.Case_location_detail, &cusCase.Special_emergency, &cusCase.Urgent_amount,
			&cusCase.Created_date, &cusCase.Casetype_name, &cusCase.Media_type, &cusCase.VOwner, &cusCase.VVin)
		if err != nil {
			logger.Warn("Query failed", zap.Error(err))
			errorMsg = err.Error()
			continue
		}

		jsonData, _ := json.MarshalIndent(casetypeMap, "", "  ")
		logger.Debug("casetypeMap contents",
			zap.String("data", string(jsonData)), // Correct: wraps string as zap.Field
		)
		logger.Debug(*cusCase.Casetype_code)
		jsonData, _ = json.MarshalIndent(stationMap, "", "  ")
		logger.Debug("stationMap contents",
			zap.String("data", string(jsonData)), // Correct: wraps string as zap.Field
		)
		logger.Debug(*cusCase.Police_station_name)
		jsonData, _ = json.MarshalIndent(caseStatusMap, "", "  ")
		logger.Debug("caseStatusMap contents",
			zap.String("data", string(jsonData)), // Correct: wraps string as zap.Field
		)
		logger.Debug(*cusCase.Case_status_code)

		cusCase.Case_status_name = Convert(caseStatusMap, cusCase.Case_status_code)
		cusCase.Casetype_name = Convert(casetypeMap, cusCase.Casetype_code)
		cusCase.Police_station_name = Convert(stationMap, cusCase.Police_station_name)

		caseLists = append(caseLists, cusCase)
	}

	// Total count (for frontend pagination)
	var totalCount int
	err = conn.QueryRow(ctx, `SELECT COUNT(*) FROM public."case"`).Scan(&totalCount)
	if err != nil {
		logger.Warn("Query failed", zap.Error(err))
		totalCount = 0
	}

	// Final JSON
	response := model.CaseListResponse{
		Status: "0",
		Msg:    "Success",
		Data: model.CaseListData{
			Draw:            start,
			RecordsTotal:    totalCount,
			RecordsFiltered: length,
			Data:            caseLists,
			Error:           errorMsg,
		},
		Desc: "",
	}
	c.JSON(http.StatusOK, response)

	paramQuery := c.Request.URL.RawQuery
	logStr := Process("ListCase", paramQuery, response.Status, paramQuery, response)
	logger.Info(logStr)
}

// @summary Search Case
// @security ApiKeyAuth
// @id SearchCase
// @tags Cases
// @accept json
// @produce json
// @Param start query int false "start" default(0)
// @Param length query int false "length" default(10)
// @Param keyword query string false "keyword"
// @Param css query string false "case_status_code" default("-1")
// @Param cid query string false "case_id"
// @Param cdl query string false "case_detail"
// @Param cmc query string false "command_code"
// @Param stc query string false "station_code"
// @Param fdt query string false "opened_date (date start)"
// @Param tdt query string false "date_to (date end)"
// @Param uce query string false "user_create"
// @Param ctc query string false "casetype_code"
// @Param odb query string false "order_by" default("5-DESC")
// @response 200 {object} model.CaseListData "OK - Request successful"
// @Router /api/v1/cases/search [get]
func SearchCase(c *gin.Context) {
	logger := config.GetLog()
	keyword := c.Query("keyword")
	start := c.DefaultQuery("start", "0")
	length := c.DefaultQuery("length", "0")
	case_status_code := c.Query("css")
	case_id := c.Query("cid")
	case_detail := c.Query("cdl")
	command_code := c.Query("cmc")
	station_code := c.Query("stc")
	opened_date := c.Query("fdt")
	date_to := c.Query("tdt")
	user_create := c.Query("uce")
	casetype_code := c.Query("ctc")
	order_by := c.Query("odb")
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)

	caseStatusMap, err := GetCaseStatusMap(ctx, conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query failed",
			"message": err.Error()})
		logger.Warn("Query failed", zap.Error(err))
		return
	}
	//---
	stationMap, err := GetStationMap(ctx, conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query failed",
			"message": err.Error()})
		logger.Warn("Query failed", zap.Error(err))
		return
	}

	//-------
	casetypeMap, err := GetCaseTypeMap(ctx, conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query failed",
			"message": err.Error()})
		logger.Warn("Query failed", zap.Error(err))
		return
	}

	var (
		whereClauses []string
		args         []interface{}
		argIndex     = 1
	)

	// Always required for LIMIT and OFFSET
	args = append(args, length, start)
	argIndex += 2

	// Build dynamic filters
	if keyword != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("caseid ILIKE '%%' || $%d || '%%'", argIndex))
		args = append(args, keyword)
		argIndex++
	}

	if case_id != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("caseid ILIKE '%%' || $%d || '%%'", argIndex))
		args = append(args, case_id)
		argIndex++
	}

	if case_status_code != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("casestatuscode ILIKE '%%' || $%d || '%%'", argIndex))
		args = append(args, case_status_code)
		argIndex++
	}

	if case_detail != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("casedetail ILIKE '%%' || $%d || '%%'", argIndex))
		args = append(args, case_detail)
		argIndex++
	}

	if command_code != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("commandcode ILIKE '%%' || $%d || '%%'", argIndex))
		args = append(args, command_code)
		argIndex++
	}

	if station_code != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("policestationcode ILIKE '%%' || $%d || '%%'", argIndex))
		args = append(args, station_code)
		argIndex++
	}

	if opened_date != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("openeddate >= $%d::date", argIndex))
		args = append(args, opened_date)
		argIndex++
	}

	if date_to != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("closeddate <= $%d::date", argIndex))
		args = append(args, date_to)
		argIndex++
	}

	if user_create != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("usercreate ILIKE '%%' || $%d || '%%'", argIndex))
		args = append(args, user_create)
		argIndex++
	}

	if casetype_code != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("casetypecode ILIKE '%%' || $%d || '%%'", argIndex))
		args = append(args, casetype_code)
		argIndex++
	}

	// Validate and inject ORDER BY direction
	cleanOrder := strings.ToUpper(strings.TrimSpace(order_by))
	if cleanOrder != "ASC" && cleanOrder != "DESC" {
		cleanOrder = "DESC"
	}

	// Final query
	query := fmt.Sprintf(`
	SELECT id, caseid, casetypecode, priority, phonenumber, casestatuscode,
	       resultcode, casedetail, policestationcode, caselocationaddress, caselocationdetail,
	       specialemergency, urgentamount, createddate, casetypecode, mediatype, vowner, vvin
	FROM public."case"
	%s
	ORDER BY caseid %s
	LIMIT $1 OFFSET $2
`, func() string {
		if len(whereClauses) > 0 {
			return "WHERE " + strings.Join(whereClauses, " AND ")
		}
		return ""
	}(), cleanOrder)

	logger.Debug("Final Query", zap.String("sql", query))

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		logger.Warn("Query failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.CaseListResponse{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		return
	}
	defer rows.Close()

	var caseLists []model.OutputCase
	var errorMsg string
	for rows.Next() {
		var cusCase model.OutputCase
		err := rows.Scan(&cusCase.Id, &cusCase.CaseId, &cusCase.Casetype_code, &cusCase.Priority, &cusCase.Phone_number,
			&cusCase.Case_status_code, &cusCase.Case_status_name, &cusCase.Case_detail, &cusCase.Police_station_name,
			&cusCase.Case_location_address, &cusCase.Case_location_detail, &cusCase.Special_emergency, &cusCase.Urgent_amount,
			&cusCase.Created_date, &cusCase.Casetype_name, &cusCase.Media_type, &cusCase.VOwner, &cusCase.VVin)
		if err != nil {
			logger.Warn("Query failed", zap.Error(err))
			errorMsg = err.Error()
			continue
		}

		jsonData, _ := json.MarshalIndent(casetypeMap, "", "  ")
		logger.Debug("casetypeMap contents",
			zap.String("data", string(jsonData)), // Correct: wraps string as zap.Field
		)
		logger.Debug(*cusCase.Casetype_code)
		//--
		jsonData, _ = json.MarshalIndent(stationMap, "", "  ")
		logger.Debug("stationMap contents",
			zap.String("data", string(jsonData)), // Correct: wraps string as zap.Field
		)
		logger.Debug(*cusCase.Police_station_name)
		//--
		jsonData, _ = json.MarshalIndent(caseStatusMap, "", "  ")
		logger.Debug("caseStatusMap contents",
			zap.String("data", string(jsonData)), // Correct: wraps string as zap.Field
		)
		logger.Debug(*cusCase.Case_status_code)

		cusCase.Case_status_name = Convert(caseStatusMap, cusCase.Case_status_code)
		cusCase.Casetype_name = Convert(casetypeMap, cusCase.Casetype_code)
		cusCase.Police_station_name = Convert(stationMap, cusCase.Police_station_name)

		caseLists = append(caseLists, cusCase)
	}

	// Total count (for frontend pagination)
	var totalCount int
	err = conn.QueryRow(ctx, `SELECT COUNT(*) FROM public."case"`).Scan(&totalCount)
	if err != nil {
		logger.Warn("Query failed", zap.Error(err))
		totalCount = 0
	}

	// Final JSON
	c.JSON(http.StatusOK, model.CaseListResponse{
		Status: "0",
		Msg:    "Success",
		Data: model.CaseListData{
			Draw:            ToInt(start), // Default or parsed from query
			RecordsTotal:    totalCount,
			RecordsFiltered: ToInt(length), // Your logic or default
			Data:            caseLists,
			Error:           errorMsg,
		},
		Desc: "",
	})
}

// @summary Get a specify case by record ID
// @security ApiKeyAuth
// @id Get a specify case by record ID
// @tags Cases
// @accept json
// @produce json
// @Param id path int true "id" default(0)
// @response 200 {object} model.CaseResponse "OK - Request successful"
// @Router /api/v1/cases/{id} [get]
func SearchCaseById(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()
	id := c.Param("id")
	query := `SELECT id, caseid, refercaseid, casetypecode, priority, ways, phonenumber, casestatuscode, casedetail,
	commandcode, policestationcode, caselocationaddress, caselocationdetail,caselat, caselon, transimg,citizencode, 
	extensionreceive, specialemergency,urgentamount, openeddate, saveddate, 
	createddate, starteddate, modifieddate, usercreate, usermodify,responsible,approvedstatus, casetypecode,
	mediatype, vowner, vvin, destlocationaddress, destlocationdetail, destlat, destlon
	FROM public."case" WHERE id = $1 `
	var cusCase model.CaseDetailData
	err := conn.QueryRow(ctx, query, id).Scan(
		&cusCase.ID, &cusCase.CaseID, &cusCase.ReferCaseID, &cusCase.CasetypeCode, &cusCase.Priority,
		&cusCase.Ways, &cusCase.PhoneNumber, &cusCase.CaseStatusCode, &cusCase.CaseDetail,
		&cusCase.CommandCode, &cusCase.PoliceStationCode, &cusCase.CaseLocationAddress, &cusCase.CaseLocationDetail,
		&cusCase.CaseLat, &cusCase.CaseLon, &cusCase.TransImg, &cusCase.CitizenCode, &cusCase.ExtensionReceive,
		&cusCase.SpecialEmergency, &cusCase.UrgentAmount, &cusCase.OpenedDate, &cusCase.SavedDate,
		&cusCase.CreatedDate, &cusCase.StartedDate, &cusCase.ModifiedDate, &cusCase.UserCreate,
		&cusCase.UserModify, &cusCase.Responsible, &cusCase.ApprovedStatus,
		&cusCase.CasetypeName, &cusCase.MediaType, &cusCase.VOwner, &cusCase.VVin, &cusCase.DestLocationAddress,
		&cusCase.DestLocationDetail, &cusCase.DestLat, &cusCase.DestLon,
	)

	// cusCase.CaseStatusName = Convert(caseStatusMap, cusCase.Case_status_code)
	// cusCase.CommandName = Convert(caseStatusMap, cusCase.Case_status_code)
	// cusCase.CitizenFullname
	// cusCase.UserCreateID
	// DepartmentName
	// CasetypeName
	// cusCase.PoliceStationName

	logger.Debug("Query", zap.String("query", query), zap.Any("id", id))

	if err != nil {
		logger.Warn("Query failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.CaseResponse{
			Status: "-1",
			Msg:    "Failure",
			Data:   nil,
			Desc:   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, model.CaseResponse{
		Status: "0",
		Msg:    "Success",
		Data:   &cusCase,
		Desc:   "",
	})
}

// @summary Get a specify case by case code (case_id)
// @security ApiKeyAuth
// @id Get a specify case by case code (case_id)
// @tags Cases
// @accept json
// @produce json
// @Param id  path string true "case code"
// @response 200 {object} model.CaseResponse "OK - Request successful"
// @Router /api/v1/cases/detail/{id} [get]
func SearchCaseByCaseCode(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()
	case_id := c.Param("id")
	query := `SELECT id, caseid, refercaseid, casetypecode, priority, ways, phonenumber, casestatuscode, casedetail,
	commandcode, policestationcode, caselocationaddress, caselocationdetail,caselat, caselon, transimg,citizencode, 
	extensionreceive, specialemergency,urgentamount, openeddate, saveddate, 
	createddate, starteddate, modifieddate, usercreate, usermodify,responsible,approvedstatus, casetypecode,
	mediatype, vowner, vvin, destlocationaddress, destlocationdetail, destlat, destlon
	FROM public."case" WHERE caseid = $1 `
	var cusCase model.CaseDetailData
	err := conn.QueryRow(ctx, query, case_id).Scan(
		&cusCase.ID, &cusCase.CaseID, &cusCase.ReferCaseID, &cusCase.CasetypeCode, &cusCase.Priority,
		&cusCase.Ways, &cusCase.PhoneNumber, &cusCase.CaseStatusCode, &cusCase.CaseDetail,
		&cusCase.CommandCode, &cusCase.PoliceStationCode, &cusCase.CaseLocationAddress, &cusCase.CaseLocationDetail,
		&cusCase.CaseLat, &cusCase.CaseLon, &cusCase.TransImg, &cusCase.CitizenCode, &cusCase.ExtensionReceive,
		&cusCase.SpecialEmergency, &cusCase.UrgentAmount, &cusCase.OpenedDate, &cusCase.SavedDate,
		&cusCase.CreatedDate, &cusCase.StartedDate, &cusCase.ModifiedDate, &cusCase.UserCreate,
		&cusCase.UserModify, &cusCase.Responsible, &cusCase.ApprovedStatus,
		&cusCase.CasetypeName, &cusCase.MediaType, &cusCase.VOwner, &cusCase.VVin, &cusCase.DestLocationAddress,
		&cusCase.DestLocationDetail, &cusCase.DestLat, &cusCase.DestLon,
	)

	// cusCase.CaseStatusName = Convert(caseStatusMap, cusCase.Case_status_code)
	// cusCase.CommandName = Convert(caseStatusMap, cusCase.Case_status_code)
	// cusCase.CitizenFullname
	// cusCase.UserCreateID
	// DepartmentName
	// CasetypeName
	// cusCase.PoliceStationName

	logger.Debug("Query", zap.String("query", query), zap.Any("case_id", case_id))

	if err != nil {
		logger.Warn("Query failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.CaseResponse{
			Status: "-1",
			Msg:    "Failure",
			Data:   nil,
			Desc:   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, model.CaseResponse{
		Status: "0",
		Msg:    "Success",
		Data:   &cusCase,
		Desc:   "",
	})

}

// @summary Create a new case
// @id Create a new case
// @security ApiKeyAuth
// @tags Cases
// @accept json
// @produce json
// @param Case body model.CaseForCreate true "Case data to be created"
// @response 200 {object} model.CreateCaseResponse "OK - Request successful"
// @Router /api/v1/cases [post]
func CreateCase(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	var req model.CaseForCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.CreateCaseResponse{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
			ID:     "",
			CaseID: "",
		})
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	req.SetDefaults() // Set default dates if they're missing

	// now req is ready to use

	var cust model.CreateCaseResponse
	query := `
INSERT INTO public."case"(
	caseid,refercaseid, casetypecode, priority, ways, phonenumber, 
	phonenumberhide, duration,casestatuscode, casecondition, casedetail, 
	caselocationtype,commandcode, policestationcode, caselocationaddress, caselocationdetail, 
	caseroute,caselat, caselon, casedirection, casephoto, 
	transimg, home, citizencode,extensionreceive, casesla, 
	actionprocode, specialemergency,mediacode, mediatype, openeddate, 
	createddate,starteddate, closeddate, modifieddate, usercreate, 
	userclose, usermodify,needambulance, backdated, escaperoute, 
	vowner, vvin, destlocationaddress,destlocationdetail, destlat, 
	destlon, isdup
	)
	VALUES (
		$1, $2, $3, $4, $5, $6, $7,
		$8, $9, $10, $11,
		$12, $13, $14, $15, $16,
		$17, $18, $19, $20, $21, $22, $23,
		$24, $25, $26, $27,
		$28, $29, $30, $31,
		$32, $33, $34, $35, $36,
		$37, $38, $39, $40, $41,
		$42, $43, $44, $45, $46 ,$47 ,$48
	)
	RETURNING id, caseid;
	`
	casePhotoJSON, err := json.Marshal(req.CasePhoto)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case_photo"})
		return
	}

	transImgJSON, err := json.Marshal(req.TransImg)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid trans_img"})
		return
	}

	err = conn.QueryRow(ctx, query, genCaseID(),
		req.ReferCaseID, req.CasetypeCode, req.Priority, req.Ways, req.PhoneNumber,
		req.PhoneNumberHide, req.Duration, req.CaseStatusCode, req.CaseCondition, req.CaseDetail,
		req.CaseLocationType, req.CommandCode, req.PoliceStationCode, req.CaseLocationAddress, req.CaseLocationDetail,
		req.CaseRoute, req.CaseLat, req.CaseLon, req.CaseDirection, casePhotoJSON,
		transImgJSON, req.Home, req.CitizenCode, req.ExtensionReceive, req.CaseSLA,
		req.ActionProCode, req.SpecialEmergency, req.MediaCode, req.MediaType, req.OpenedDate,
		req.CreatedDate, req.StartedDate, req.ClosedDate, req.ModifiedDate, req.UserCreate,
		req.UserClose, req.UserModify, req.NeedAmbulance, req.Backdated, req.EscapeRoute,
		req.VOwner, req.VVin, req.DestLocationAddress, req.DestLocationDetail, req.DestLat,
		req.DestLon, true,
	).Scan(&cust.ID, &cust.CaseID)

	if err != nil {
		// log.Printf("Insert failed: %v", err)
		c.JSON(http.StatusInternalServerError, model.CreateCaseResponse{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
			ID:     "",
			CaseID: "",
		})
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	// Continue logic...
	c.JSON(http.StatusOK, model.CreateCaseResponse{
		Status: "0",
		Msg:    "Success",
		Desc:   "Create case successfully",
		ID:     cust.ID,
		CaseID: cust.CaseID,
	})

}

// @summary Update an existing case
// @id Update an existing case
// @security ApiKeyAuth
// @accept json
// @produce json
// @tags Cases
// @Param id path int true "id"
// @param Case body model.CaseForUpdate true "Case data to be update"
// @response 200 {object} model.UpdateCaseResponse "OK - Request successful"
// @Router /api/v1/cases/{id} [patch]
func UpdateCase(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	id := c.Param("id")

	var req model.CaseForUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Update failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, model.UpdateCaseResponse{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
			ID:     ToInt(id),
		})
		return
	}
	casePhotoJSON, err := json.Marshal(req.CasePhoto)
	if err != nil {
		logger.Warn("Failed to marshal CasePhoto", zap.Error(err))
		// Handle error or return
	}

	transImgJSON, err := json.Marshal(req.TransImg)
	if err != nil {
		logger.Warn("Failed to marshal TransImg", zap.Error(err))
		// Handle error or return
	}
	query := `UPDATE public."case" SET refercaseid=$2,
    casetypecode=$3,
    priority=$4,
    ways=$5,
    phonenumber=$6,
    phonenumberhide=$7,
    duration=$8,
    casestatuscode=$9,
    casecondition=$10,
    casedetail=$11,
    caselocationtype=$12,
    commandcode=$13,
    policestationcode=$14,
    caselocationaddress=$15,
    caselocationdetail=$16,
    caseroute=$17,
    caselat=$18,
    caselon=$19,
    casedirection=$20,
    casephoto=$21,
    transimg=$22,
    home=$23,
    citizencode=$24,
    extensionreceive=$25,
    casesla=$26,
    actionprocode=$27,
    specialemergency=$28,
    mediacode=$29,
    mediatype=$30,
    openeddate=$31,
    createddate=$32,
    starteddate=$33,
    closeddate=$34,
    modifieddate=$35,
    usercreate=$36,
    userclose=$37,
    usermodify=$38,
    needambulance=$39,
    backdated=$40,
    escaperoute=$41,
    vowner=$42,
    vvin=$43,
    destlocationaddress=$44,
    destlocationdetail=$45,
    destlat=$46,
    destlon=$47
	WHERE id = $1 `
	_, err = conn.Exec(ctx, query,
		id, req.ReferCaseID, req.CasetypeCode, req.Priority, req.Ways,
		req.PhoneNumber, req.PhoneNumberHide, req.Duration, req.CaseStatusCode, req.CaseCondition,
		req.CaseDetail, req.CaseLocationType, req.CommandCode, req.PoliceStationCode, req.CaseLocationAddress,
		req.CaseLocationDetail, req.CaseRoute, req.CaseLat, req.CaseLon, req.CaseDirection,
		casePhotoJSON, transImgJSON, req.Home, req.CitizenCode, req.ExtensionReceive,
		req.CaseSLA, req.ActionProCode, req.SpecialEmergency, req.MediaCode, req.MediaType,
		req.OpenedDate, req.CreatedDate, req.StartedDate, req.ClosedDate, req.ModifiedDate,
		req.UserCreate, req.UserClose, req.UserModify, req.NeedAmbulance, req.Backdated,
		req.EscapeRoute, req.VOwner, req.VVin, req.DestLocationAddress,
		req.DestLocationDetail, req.DestLat, req.DestLon,
	)
	logger.Debug("Update Case SQL Args",
		zap.String("query", query),
		zap.Any("Input", []any{
			id,
			req.ReferCaseID, req.CasetypeCode, req.Priority, req.Ways,
			req.PhoneNumber, req.PhoneNumberHide, req.Duration, req.CaseStatusCode, req.CaseCondition,
			req.CaseDetail, req.CaseLocationType, req.CommandCode, req.PoliceStationCode, req.CaseLocationAddress,
			req.CaseLocationDetail, req.CaseRoute, req.CaseLat, req.CaseLon, req.CaseDirection,
			string(casePhotoJSON), string(transImgJSON), req.Home, req.CitizenCode, req.ExtensionReceive,
			req.CaseSLA, req.ActionProCode, req.SpecialEmergency, req.MediaCode, req.MediaType,
			req.OpenedDate, req.CreatedDate, req.StartedDate, req.ClosedDate, req.ModifiedDate,
			req.UserCreate, req.UserClose, req.UserModify, req.NeedAmbulance, req.Backdated,
			req.EscapeRoute, req.VOwner, req.VVin, req.DestLocationAddress, req.DestLocationDetail,
			req.DestLat, req.DestLon,
		}))
	if err != nil {
		// log.Printf("Insert failed: %v", err)
		c.JSON(http.StatusInternalServerError, model.UpdateCaseResponse{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
			ID:     ToInt(id),
		})
		logger.Warn("Update failed", zap.Error(err))
		return
	}

	// Continue logic...
	c.JSON(http.StatusOK, model.UpdateCaseResponse{
		Status: "0",
		Msg:    "Success",
		Desc:   "Update case successfully",
		ID:     ToInt(id),
	})
}

// @summary Delete an existing case
// @id Delete an existing case
// @security ApiKeyAuth
// @accept json
// @tags Cases
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.DeleteCaseResponse "OK - Request successful"
// @Router /api/v1/cases/{id} [delete]
func DeleteCase(c *gin.Context) {

	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	id := c.Param("id")
	query := `DELETE FROM public."case" WHERE id = $1 `
	logger.Debug("Query", zap.String("query", query), zap.Any("id", id))
	_, err := conn.Exec(ctx, query, id)
	if err != nil {
		// log.Printf("Insert failed: %v", err)
		c.JSON(http.StatusInternalServerError, model.DeleteCaseResponse{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Update failed", zap.Error(err))
		return
	}

	// Continue logic...
	c.JSON(http.StatusOK, model.DeleteCaseResponse{
		Status: "0",
		Msg:    "Success",
		Desc:   "Delete case successfully",
	})
}

// @summary Update an existing case status (close or cancel)
// @id Update an existing case status (close or cancel)
// @security ApiKeyAuth
// @accept json
// @tags Cases
// @produce json
// @Param id path int true "id"
// @param Case body model.CaseCloseInput true "Case data to be update"
// @response 200 {object} model.UpdateCaseResponse "OK - Request successful"
// @Router /api/v1/cases/close/{id} [patch]
func UpdateCaseClose(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	id := c.Param("id")

	var req model.CaseCloseInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.UpdateCaseResponse{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
			ID:     ToInt(id),
		})
		return
	}
	transImgJSON, err := json.Marshal(req.TransImg)
	if err != nil {
		logger.Warn("Failed to marshal TransImg", zap.Error(err))
		// Handle error or return
	}
	query := `UPDATE public."case" SET
    casestatuscode=$2,
    resultcode=$3,
    resultdetail=$4,
    transimg=$5,
    closeddate=$6,
    modifieddate=$7,
    userclose=$8,
    usermodify=$9
	WHERE id = $1 `
	_, err = conn.Exec(ctx, query,
		id, req.CaseStatusCode, req.ResultCode, req.ResultDetail, transImgJSON,
		req.ClosedDate, req.ModifiedDate, req.UserClose, req.UserModify,
	)
	logger.Debug("Update Case SQL Query",
		zap.String("query", query),
	)

	logger.Debug("Update Case SQL Args",
		zap.Any("Input", []any{
			id,
			req.CaseStatusCode,
			req.ResultCode,
			req.ResultDetail,
			transImgJSON,
			req.ClosedDate,
			req.ModifiedDate,
			req.UserClose,
			req.UserModify,
		}),
	)

	if err != nil {
		// log.Printf("Insert failed: %v", err)
		c.JSON(http.StatusInternalServerError, model.UpdateCaseResponse{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
			ID:     ToInt(id),
		})
		logger.Warn("Update failed", zap.Error(err))
		return
	}

	// Continue logic...
	c.JSON(http.StatusOK, model.UpdateCaseResponse{
		Status: "0",
		Msg:    "Success",
		Desc:   "Update case successfully",
		ID:     ToInt(id),
	})
}
