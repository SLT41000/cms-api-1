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

// @summary Get Customer
// @tags Customer
// @security ApiKeyAuth
// @id Get Customer
// @accept json
// @produce json
// @Param start query int false "start" default(0)
// @Param length query int false "length" default(10)
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/customer [get]
func CustomerList(c *gin.Context) {
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
	logger.Debug(`Query`, zap.Any("start", start))
	logger.Debug(`Query`, zap.Any("length", length))
	query := `SELECT id, "orgId", "displayName", title, "firstName", "middleName", "lastName", "citizenId", dob, blood, gender, "mobileNo", address, photo, email, usertype, active, "createdAt", "updatedAt", "createdBy", "updatedBy"
		FROM public.cust_customers 
		WHERE "orgId"=$1 LIMIT $2 OFFSET $3`

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query))
	orgId := GetVariableFromToken(c, "orgId")
	id := c.Param("id")
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	txtId := uuid.New().String()

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
			txtId, id, "Customer", "CustomerList", "",
			"search", -1, start_time, GetQueryParams(c), response, "Query : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer rows.Close()
	var errorMsg string
	var u model.Customer
	var userList []model.Customer
	found := false
	rowIndex := 0
	for rows.Next() {
		rowIndex++
		err := rows.Scan(
			&u.ID,
			&u.OrgID,
			&u.DisplayName,
			&u.Title,
			&u.FirstName,
			&u.MiddleName,
			&u.LastName,
			&u.CitizenID,
			&u.DOB,
			&u.Blood,
			&u.Gender,
			&u.MobileNo,
			&u.Address,
			&u.Photo,
			&u.Email,
			&u.UserType,
			&u.Active,
			&u.CreatedAt,
			&u.UpdatedAt,
			&u.CreatedBy,
			&u.UpdatedBy,
		)
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
				txtId, id, "Customer", "CustomerList", "",
				"search", -1, start_time, GetQueryParams(c), response, "Scan : "+err.Error(),
			)
			//=======AUDIT_END=====//
			c.JSON(http.StatusInternalServerError, response)
			return
		}
		userList = append(userList, u)
		found = true
	}
	if !found {
		response := model.Response{
			Status: "-1",
			Msg:    "Failed",
			Desc:   "Not found",
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Customer", "CustomerList", "",
			"search", -1, start_time, GetQueryParams(c), response, "Not Found : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
	} else {
		response := model.Response{
			Status: "0",
			Msg:    "Success",
			Data:   userList,
			Desc:   "",
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Customer", "CustomerList", "",
			"search", -1, start_time, GetQueryParams(c), response, "GetCustomerList Success",
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusOK, response)
	}
}

// @summary Get Customer by Id
// @tags Customer
// @security ApiKeyAuth
// @id Get Customer by Id
// @accept json
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/customer/{id} [get]
func CustomerById(c *gin.Context) {
	logger := utils.GetLog()
	id := c.Param("id")
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	orgId := GetVariableFromToken(c, "orgId")
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	txtId := uuid.New().String()
	defer cancel()
	defer conn.Close(ctx)
	query := `SELECT id, "orgId", "displayName", title, "firstName", "middleName", "lastName", "citizenId", dob, blood, gender, "mobileNo", address, photo, email, usertype, active, "createdAt", "updatedAt", "createdBy", "updatedBy"
		FROM public.cust_customers WHERE id=$1 AND "orgId"=$2`

	var u model.Customer
	logger.Debug(`Query`, zap.String("query", query))
	err := conn.QueryRow(ctx, query, id, orgId).Scan(
		&u.ID,
		&u.OrgID,
		&u.DisplayName,
		&u.Title,
		&u.FirstName,
		&u.MiddleName,
		&u.LastName,
		&u.CitizenID,
		&u.DOB,
		&u.Blood,
		&u.Gender,
		&u.MobileNo,
		&u.Address,
		&u.Photo,
		&u.Email,
		&u.UserType,
		&u.Active,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.CreatedBy,
		&u.UpdatedBy,
	)
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
			txtId, "", "Customer", "CustomerById", "",
			"search", -1, start_time, GetQueryParams(c), response, "Query failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Data:   u,
		Desc:   "",
	}
	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, "", "Customer", "CustomerById", "",
		"search", 0, start_time, GetQueryParams(c), response, "GetCustomerById Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)

}

// @summary Get Customer by PhoneNo
// @tags Customer
// @security ApiKeyAuth
// @id Get Customer by PhoneNo
// @accept json
// @produce json
// @Param phoneNo path string true "phoneNo"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/customer/byPhoneNo/{phoneNo} [get]
func CustomerByPhoneNo(c *gin.Context) {
	logger := utils.GetLog()
	phoneNo := c.Param("phoneNo")
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	orgId := GetVariableFromToken(c, "orgId")
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	txtId := uuid.New().String()
	defer cancel()
	defer conn.Close(ctx)
	query := `SELECT id, "orgId", "displayName", title, "firstName", "middleName", "lastName", "citizenId", dob, blood, gender, "mobileNo", address, photo, email, usertype, active, "createdAt", "updatedAt", "createdBy", "updatedBy"
		FROM public.cust_customers WHERE "mobileNo"=$1 AND "orgId"=$2`

	var u model.Customer
	logger.Debug(`Query`, zap.String("query", query))
	err := conn.QueryRow(ctx, query, phoneNo, orgId).Scan(
		&u.ID,
		&u.OrgID,
		&u.DisplayName,
		&u.Title,
		&u.FirstName,
		&u.MiddleName,
		&u.LastName,
		&u.CitizenID,
		&u.DOB,
		&u.Blood,
		&u.Gender,
		&u.MobileNo,
		&u.Address,
		&u.Photo,
		&u.Email,
		&u.UserType,
		&u.Active,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.CreatedBy,
		&u.UpdatedBy,
	)
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
			txtId, "", "Customer", "CustomerByPhoneNo", "",
			"search", -1, start_time, GetQueryParams(c), response, "Query failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Data:   u,
		Desc:   "",
	}
	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, "", "Customer", "CustomerByPhoneNo", "",
		"search", 0, start_time, GetQueryParams(c), response, "GetCustomerByPhoneNo Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)

}

// @summary Create Customer
// @tags Customer
// @security ApiKeyAuth
// @id Create Customer
// @accept json
// @produce json
// @param Case body model.CustomerInsert true "Customer to be created"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/customer/add [post]
func CustomerAdd(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)

	id := c.Param("id")
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()

	var req model.CustomerInsert
	if err := c.ShouldBindJSON(&req); err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Customer", "CustomerAdd", "",
			"create", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusBadRequest, response)
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	now := time.Now()
	query := `
		INSERT INTO public.cust_customers(
	"orgId", "displayName", title, "firstName", "middleName", "lastName", "citizenId", dob, blood, gender, "mobileNo", address, photo, email, usertype, active, "createdAt", "updatedAt", "createdBy", "updatedBy")
	VALUES (
		$1, $2, $3, $4, $5, $6, $7,
		$8, $9, $10, $11, $12, $13, $14,
		$15, $16, $17, $18, $19, $20
	)
	RETURNING id;
	`
	logger.Debug(`Query`, zap.String("query", query))
	logger.Debug(`request input`, zap.Any("Input", []any{req}))
	_, err := conn.Exec(ctx, query,
		orgId, req.DisplayName, req.Title, req.FirstName, req.MiddleName,
		req.LastName, req.CitizenID, req.DOB, req.Blood,
		req.Gender, req.MobileNo, req.Address, req.Photo, req.Email, req.UserType,
		req.Active, now, now, username, username,
	)

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
			txtId, id, "Customer", "CustomerAdd", "",
			"create", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusUnauthorized, response)
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
		txtId, id, "Customer", "CustomerAdd", "",
		"create", 0, start_time, GetQueryParams(c), response, "Add Customer Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)
}

// @summary Update Customer
// @tags Customer
// @security ApiKeyAuth
// @id Update Customer
// @accept json
// @produce json
// @Param id path int true "id"
// @param Body body model.CustomerUpdate true "Data Update"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/customer/{id} [patch]
func CustomerUpdate(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	id := c.Param("id")
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()
	var req model.CustomerUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Customer", "CustomerUpdate", "",
			"update", -1, start_time, GetQueryParams(c), response, "Failed = "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	now := time.Now()
	query := `
	UPDATE public.cust_customers
	SET "displayName"=$1, title=$2, "firstName"=$3, "middleName"=$4, "lastName"=$5, "citizenId"=$6,
	 dob=$7, blood=$8, gender=$9, "mobileNo"=$10, address=$11, photo=$12, email=$13, usertype=$14, active=$15,
	  "updatedAt"=$16, "updatedBy"=$17
	WHERE id = $18 AND "orgId"=$19`

	logger.Debug(`Query`, zap.String("query", query))
	logger.Debug(`request input`, zap.Any("Input", []any{req}))
	_, err := conn.Exec(ctx, query,
		req.DisplayName, req.Title, req.FirstName, req.MiddleName,
		req.LastName, req.CitizenID, req.DOB, req.Blood,
		req.Gender, req.MobileNo, req.Address, req.Photo, req.Email, req.UserType,
		req.Active, now, username, id, orgId,
	)

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
			txtId, id, "Customer", "CustomerUpdate", "",
			"update", -1, start_time, GetQueryParams(c), response, "Failed = "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusUnauthorized, response)
		logger.Warn("Insert failed", zap.Error(err))
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
		txtId, id, "Customer", "CustomerUpdate", "",
		"update", 0, start_time, GetQueryParams(c), response, "Custommer Update Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)
}

// @summary Delete Customer
// @tags Customer
// @security ApiKeyAuth
// @id Delete Customer
// @accept json
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/customer/{id} [delete]
func CustomerDelete(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	id := c.Param("id")
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()
	query := `DELETE FROM public."cust_customers" WHERE id = $1 AND "orgId"=$2`
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
			txtId, id, "Customer", "CustomerDelete", "",
			"delete", -1, start_time, GetQueryParams(c), response, "Update failed : "+err.Error(),
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
		txtId, id, "Customer", "CustomerDelete", "",
		"delete", -1, start_time, GetQueryParams(c), response, "Delete Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)
}

// @summary Get Customer with Social
// @tags Customer
// @security ApiKeyAuth
// @id Get Customer with Social
// @accept json
// @produce json
// @Param start query int false "start" default(0)
// @Param length query int false "length" default(10)
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/customer_with_socials [get]
func CustomerSocialList(c *gin.Context) {
	logger := utils.GetLog()
	orgId := GetVariableFromToken(c, "orgId")
	id := c.Param("id")
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
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
	logger.Debug(`Query`, zap.Any("start", start))
	logger.Debug(`Query`, zap.Any("length", length))
	query := `SELECT id, "orgId", "custId", "socialType", "socialId", "socialName", "imgUrl", "createdAt", "updatedAt", "createdBy", "updatedBy"
		FROM public.cust_customer_with_socials 
		WHERE "orgId"=$1 LIMIT $2 OFFSET $3`

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
			txtId, id, "Customer", "CustomerSocialList", "",
			"search", -1, start_time, GetQueryParams(c), response, "Query failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer rows.Close()
	var errorMsg string
	var u model.CustomerSocial
	var userList []model.CustomerSocial
	found := false
	rowIndex := 0
	for rows.Next() {
		rowIndex++
		err := rows.Scan(
			&u.ID,
			&u.OrgID,
			&u.CustID,
			&u.SocialType,
			&u.SocialID,
			&u.SocialName,
			&u.ImgURL,
			&u.CreatedAt,
			&u.UpdatedAt,
			&u.CreatedBy,
			&u.UpdatedBy,
		)
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
				txtId, id, "Customer", "CustomerSocialList", "",
				"search", -1, start_time, GetQueryParams(c), response, "Scan failed : "+err.Error(),
			)
			//=======AUDIT_END=====//
			c.JSON(http.StatusInternalServerError, response)
			return
		}
		userList = append(userList, u)
		found = true
	}
	if !found {
		response := model.Response{
			Status: "-1",
			Msg:    "Failed",
			Desc:   "Not found",
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Customer", "CustomerSocialList", "",
			"search", -1, start_time, GetQueryParams(c), response, "Not Found.",
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
	} else {
		response := model.Response{
			Status: "0",
			Msg:    "Success",
			Data:   userList,
			Desc:   "",
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Customer", "CustomerSocialList", "",
			"search", 0, start_time, GetQueryParams(c), response, "Get CustomerSocialList Success.",
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusOK, response)
	}
}

// @summary Get Customer with Social by Id
// @tags Customer
// @security ApiKeyAuth
// @id Get Customer with Social by Id
// @accept json
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/customer_with_socials/{id} [get]
func CustomerWithSocialById(c *gin.Context) {
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
	query := `SELECT id, "orgId", "custId", "socialType", "socialId", "socialName", "imgUrl", "createdAt", "updatedAt", "createdBy", "updatedBy"
		FROM public.cust_customer_with_socials  
		WHERE id=$1 AND "orgId"=$2`

	var u model.CustomerSocial
	logger.Debug(`Query`, zap.String("query", query))
	err := conn.QueryRow(ctx, query, id, orgId).Scan(
		&u.ID,
		&u.OrgID,
		&u.CustID,
		&u.SocialType,
		&u.SocialID,
		&u.SocialName,
		&u.ImgURL,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.CreatedBy,
		&u.UpdatedBy,
	)
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
			txtId, "", "Station", "CustomerWithSocialById", "",
			"search", -1, start_time, GetQueryParams(c), response, "Query failed : "+err.Error(),
		)
		//=======AUDIT_END=====//

		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Data:   u,
		Desc:   "",
	}
	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, "", "Station", "CustomerWithSocialById", "",
		"search", 0, start_time, GetQueryParams(c), response, "Get CustomerWithSocialById Success",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)

}

// @summary Create Customer with Social
// @tags Customer
// @security ApiKeyAuth
// @id Create Customer with Social
// @accept json
// @produce json
// @param Body body model.CustomerSocialInsert true "Data to be created"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/customer_with_socials/add [post]
func CustomerSocialAdd(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	id := c.Param("id")
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()
	var req model.CustomerSocialInsert
	if err := c.ShouldBindJSON(&req); err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Customer", "CustomerSocialAdd", "",
			"create", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusBadRequest, response)
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	now := time.Now()
	query := `
		INSERT INTO public.cust_customer_with_socials(
	 "orgId", "custId", "socialType", "socialId", "socialName", "imgUrl", "createdAt", "updatedAt", "createdBy", "updatedBy")
	VALUES (
		$1, $2, $3, $4, $5, $6, $7,
		$8, $9, $10
	)
	`
	logger.Debug(`Query`, zap.String("query", query))
	logger.Debug(`request input`, zap.Any("Input", []any{req}))
	_, err := conn.Exec(ctx, query,
		orgId, req.CustID, req.SocialType, req.SocialID, req.SocialName,
		req.ImgURL, now, now, username, username,
	)

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
			txtId, id, "Customer", "CustomerSocialAdd", "",
			"create", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusUnauthorized, response)
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
		txtId, id, "Customer", "CustomerSocialAdd", "",
		"create", 0, start_time, GetQueryParams(c), response, "CustommerSociallAdd Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)
}

// @summary Update Customer with Social
// @tags Customer
// @security ApiKeyAuth
// @id Update Customer with Social
// @accept json
// @produce json
// @Param id path int true "id"
// @param Body body model.CustomerSocialUpdate true "Data Update"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/customer_with_socials/{id} [patch]
func CustomerSocialUpdate(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	id := c.Param("id")
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()

	var req model.CustomerSocialUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Customer", "CustomerSocialUpdate", "",
			"update", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	now := time.Now()
	query := `
	UPDATE public.cust_customer_with_socials
	SET "custId"=$3, socialType=$4, "socialId"=$5, "socialName"=$6, "imgUrl"=$7,
	  "updatedAt"=$8, "updatedBy"=$9
	WHERE id = $1 AND "orgId"=$2`

	logger.Debug(`Query`, zap.String("query", query))
	logger.Debug(`request input`, zap.Any("Input", []any{req}))
	_, err := conn.Exec(ctx, query, id, orgId,
		req.CustID, req.SocialType, req.SocialID, req.SocialName, req.ImgURL, now, username,
	)

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
			txtId, id, "Customer", "CustomerSocialUpdate", "",
			"update", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusUnauthorized, response)
		logger.Warn("Insert failed", zap.Error(err))
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
		txtId, id, "Customer", "CustomerSocialUpdate", "",
		"update", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)
}

// @summary Delete Customer with Social
// @tags Customer
// @security ApiKeyAuth
// @id Delete Customer with Social
// @accept json
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/customer_with_socials/{id} [delete]
func CustomerSocialDelete(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	id := c.Param("id")
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()
	query := `DELETE FROM public."cust_customer_with_socials" WHERE id = $1 AND "orgId"=$2`
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
			txtId, id, "Customer", "CustomerSocialDelete", "",
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
		txtId, id, "Customer", "CustomerSocialDelete", "",
		"delete", 0, start_time, GetQueryParams(c), response, "CustomerSocialDelete Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)
}

// @summary Get Customer Contact
// @tags Customer
// @security ApiKeyAuth
// @id Get Customer Contact
// @accept json
// @produce json
// @Param start query int false "start" default(0)
// @Param length query int false "length" default(10)
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/customer_contacts [get]
func CustomerContactList(c *gin.Context) {
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
	logger.Debug(`Query`, zap.Any("start", start))
	logger.Debug(`Query`, zap.Any("length", length))
	query := `SELECT id, "orgId", "custId", "contactName", "contactPhone", "contactAddr", "createdAt", "updatedAt", "createdBy", "updatedBy"
		FROM public.cust_contacts
		WHERE "orgId"=$1 LIMIT $2 OFFSET $3`

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
	var u model.CustomerContact
	var userList []model.CustomerContact
	found := false
	rowIndex := 0
	for rows.Next() {
		rowIndex++
		err := rows.Scan(
			&u.ID,
			&u.OrgID,
			&u.CustID,
			&u.ContactName,
			&u.ContactPhone,
			&u.ContactAddr,
			&u.CreatedAt,
			&u.UpdatedAt,
			&u.CreatedBy,
			&u.UpdatedBy,
		)
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
				txtId, id, "Customer", "CustomerContactList", "",
				"search", -1, start_time, GetQueryParams(c), response, "Scan failed = "+err.Error(),
			)
			//=======AUDIT_END=====//
			c.JSON(http.StatusInternalServerError, response)
			return
		}
		userList = append(userList, u)
		found = true
	}
	if !found {
		response := model.Response{
			Status: "-1",
			Msg:    "Failed",
			Desc:   "Not found",
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Customer", "CustomerContactList", "",
			"search", -1, start_time, GetQueryParams(c), response, "Not Found.",
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
	} else {
		response := model.Response{
			Status: "0",
			Msg:    "Success",
			Data:   userList,
			Desc:   "",
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Customer", "CustomerContactList", "",
			"search", 0, start_time, GetQueryParams(c), response, "CustomerContactList Success.",
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusOK, response)
	}
}

// @summary Get Customer Contact by Id
// @tags Customer
// @security ApiKeyAuth
// @id Get Customer Contact by Id
// @accept json
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/customer_contacts/{id} [get]
func CustomerContactById(c *gin.Context) {
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
	query := `SELECT id, "orgId", "custId", "contactName", "contactPhone", "contactAddr", "createdAt", "updatedAt", "createdBy", "updatedBy"
		FROM public.cust_contacts 
		WHERE id=$1 AND "orgId"=$2`

	var u model.CustomerContact
	logger.Debug(`Query`, zap.String("query", query))
	err := conn.QueryRow(ctx, query, id, orgId).Scan(
		&u.ID,
		&u.OrgID,
		&u.CustID,
		&u.ContactName,
		&u.ContactPhone,
		&u.ContactAddr,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.CreatedBy,
		&u.UpdatedBy,
	)
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
			txtId, id, "Customer", "CustomerContactById", "",
			"search", -1, start_time, GetQueryParams(c), response, "Query failed = "+err.Error(),
		)
		//=======AUDIT_END=====//

		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Data:   u,
		Desc:   "",
	}
	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, id, "Customer", "CustomerContactById", "",
		"search", 0, start_time, GetQueryParams(c), response, "GetCustomerContactById Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)

}

// @summary Create Customer Contact
// @tags Customer
// @security ApiKeyAuth
// @id Create Customer Contact
// @accept json
// @produce json
// @param Body body model.CustomerContactInsert true "Data to be created"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/customer_contacts/add [post]
func CustomerContactAdd(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	id := c.Param("id")
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()

	var req model.CustomerContactInsert
	if err := c.ShouldBindJSON(&req); err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Customer", "CustomerContactAdd", "",
			"create", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusBadRequest, response)
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	now := time.Now()
	query := `
		INSERT INTO public.cust_contacts(
	"orgId", "custId", "contactName", "contactPhone", "contactAddr", "createdAt", "updatedAt", "createdBy", "updatedBy")
	VALUES ($1, $2, $3, $4, $5, $6, $7,$8, $9)`

	logger.Debug(`Query`, zap.String("query", query))
	logger.Debug(`request input`, zap.Any("Input", []any{req}))
	_, err := conn.Exec(ctx, query,
		orgId, req.CustID, req.ContactName, req.ContactPhone, req.ContactAddr, now, now, username, username,
	)

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
			txtId, id, "Customer", "CustomerContactAdd", "",
			"create", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusUnauthorized, response)
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
		txtId, id, "Customer", "CustomerContactAdd", "",
		"create", 0, start_time, GetQueryParams(c), response, "CustomerContactAdd Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)
}

// @summary Update Customer Contact
// @tags Customer
// @security ApiKeyAuth
// @id Update Customer Contact
// @accept json
// @produce json
// @Param id path int true "id"
// @param Body body model.CustomerContactUpdate true "Data Update"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/customer_contacts/{id} [patch]
func CustomerContactUpdate(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	id := c.Param("id")
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()

	var req model.CustomerContactUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Customer", "GetCustomerContactUpdate", "",
			"update", -1, start_time, GetQueryParams(c), response, "Failure : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	now := time.Now()
	query := `
	UPDATE public.cust_customers
	SET "custId"=$3, "contactName"=$4, "contactPhone"=$5, "contactAddr"=$6,
	  "updatedAt"=$7, "updatedBy"=$8
		WHERE id = $1 AND "orgId"=$2`

	logger.Debug(`Query`, zap.String("query", query))
	logger.Debug(`request input`, zap.Any("Input", []any{req}))
	_, err := conn.Exec(ctx, query, id, orgId,
		req.CustID, req.ContactName, req.ContactPhone, req.ContactAddr,
		now, username,
	)

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
			txtId, id, "Customer", "GetCustomerContactUpdate", "",
			"update", -1, start_time, GetQueryParams(c), response, "Failure : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusUnauthorized, response)
		logger.Warn("Insert failed", zap.Error(err))
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
		txtId, id, "Customer", "GetCustomerContactUpdate", "",
		"update", 0, start_time, GetQueryParams(c), response, "CustomerContactUpdate Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)
}

// @summary Delete Customer Contact
// @tags Customer
// @security ApiKeyAuth
// @id Delete Customer Contact
// @accept json
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/customer_contacts/{id} [delete]
func CustomerContactDelete(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	id := c.Param("id")
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()
	query := `DELETE FROM public."cust_contacts" WHERE id = $1 AND "orgId"=$2`
	logger.Debug("Query", zap.String("query", query), zap.Any("id", id))
	_, err := conn.Exec(ctx, query, id, orgId)
	if err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Customer", "CustomerContactDelete", "",
			"delete", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
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
		txtId, id, "Customer", "CustomerContactDelete", "",
		"delete", 0, start_time, GetQueryParams(c), response, "CustomerContactDete Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)
}
