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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type CaseHandler struct {
	DB *pgx.Conn
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
// @response 201 {object} model.CaseListData "Created - Resource created successfully"
// @response 400 {object} model.CaseListData "Bad Request - Invalid request parameters"
// @response 401 {object} model.CaseListData "Unauthorized - Invalid or missing authentication"
// @response 403 {object} model.CaseListData "Forbidden - Insufficient permissions"
// @response 404 {object} model.CaseListData "Not Found - Resource doesn't exist"
// @response 422 {object} model.CaseListData "Bad Request and Not Found (temporary)"
// @response 429 {object} model.CaseListData "Too Many Requests - Rate limit exceeded"
// @response 500 {object} model.CaseListData "Internal Server Error"
// @Router /api/v1/cases [get]
func (h *CaseHandler) ListCase(c *gin.Context) {
	logger := config.GetLog()
	keyword := c.Query("keyword")
	startStr := c.DefaultQuery("start", "0")
	lengthStr := c.DefaultQuery("length", "0")

	start, err := strconv.Atoi(startStr)
	if err != nil {
		start = 1 // fallback default
	}

	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		length = 10 // fallback default
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	//---

	caseStatusMap, err := GetCaseStatusMap(ctx, h.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query failed"})
		logger.Warn("Query failed", zap.Error(err))
		return
	}
	//---
	stationMap, err := GetStationMap(ctx, h.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query failed"})
		logger.Warn("Query failed", zap.Error(err))
		return
	}

	//-------
	casetypeMap, err := GetCaseTypeMap(ctx, h.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query failed"})
		logger.Warn("Query failed", zap.Error(err))
		return
	}
	//----

	query := `SELECT id, caseid,casetypecode,priority,phonenumber,casestatuscode ,
			resultcode, casedetail,policestationcode,caselocationaddress,caselocationdetail,
			specialemergency,urgentamount,createddate,casetypecode,mediatype,vowner,vvin
			FROM public."case" WHERE caseid ILIKE '%' || $3 || '%' LIMIT $1 OFFSET $2`
	// finalQuery := fmt.Sprintf(`
	// 		SELECT id, caseid, casetypecode, priority, phonenumber, casestatuscode,
	// 		resultcode, casedetail, policestationcode, caselocationaddress, caselocationdetail,
	// 		specialemergency, urgentamount, createddate, casetypecode, mediatype, vowner, vvin
	// 		FROM public."case"
	// 		WHERE caseid ILIKE '%%%s%%'
	// 		LIMIT %d OFFSET %d`, keyword, length, start)

	// logger.Debug("Final Query", zap.String("sql", finalQuery))

	rows, err := h.DB.Query(ctx, query, length, start, keyword)
	if err != nil {
		logger.Warn("Query failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
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

		fmt.Printf("%+v\n", casetypeMap)
		fmt.Println(*cusCase.Casetype_code)
		fmt.Printf("%+v\n", stationMap)
		fmt.Println(*cusCase.Police_station_name)
		fmt.Printf("%+v\n", caseStatusMap)
		fmt.Println(*cusCase.Case_status_code)

		cusCase.Case_status_name = Convert(caseStatusMap, cusCase.Case_status_code)
		cusCase.Casetype_name = Convert(casetypeMap, cusCase.Casetype_code)
		cusCase.Police_station_name = Convert(stationMap, cusCase.Police_station_name)

		caseLists = append(caseLists, cusCase)
	}

	// Total count (for frontend pagination)
	var totalCount int
	err = h.DB.QueryRow(ctx, `SELECT COUNT(*) FROM public."case"`).Scan(&totalCount)
	if err != nil {
		logger.Warn("Query failed", zap.Error(err))
		totalCount = 0
	}

	// Final JSON
	c.JSON(http.StatusOK, model.CaseListResponse{
		Status: "0",
		Msg:    "Success",
		Data: model.CaseListData{
			Draw:            start, // Default or parsed from query
			RecordsTotal:    totalCount,
			RecordsFiltered: length, // Your logic or default
			Data:            caseLists,
			Error:           errorMsg,
		},
		Desc: "",
	})
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
// @response 201 {object} model.CaseListData "Created - Resource created successfully"
// @response 400 {object} model.CaseListData "Bad Request - Invalid request parameters"
// @response 401 {object} model.CaseListData "Unauthorized - Invalid or missing authentication"
// @response 403 {object} model.CaseListData "Forbidden - Insufficient permissions"
// @response 404 {object} model.CaseListData "Not Found - Resource doesn't exist"
// @response 422 {object} model.CaseListData "Bad Request and Not Found (temporary)"
// @response 429 {object} model.CaseListData "Too Many Requests - Rate limit exceeded"
// @response 500 {object} model.CaseListData "Internal Server Error"
// @Router /api/v1/cases/search [get]
func (h *CaseHandler) SearchCase(c *gin.Context) {}

// @summary Get a specify case by record ID
// @security ApiKeyAuth
// @id Get a specify case by record ID
// @tags Cases
// @accept json
// @produce json
// @Param id path int true "id" default(0)
// @response 200 {object} model.CaseDetailResponse "OK - Request successful"
// @response 201 {object} model.CaseDetailResponse "Created - Resource created successfully"
// @response 400 {object} model.CaseDetailResponse "Bad Request - Invalid request parameters"
// @response 401 {object} model.CaseDetailResponse "Unauthorized - Invalid or missing authentication"
// @response 403 {object} model.CaseDetailResponse "Forbidden - Insufficient permissions"
// @response 404 {object} model.CaseDetailResponse "Not Found - Resource doesn't exist"
// @response 422 {object} model.CaseDetailResponse "Bad Request and Not Found (temporary)"
// @response 429 {object} model.CaseDetailResponse "Too Many Requests - Rate limit exceeded"
// @response 500 {object} model.CaseDetailResponse "Internal Server Error"
// @Router /api/v1/cases/{id} [get]
func (h *CaseHandler) SearchCaseById(c *gin.Context) {}

// @summary Get a specify case by case code (case_id)
// @security ApiKeyAuth
// @id Get a specify case by case code (case_id)
// @tags Cases
// @accept json
// @produce json
// @Param id path string true "case code"
// @response 200 {object} model.CaseDetailResponse "OK - Request successful"
// @response 201 {object} model.CaseDetailResponse "Created - Resource created successfully"
// @response 400 {object} model.CaseDetailResponse "Bad Request - Invalid request parameters"
// @response 401 {object} model.CaseDetailResponse "Unauthorized - Invalid or missing authentication"
// @response 403 {object} model.CaseDetailResponse "Forbidden - Insufficient permissions"
// @response 404 {object} model.CaseDetailResponse "Not Found - Resource doesn't exist"
// @response 422 {object} model.CaseDetailResponse "Bad Request and Not Found (temporary)"
// @response 429 {object} model.CaseDetailResponse "Too Many Requests - Rate limit exceeded"
// @response 500 {object} model.CaseDetailResponse "Internal Server Error"
// @Router /api/v1/cases/detail/{id} [get]
func (h *CaseHandler) SearchCaseByCaseCode(c *gin.Context) {}

// @summary Create a new case
// @id Create a new case
// @security ApiKeyAuth
// @tags Cases
// @accept json
// @produce json
// @param Case body model.CaseForCreate true "Case data to be created"
// @response 200 {object} model.CreateCaseResponse "OK - Request successful"
// @response 201 {object} model.CreateCaseResponse "Created - Resource created successfully"
// @response 400 {object} model.CreateCaseResponse "Bad Request - Invalid request parameters"
// @response 401 {object} model.CreateCaseResponse "Unauthorized - Invalid or missing authentication"
// @response 403 {object} model.CreateCaseResponse "Forbidden - Insufficient permissions"
// @response 404 {object} model.CreateCaseResponse "Not Found - Resource doesn't exist"
// @response 422 {object} model.CreateCaseResponse "Bad Request and Not Found (temporary)"
// @response 429 {object} model.CreateCaseResponse "Too Many Requests - Rate limit exceeded"
// @response 500 {object} model.CreateCaseResponse "Internal Server Error"
// @Router /api/v1/cases [post]
func (h *CaseHandler) CreateCase(c *gin.Context) {
	logger := config.GetLog()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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

	err = h.DB.QueryRow(ctx, query, genCaseID(),
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
// @response 201 {object} model.UpdateCaseResponse "Created - Resource created successfully"
// @response 400 {object} model.UpdateCaseResponse "Bad Request - Invalid request parameters"
// @response 401 {object} model.UpdateCaseResponse "Unauthorized - Invalid or missing authentication"
// @response 403 {object} model.UpdateCaseResponse "Forbidden - Insufficient permissions"
// @response 404 {object} model.UpdateCaseResponse "Not Found - Resource doesn't exist"
// @response 422 {object} model.UpdateCaseResponse "Bad Request and Not Found (temporary)"
// @response 429 {object} model.UpdateCaseResponse "Too Many Requests - Rate limit exceeded"
// @response 500 {object} model.UpdateCaseResponse "Internal Server Error"
// @Router /api/v1/cases/{id} [patch]
func (h *CaseHandler) UpdateCase(c *gin.Context) {}

// @summary Delete an existing case
// @id Delete an existing case
// @security ApiKeyAuth
// @accept json
// @tags Cases
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.DeleteCaseResponse "OK - Request successful"
// @response 201 {object} model.DeleteCaseResponse "Created - Resource created successfully"
// @response 400 {object} model.DeleteCaseResponse "Bad Request - Invalid request parameters"
// @response 401 {object} model.DeleteCaseResponse "Unauthorized - Invalid or missing authentication"
// @response 403 {object} model.DeleteCaseResponse "Forbidden - Insufficient permissions"
// @response 404 {object} model.DeleteCaseResponse "Not Found - Resource doesn't exist"
// @response 422 {object} model.DeleteCaseResponse "Bad Request and Not Found (temporary)"
// @response 429 {object} model.DeleteCaseResponse "Too Many Requests - Rate limit exceeded"
// @response 500 {object} model.DeleteCaseResponse "Internal Server Error"
// @Router /api/v1/cases/{id} [delete]
func (h *CaseHandler) DeleteCase(c *gin.Context) {}

// @summary Update an existing case status (close or cancel)
// @id Update an existing case status (close or cancel)
// @security ApiKeyAuth
// @accept json
// @tags Cases
// @produce json
// @Param id path int true "id"
// @param Case body model.CaseCloseInput true "Case data to be update"
// @response 200 {object} model.UpdateCaseResponse "OK - Request successful"
// @response 201 {object} model.UpdateCaseResponse "Created - Resource created successfully"
// @response 400 {object} model.UpdateCaseResponse "Bad Request - Invalid request parameters"
// @response 401 {object} model.UpdateCaseResponse "Unauthorized - Invalid or missing authentication"
// @response 403 {object} model.UpdateCaseResponse "Forbidden - Insufficient permissions"
// @response 404 {object} model.UpdateCaseResponse "Not Found - Resource doesn't exist"
// @response 422 {object} model.UpdateCaseResponse "Bad Request and Not Found (temporary)"
// @response 429 {object} model.UpdateCaseResponse "Too Many Requests - Rate limit exceeded"
// @response 500 {object} model.UpdateCaseResponse "Internal Server Error"
// @Router /api/v1/cases/close/{id} [patch]
func (h *CaseHandler) UpdateCaseClose(c *gin.Context) {}
