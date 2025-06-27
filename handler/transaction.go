package handler

import (
	"mainPackage/config"
	"mainPackage/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// ListCase godoc
// @summary List Transaction
// @tags Transaction
// @security ApiKeyAuth
// @id ListTransaction
// @accept json
// @produce json
// @Param start query int false "start" default(0)
// @Param length query int false "length" default(10)
// @Param keyword query string false "keyword"
// @response 200 {object} model.CaseListData "OK - Request successful"
// @Router /api/v1/trans [get]
func ListTransaction(c *gin.Context) {
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
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	query := `SELECT id, caseid, usercode, username, receivedate, arrivedate, closedate, canceldate, duration,
	       suggestroute, usersla, viewed, casestatuscode, notistage, userclosedjob, resultcodet, resultdetailt,
	       createddate, createdmodify, owner, updatedaccount, vehiclecode, actioncartype, timetoarrive, lat, lon
	FROM public.case_transaction`

	var rows pgx.Rows

	if keyword == "" {
		query += ` LIMIT $1 OFFSET $2`
		rows, err = conn.Query(ctx, query, length, start)
	} else {
		query += ` WHERE caseid ILIKE '%' || $3 || '%' LIMIT $1 OFFSET $2`
		rows, err = conn.Query(ctx, query, length, start, keyword)
	}

	logger.Debug(`Query`,
		zap.String("query", query),
		zap.Any("args", []any{
			length, start, keyword,
		}))

	if err != nil {
		logger.Warn("Query failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.CaseTransactionResponse{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
			Data:   []model.CaseTransactionData{},
		})
		return
	}
	defer rows.Close()
	var DataOpt []model.CaseTransactionData
	var errorMsg string
	for rows.Next() {
		var Transaction model.CaseTransactionData
		err := rows.Scan(&Transaction.ID, &Transaction.CaseID, &Transaction.UserCode, &Transaction.UserName, &Transaction.ReceiveDate,
			&Transaction.ArriveDate, &Transaction.CloseDate, &Transaction.CancelDate, &Transaction.Duration,
			&Transaction.SuggestRoute, &Transaction.UserSLA, &Transaction.Viewed, &Transaction.CaseStatusCode,
			&Transaction.NotiStage, &Transaction.UserClosedJob, &Transaction.ResultCodeT, &Transaction.ResultDetailT, &Transaction.CreatedDate,
			&Transaction.CreatedModify, &Transaction.Owner, &Transaction.UpdatedAccount, &Transaction.VehicleCode, &Transaction.ActionCarType,
			&Transaction.TimeToArrive, &Transaction.Lat, &Transaction.Lon)
		if err != nil {
			logger.Warn("Query failed", zap.Error(err))
			errorMsg = err.Error()
			continue
		}
		DataOpt = append(DataOpt, Transaction)
	}
	if errorMsg != "" {
		response := model.CaseTransactionResponse{
			Status: "-1",
			Msg:    "Failed",
			Data:   DataOpt,
			Desc:   errorMsg,
		}
		c.JSON(http.StatusInternalServerError, response)
	} else {
		response := model.CaseTransactionResponse{
			Status: "0",
			Msg:    "Success",
			Data:   DataOpt,
			Desc:   "",
		}
		c.JSON(http.StatusOK, response)
	}
}

// ListCase godoc
// @summary Get a specify transaction by record ID or case code (caseId)
// @tags Transaction
// @security ApiKeyAuth
// @id Get a specify transaction by record ID or case code (caseId)
// @accept json
// @produce json
// @Param id path string true "id"
// @response 200 {object} model.CaseListData "OK - Request successful"
// @Router /api/v1/trans/{id} [get]
func SearchTransaction(c *gin.Context) {
	logger := config.GetLog()
	id := c.Param("id")

	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	query := `SELECT id, caseid, usercode, username, receivedate, arrivedate, closedate, canceldate, duration,
	       suggestroute, usersla, viewed, casestatuscode, notistage, userclosedjob, resultcodet, resultdetailt,
	       createddate, createdmodify, owner, updatedaccount, vehiclecode, actioncartype, timetoarrive, lat, lon
	FROM public.case_transaction WHERE (caseid ILIKE '%' || $1 || '%' OR id::text ILIKE '%' || $1 || '%')`

	var rows pgx.Rows

	rows, err := conn.Query(ctx, query, id)

	logger.Debug(`Query`,
		zap.String("query", query),
		zap.Any("args", []any{
			id,
		}))

	if err != nil {
		logger.Warn("Query failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.CaseTransactionResponse{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		return
	}
	defer rows.Close()
	var DataOpt []model.CaseTransactionData
	var errorMsg string
	for rows.Next() {
		var Transaction model.CaseTransactionData
		err := rows.Scan(&Transaction.ID, &Transaction.CaseID, &Transaction.UserCode, &Transaction.UserName, &Transaction.ReceiveDate,
			&Transaction.ArriveDate, &Transaction.CloseDate, &Transaction.CancelDate, &Transaction.Duration,
			&Transaction.SuggestRoute, &Transaction.UserSLA, &Transaction.Viewed, &Transaction.CaseStatusCode,
			&Transaction.NotiStage, &Transaction.UserClosedJob, &Transaction.ResultCodeT, &Transaction.ResultDetailT, &Transaction.CreatedDate,
			&Transaction.CreatedModify, &Transaction.Owner, &Transaction.UpdatedAccount, &Transaction.VehicleCode, &Transaction.ActionCarType,
			&Transaction.TimeToArrive, &Transaction.Lat, &Transaction.Lon)
		if err != nil {
			logger.Warn("Query failed", zap.Error(err))
			errorMsg = err.Error()
			continue
		}
		DataOpt = append(DataOpt, Transaction)
	}
	if errorMsg != "" {
		response := model.CaseTransactionResponse{
			Status: "-1",
			Msg:    "Failed",
			Data:   DataOpt,
			Desc:   errorMsg,
		}
		c.JSON(http.StatusInternalServerError, response)
	} else {
		response := model.CaseTransactionResponse{
			Status: "0",
			Msg:    "Success",
			Data:   DataOpt,
			Desc:   "",
		}
		c.JSON(http.StatusOK, response)
	}
}

// ListCase godoc
// @summary List Transaction Note
// @tags Transaction
// @security ApiKeyAuth
// @id ListTransactionNote
// @accept json
// @produce json
// @Param id path string true "id"
// @response 200 {object} model.CaseListData "OK - Request successful"
// @Router /api/v1/notes/{id} [get]
func ListTransactionNote(c *gin.Context) {
	logger := config.GetLog()
	id := c.Param("id")

	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	query := `SELECT  * FROM public.case_note WHERE (caseid ILIKE '%' || $1 || '%')`

	var rows pgx.Rows

	rows, err := conn.Query(ctx, query, id)

	logger.Debug(`Query`,
		zap.String("query", query),
		zap.Any("args", []any{
			id,
		}))

	if err != nil {
		logger.Warn("Query failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.CaseTransactionResponse{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		return
	}
	defer rows.Close()
	var DataOpt []model.CaseNote
	var errorMsg string
	for rows.Next() {
		var Transaction model.CaseNote
		err := rows.Scan(&Transaction.ID, &Transaction.CaseID, &Transaction.Detail, &Transaction.CreatedDate, &Transaction.ModifiedDate,
			&Transaction.UserCreate, &Transaction.UserModify)
		if err != nil {
			logger.Warn("Query failed", zap.Error(err))
			errorMsg = err.Error()
			continue
		}
		DataOpt = append(DataOpt, Transaction)
	}
	if errorMsg != "" {
		response := model.CaseNoteResponse{
			Status: "-1",
			Msg:    "Failed",
			Data:   DataOpt,
			Desc:   errorMsg,
		}
		c.JSON(http.StatusInternalServerError, response)
	} else {
		response := model.CaseNoteResponse{
			Status: "0",
			Msg:    "Success",
			Data:   DataOpt,
			Desc:   "",
		}
		c.JSON(http.StatusOK, response)
	}
}

// @summary Create transaction
// @security ApiKeyAuth
// @id Create transaction
// @tags Transaction
// @accept json
// @produce json
// @param Case body model.CaseTransactionModelInput true "Case data to be created"
// @response 200 {object} model.CaseTransactionCRUDResponse "OK - Request successful"
// @Router /api/v1/trans [post]
func CreateTransaction(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	var req model.CaseTransactionModelInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.CaseTransactionCRUDResponse{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	// now req is ready to use

	var cust model.CaseTransactionCRUDResponse
	query := `
INSERT INTO public.case_transaction(
	caseid, usercode, username, receivedate, arrivedate, closedate, canceldate,
	 duration, suggestroute, usersla, viewed, casestatuscode, notistage,
	  userclosedjob, resultcode, resultdetail, createddate, createdmodify,
	   owner, updatedaccount, vehiclecode, actioncartype, timetoarrive, lat, lon
	)
	VALUES (
		$1, $2, $3, $4, $5, $6, $7,
		$8, $9, $10, $11,
		$12, $13, $14, $15, $16,
		$17, $18, $19, $20, $21, $22, $23,
		$24, $25
	)
	RETURNING id, caseid;
	`
	logger.Debug(`Query`,
		zap.String("query", query),
		zap.Any("args", []any{
			req,
		}))
	err := conn.QueryRow(ctx, query,
		req.CaseID, req.UserCode, req.UserName, req.ReceiveDate, req.ArriveDate,
		req.CloseDate, req.CancelDate, req.Duration, req.SuggestRoute, req.UserSLA,
		req.Viewed, req.CaseStatusCode, req.NotiStage, req.UserClosedJob, req.ResultCode,
		req.ResultDetail, req.CreatedDate, req.CreatedModify, req.Owner, req.UpdatedAccount,
		req.VehicleCode, req.ActionCarType, req.TimeToArrive, req.Lat, req.Lon,
	).Scan(&cust.ID, &cust.CaseID)

	if err != nil {
		// log.Printf("Insert failed: %v", err)
		c.JSON(http.StatusInternalServerError, model.CaseTransactionCRUDResponse{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	// Continue logic...
	c.JSON(http.StatusOK, model.CaseTransactionCRUDResponse{
		Status: "0",
		Msg:    "Success",
		Desc:   "Create case successfully",
		ID:     cust.ID,
		CaseID: cust.CaseID,
	})

}

// @summary Create transaction Note
// @security ApiKeyAuth
// @id Create transaction Note
// @tags Transaction
// @accept json
// @produce json
// @param Case body model.CaseNoteInput true "Case data to be created"
// @response 200 {object} model.CaseListData "OK - Request successful"
// @Router /api/v1/notes [post]
func CreateTransactionNote(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	var req model.CaseNoteInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.CaseNoteInputResponse{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	// now req is ready to use

	var cust model.CaseNoteInputResponse
	query := `INSERT INTO public.case_note(caseid, detail, createddate, modifieddate, usercreate,
 				usermodify) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, caseid;`
	logger.Debug(`Query`,
		zap.String("query", query),
		zap.Any("args", []any{
			req,
		}))
	err := conn.QueryRow(ctx, query,
		req.CaseID, req.Detail, req.CreatedDate, req.ModifiedDate, req.UserCreate, req.UserModify,
	).Scan(&cust.ID, &cust.CaseID)

	if err != nil {
		// log.Printf("Insert failed: %v", err)
		c.JSON(http.StatusInternalServerError, model.CaseNoteInputResponse{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	// Continue logic...
	c.JSON(http.StatusOK, model.CaseNoteInputResponse{
		Status: "0",
		Msg:    "Success",
		Desc:   "Create note of transaction successfully",
		CaseID: cust.CaseID,
	})

}

// @summary Update an existing transaction
// @id Update an existing transaction
// @security ApiKeyAuth
// @tags Transaction
// @accept json
// @produce json
// @Param id path int true "id"
// @param Body body model.CaseTransactionUpdateInput true "Body"
// @response 200 {object} model.CreateCaseResponse "OK - Request successful"
// @Router /api/v1/trans/{id} [patch]
func UpdateTransaction(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()
	id := c.Param("id")
	var req model.CaseTransactionUpdateInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.CaseTransactionCRUDResponse{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	// now req is ready to use

	var cust model.CaseTransactionCRUDResponse
	query := `
UPDATE public.case_transaction SET
	caseid=$2, usercode=$3, username=$4, receivedate=$5, arrivedate=$6, closedate=$7, canceldate=$8,
	 duration=$9, suggestroute=$10, usersla=$11, viewed=$12, casestatuscode=$13, notistage=$14,
	  userclosedjob=$15, resultcode=$16, resultdetail=$17, createddate=$18, createdmodify=$19,
	   owner=$20, updatedaccount=$21, vehiclecode=$22, actioncartype=$23, timetoarrive=$24, lat=$25, lon=$26
	WHERE id = $1

	RETURNING id, caseid;
	`
	logger.Debug(`Query`,
		zap.String("query", query),
		zap.Any("args", []any{
			req,
		}))
	err := conn.QueryRow(ctx, query, id,
		req.CaseID, req.UserCode, req.UserName, req.ReceiveDate, req.ArriveDate,
		req.CloseDate, req.CancelDate, req.Duration, req.SuggestRoute, req.UserSLA,
		req.Viewed, req.CaseStatusCode, req.NotiStage, req.UserClosedJob, req.ResultCode,
		req.ResultDetail, req.CreatedDate, req.CreatedModify, req.Owner, req.UpdatedAccount,
		req.VehicleCode, req.ActionCarType, req.TimeToArrive, req.Lat, req.Lon,
	).Scan(&cust.ID, &cust.CaseID)

	if err != nil {
		// log.Printf("Insert failed: %v", err)
		c.JSON(http.StatusInternalServerError, model.CaseTransactionCRUDResponse{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	// Continue logic...
	c.JSON(http.StatusOK, model.CaseTransactionCRUDResponse{
		Status: "0",
		Msg:    "Success",
		Desc:   "Update case transaction successfully",
		ID:     cust.ID,
	})

}

// @summary Delete an existing transaction
// @id Delete an existing transaction
// @security ApiKeyAuth
// @accept json
// @tags Transaction
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.DeleteCaseResponse "OK - Request successful"
// @Router /api/v1/trans/{id} [delete]
func DeleteTransaction(c *gin.Context) {

	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	id := c.Param("id")
	query := `DELETE FROM public."case_transaction" WHERE id = $1 `
	logger.Debug("Query", zap.String("query", query), zap.Any("id", id))
	_, err := conn.Exec(ctx, query, id)
	if err != nil {
		// log.Printf("Insert failed: %v", err)
		c.JSON(http.StatusInternalServerError, model.CaseTransactionCRUDResponse{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Update failed", zap.Error(err))
		return
	}

	// Continue logic...
	c.JSON(http.StatusOK, model.CaseTransactionCRUDResponse{
		Status: "0",
		Msg:    "Success",
		Desc:   "Delete case transaction successfully",
	})
}
