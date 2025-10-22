package handler

import (
	"crypto/subtle"
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

// @summary Get User
// @tags User
// @security ApiKeyAuth
// @id Get User
// @accept json
// @produce json
// @Param start query int false "start" default(0)
// @Param length query int false "length" default(10)
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/users [get]
func GetUmUserList(c *gin.Context) {
	logger := utils.GetLog()
	orgId := GetVariableFromToken(c, "orgId")
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	id := c.Param("id")
	now := time.Now()
	username := GetVariableFromToken(c, "username")
	txtId := uuid.New().String()
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
	query := `SELECT "id","orgId", "displayName", title, "firstName", "middleName", "lastName", "citizenId", bod,
	blood, gender, "mobileNo", address, photo, username, password, email, "roleId", "userType", "empId",
	"deptId", "commId", "stnId", active, "activationToken", "lastActivationRequest", "lostPasswordRequest",
	"signupStamp", islogin, "lastLogin", "createdAt", "updatedAt", "createdBy", "updatedBy" 
	FROM public.um_users 
	WHERE "orgId"=$1 
	ORDER BY "firstName" ASC, "lastName" ASC 
	LIMIT $2 OFFSET $3 `

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
			txtId, id, "um_user", "GetUmUserList", "",
			"view", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer rows.Close()
	var errorMsg string
	var u model.Um_User
	var userList []model.Um_User
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
			&u.Bod,
			&u.Blood,
			&u.Gender,
			&u.MobileNo,
			&u.Address,
			&u.Photo,
			&u.Username,
			&u.Password,
			&u.Email,
			&u.RoleID,
			&u.UserType,
			&u.EmpID,
			&u.DeptID,
			&u.CommID,
			&u.StnID,
			&u.Active,
			&u.ActivationToken,
			&u.LastActivationRequest,
			&u.LostPasswordRequest,
			&u.SignupStamp,
			&u.IsLogin,
			&u.LastLogin,
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
				txtId, id, "um_user", "GetUmUserList", "",
				"view", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
			)
			//=======AUDIT_END=====//
			c.JSON(http.StatusInternalServerError, response)
			return
		}
		userList = append(userList, u)
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
			txtId, id, "um_user", "GetUmUserList", "",
			"view", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
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
			txtId, id, "um_user", "GetUmUserList", "",
			"view", 0, now, GetQueryParams(c), response, "GetUmUserList Success.",
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusOK, response)
	}
}

// @summary Get User by username
// @tags User
// @security ApiKeyAuth
// @id Get User by username
// @accept json
// @produce json
// @Param username path string true "username"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/users/username/{username} [get]
func GetUmUserByUsername(c *gin.Context) {
	logger := utils.GetLog()
	id := c.Param("id")
	now := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	query := `
	SELECT t1."id",t1."orgId", t1."displayName", t1.title, t1."firstName", t1."middleName", t1."lastName",
		t1."citizenId", t1.bod, t1.blood, t1.gender, t1."mobileNo", t1.address, t1.photo, t1.username, t1.password,
		t1.email, t1."roleId", t1."userType", t1."empId", t1."deptId", t1."commId", t1."stnId", t1.active,
		t1."activationToken",t1."lastActivationRequest", t1."lostPasswordRequest", t1."signupStamp",
		t1.islogin, t1."lastLogin",t1."createdAt", t1."updatedAt", t1."createdBy", t1."updatedBy",t2.name,t3."roleName"
		FROM public.um_users t1
		JOIN public.organizations t2 ON t1."orgId" = t2.id
		JOIN public.um_roles t3 ON t1."roleId" = t3.id
		WHERE t1.username=$1 AND t1."orgId"=$2`

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query),
		zap.Any("Input", []any{
			username, orgId,
		}))
	rows, err := conn.Query(ctx, query, username, orgId)
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
			txtId, id, "um_user", "GetUmUserByUsername", "",
			"view", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer rows.Close()
	var errorMsg string
	var u model.Um_User
	if rows.Next() {
		err = rows.Scan(&u.ID,
			&u.OrgID, &u.DisplayName, &u.Title, &u.FirstName, &u.MiddleName, &u.LastName, &u.CitizenID, &u.Bod, &u.Blood,
			&u.Gender, &u.MobileNo, &u.Address, &u.Photo, &u.Username, &u.Password, &u.Email, &u.RoleID, &u.UserType,
			&u.EmpID, &u.DeptID, &u.CommID, &u.StnID, &u.Active, &u.ActivationToken, &u.LastActivationRequest,
			&u.LostPasswordRequest, &u.SignupStamp, &u.IsLogin, &u.LastLogin, &u.CreatedAt, &u.UpdatedAt,
			&u.CreatedBy, &u.UpdatedBy, &u.OrgName, &u.RoleName,
		)
		if err != nil {
			errorMsg = err.Error()
			logger.Warn("Scan failed", zap.Error(err))
			response := model.Response{
				Status: "-1",
				Msg:    "Failed",
				Desc:   errorMsg,
			}
			//=======AUDIT_START=====//
			_ = utils.InsertAuditLogs(
				c, conn, orgId.(string), username.(string),
				txtId, id, "um_user", "GetUmUserByUsername", "",
				"view", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
			)
			//=======AUDIT_END=====//
			c.JSON(http.StatusInternalServerError, response)
			return
		}
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
		txtId, id, "um_user", "GetUmUserByUsername", "",
		"view", 0, now, GetQueryParams(c), response, "GetUmUserByUsername Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)
}

// @summary Get User by Id
// @tags User
// @security ApiKeyAuth
// @id Get User by Id
// @accept json
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/users/{id} [get]
func GetUmUserById(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	id := c.Param("id")
	now := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()
	defer cancel()
	defer conn.Close(ctx)

	// Step 1: Get user profile
	queryUser := `
		SELECT "orgId", "displayName", title, "firstName", "middleName", "lastName", "citizenId", bod, blood, gender, 
		       "mobileNo", address, photo, username, password, email, "roleId", "userType", "empId", "deptId", "commId", 
		       "stnId", active, "activationToken", "lastActivationRequest", "lostPasswordRequest", "signupStamp", 
		       islogin, "lastLogin", "createdAt", "updatedAt", "createdBy", "updatedBy"
		FROM public.um_users 
		WHERE id=$1 AND "orgId"=$2
	`
	var u model.Um_User
	err := conn.QueryRow(ctx, queryUser, id, orgId).Scan(
		&u.OrgID, &u.DisplayName, &u.Title, &u.FirstName, &u.MiddleName, &u.LastName, &u.CitizenID,
		&u.Bod, &u.Blood, &u.Gender, &u.MobileNo, &u.Address, &u.Photo, &u.Username, &u.Password, &u.Email,
		&u.RoleID, &u.UserType, &u.EmpID, &u.DeptID, &u.CommID, &u.StnID, &u.Active, &u.ActivationToken,
		&u.LastActivationRequest, &u.LostPasswordRequest, &u.SignupStamp, &u.IsLogin, &u.LastLogin,
		&u.CreatedAt, &u.UpdatedAt, &u.CreatedBy, &u.UpdatedBy,
	)
	if err != nil {
		logger.Warn("Query user failed", zap.Error(err))
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "um_user", "GetUmUserByUserById", "",
			"search", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Step 2: Get skills
	querySkills := `
		SELECT s."skillId", s."en", s."th"
		FROM public.um_user_with_skills us
		JOIN public.um_skills s ON us."skillId" = s."skillId"
		WHERE us."orgId" = $1 AND us."userName" = $2 AND us.active = TRUE
	`
	rows, err := conn.Query(ctx, querySkills, orgId, u.Username)
	if err != nil {
		logger.Warn("Query skills failed", zap.Error(err))
	}
	defer rows.Close()

	var skills []map[string]interface{}
	for rows.Next() {
		var skillId, en, th string

		if err := rows.Scan(&skillId, &en, &th); err == nil {

			skills = append(skills, map[string]interface{}{
				"skillId": skillId,
				"en":      en,
				"th":      th,
			})
		}
	}

	// Step 3: Get areas from um_user_with_area_response
	queryAreas := `
		SELECT
		ad."distId",
		ad."th",
		ad."en"
		FROM "um_user_with_area_response" ua
		LEFT JOIN "area_districts" ad ON ad."orgId" = ua."orgId"
		AND ad."distId" IN (
			SELECT jsonb_array_elements_text(ua."distIdLists")
		)
		WHERE ua."orgId" = $1 AND ua."username" = $2 
	`

	areaRows, err := conn.Query(ctx, queryAreas, orgId, u.Username)
	if err != nil {
		logger.Warn("Query areas failed", zap.Error(err))
	}
	defer areaRows.Close()

	var areas []map[string]interface{}
	for areaRows.Next() {
		var distId, th, en string
		if err := areaRows.Scan(&distId, &th, &en); err == nil {
			areas = append(areas, map[string]interface{}{
				"distId": distId,
				"th":     th,
				"en":     en,
			})
		}
	}

	// Step 4: Combine into response
	u.Skills = &skills
	u.Areas = &areas
	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Data:   u,
		Desc:   "",
	}
	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, id, "um_user", "GetUmUserByUserById", "",
		"search", 0, now, GetQueryParams(c), response, "GetUmUserByUserById Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)
}

// @summary Create User
// @tags User
// @security ApiKeyAuth
// @id Create User
// @accept json
// @produce json
// @param Case body model.UserInput true "User to be created"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/users/add [post]
func UserAdd(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)

	var req model.UserInput
	now := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()
	tokenString := GetVariableFromToken(c, "tokenString")
	if err := c.ShouldBindJSON(&req); err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, "", "um_user", "UserAdd", "",
			"create", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusBadRequest, response)
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	// now req is ready to use

	var enc string
	var err error
	var id int
	enc, err = encrypt(req.Password)

	if err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, "", "um_user", "UserAdd", "",
			"create", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusBadRequest, response)
		return
	}

	query := `
		INSERT INTO public.um_users(
		"orgId", "displayName", title, "firstName", "middleName", "lastName", "citizenId", bod, blood, gender,
		"mobileNo", address, photo, username, password, email, "roleId", "userType", "empId", "deptId",
		"commId", "stnId", active, "activationToken", "lastActivationRequest", "lostPasswordRequest",
		"signupStamp", islogin, "lastLogin", "createdAt", "updatedAt", "createdBy", "updatedBy"
			)
	VALUES (
		$1, $2, $3, $4, $5, $6, $7,
		$8, $9, $10, $11,
		$12, $13, $14, $15, $16,
		$17, $18, $19, $20, $21, 
		$22, $23, $24, $25, $26,
		$27, $28, $29, $30, $31, 
		$32, $33
	)
	RETURNING id;
	`
	logger.Debug(`Query`, zap.String("query", query))
	logger.Debug(`request input`, zap.Any("Input", []any{req}))
	logger.Debug(`Encrypt Password :` + enc)
	err = conn.QueryRow(ctx, query,
		orgId, req.DisplayName, req.Title, req.FirstName, req.MiddleName,
		req.LastName, req.CitizenID, req.Bod, req.Blood,
		req.Gender, req.MobileNo, req.Address, req.Photo, req.Username,
		enc, req.Email, req.RoleID, req.UserType, req.EmpID, req.DeptID, req.CommID, req.StnID,
		req.Active, tokenString, req.LastActivationRequest, req.LostPasswordRequest, req.SignupStamp,
		req.IsLogin, req.LastLogin, now, now, username, username,
	).Scan(&id)

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
			txtId, "", "um_user", "UserAdd", "",
			"create", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
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
		txtId, "", "um_user", "UserAdd", "",
		"create", 0, now, GetQueryParams(c), response, "UserAdd Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)
}

// @summary Update User
// @tags User
// @security ApiKeyAuth
// @id Update User
// @accept json
// @produce json
// @Param id path int true "id"
// @param Body body model.UserUpdate true "Data Update"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/users/{id} [patch]
func UserUpdate(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)

	var req model.UserUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	// now req is ready to use

	var err error
	id := c.Param("id")
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()
	now := time.Now()
	query := `
	UPDATE public.um_users
	SET "displayName"=$1, title=$2, "firstName"=$3, "middleName"=$4, "lastName"=$5, "citizenId"=$6,
	bod=$7, blood=$8, gender=$9, "mobileNo"=$10, address=$11, photo=$12, username=$13, email=$14, "roleId"=$15,
	"userType"=$16, "empId"=$17, "deptId"=$18, "commId"=$19, "stnId"=$20, active=$21,
	"lastActivationRequest"=$22, "lostPasswordRequest"=$23, "signupStamp"=$24, islogin=$25, "lastLogin"=$26,
	"updatedAt"=$27,"updatedBy"=$28 WHERE id = $29 AND "orgId"=$30`

	logger.Debug(`Query`, zap.String("query", query))
	logger.Debug(`request input`, zap.Any("Input", []any{req}))
	_, err = conn.Exec(ctx, query,
		req.DisplayName, req.Title, req.FirstName, req.MiddleName,
		req.LastName, req.CitizenID, req.Bod, req.Blood,
		req.Gender, req.MobileNo, req.Address, req.Photo, req.Username,
		req.Email, req.RoleID, req.UserType, req.EmpID, req.DeptID, req.CommID, req.StnID,
		req.Active, req.LastActivationRequest, req.LostPasswordRequest, req.SignupStamp,
		req.IsLogin, req.LastLogin, now, username, id, orgId,
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
			txtId, id, "um_user", "UserUpdate", "",
			"update", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
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
		txtId, id, "um_user", "UserUpdate", "",
		"update", 0, now, GetQueryParams(c), response, "UserUpdate Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)
}

// @summary Update User By Username
// @tags User
// @security ApiKeyAuth
// @id Update User By Username
// @accept json
// @produce json
// @Param username path string true "username"
// @param Body body model.UserUpdate true "Data Update"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/users/username/{username} [patch]
func UserUpdateByUsername(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	id := c.Param("id")
	now := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()
	var req model.UserUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "um_user", "UserUpdateByUsername", "",
			"update", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	// now req is ready to use

	var err error
	query := `
	UPDATE public.um_users
	SET "displayName"=$1, title=$2, "firstName"=$3, "middleName"=$4, "lastName"=$5, "citizenId"=$6,
	bod=$7, blood=$8, gender=$9, "mobileNo"=$10, address=$11, photo=$12, username=$13, email=$14, "roleId"=$15,
	"userType"=$16, "empId"=$17, "deptId"=$18, "commId"=$19, "stnId"=$20, active=$21,
	"lastActivationRequest"=$22, "lostPasswordRequest"=$23, "signupStamp"=$24, islogin=$25, "lastLogin"=$26,
	"updatedAt"=$27,"updatedBy"=$28 WHERE username = $29 AND "orgId"=$30`

	logger.Debug(`Query`, zap.String("query", query))
	logger.Debug(`request input`, zap.Any("Input", []any{req}))
	_, err = conn.Exec(ctx, query,
		req.DisplayName, req.Title, req.FirstName, req.MiddleName,
		req.LastName, req.CitizenID, req.Bod, req.Blood,
		req.Gender, req.MobileNo, req.Address, req.Photo, req.Username,
		req.Email, req.RoleID, req.UserType, req.EmpID, req.DeptID, req.CommID, req.StnID,
		req.Active, req.LastActivationRequest, req.LostPasswordRequest, req.SignupStamp,
		req.IsLogin, req.LastLogin, now, username, id, orgId,
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
			txtId, id, "um_user", "UserUpdateByUsername", "",
			"update", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusUnauthorized, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
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
		txtId, id, "um_user", "UserUpdateByUsername", "",
		"update", 0, now, GetQueryParams(c), response, "UserUpdateByUsername Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)
}

// @summary Delete User
// @tags User
// @security ApiKeyAuth
// @id Delete User
// @accept json
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/users/{id} [delete]
func UserDelete(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	id := c.Param("id")
	now := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()
	query := `DELETE FROM public."um_users" WHERE id = $1 AND "orgId"=$2`
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
			txtId, id, "um_user", "UserDelete", "",
			"delete", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
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
		txtId, id, "um_user", "UserDelete", "",
		"delete", 0, now, GetQueryParams(c), response, "UserDelete Success",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)
}

// @summary Get User with skills
// @tags User
// @security ApiKeyAuth
// @id Get User with skills
// @accept json
// @produce json
// @Param start query int false "start" default(0)
// @Param length query int false "length" default(10)
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/users_with_skills [get]
func GetUserWithSkills(c *gin.Context) {
	logger := utils.GetLog()
	id := c.Param("id")
	now := time.Now()
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

	query := `SELECT "orgId", "userName", "skillId", active, "createdAt", "updatedAt", "createdBy", "updatedBy" 
	FROM public.um_user_with_skills WHERE "orgId"=$1 LIMIT $2 OFFSET $3`

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
			txtId, id, "um_user", "GetUserWithSkill", "",
			"view", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer rows.Close()
	var errorMsg string
	var u model.UserSkill
	var userList []model.UserSkill
	rowIndex := 0
	for rows.Next() {
		rowIndex++
		err := rows.Scan(
			&u.OrgID,
			&u.UserName,
			&u.SkillID,
			&u.Active,
			&u.CreatedAt,
			&u.UpdatedAt,
			&u.CreatedBy,
			&u.UpdatedBy,
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
				txtId, id, "um_user", "GetUserWithSkill", "",
				"view", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
			)
			//=======AUDIT_END=====//
			c.JSON(http.StatusInternalServerError, response)
		}
		userList = append(userList, u)
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
			txtId, id, "um_user", "GetUserWithSkill", "",
			"view", -1, now, GetQueryParams(c), response, "Failed : "+errorMsg,
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
			txtId, id, "um_user", "GetUserWithSkill", "",
			"view", 0, now, GetQueryParams(c), response, "GetUserWithSkills Success.",
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusOK, response)
	}
}

// @summary Get User with skills by id
// @tags User
// @security ApiKeyAuth
// @id Get User with skills by id
// @Param id path int true "id"
// @accept json
// @produce json
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/users_with_skills/{id} [get]
func GetUserWithSkillsById(c *gin.Context) {
	logger := utils.GetLog()
	id := c.Param("id")
	now := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	query := `SELECT "orgId", "userName", "skillId", active, "createdAt", "updatedAt", "createdBy", "updatedBy" 
	FROM public.um_user_with_skills WHERE id=$1 AND "orgId"=$2`

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
			txtId, id, "um_user", "GetUserWithSkillsById", "",
			"search", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer rows.Close()
	var u model.UserSkill
	err = rows.Scan(
		&u.OrgID,
		&u.UserName,
		&u.SkillID,
		&u.Active,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.CreatedBy,
		&u.UpdatedBy,
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
			txtId, id, "um_user", "GetUserWithSkillsById", "",
			"search", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
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
		txtId, id, "um_user", "GetUserWithSkillsById", "",
		"search", 0, now, GetQueryParams(c), response, "GetUserWithSkillsById Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)

}

// @summary Get User with skills by skill id
// @tags User
// @security ApiKeyAuth
// @id Get User with skills by skillId
// @Param skillId path string true "skillId"
// @accept json
// @produce json
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/users_with_skills/skillId/{skillId} [get]
func GetUserWithSkillsBySkillId(c *gin.Context) {
	logger := utils.GetLog()
	skillId := c.Param("skillId")
	id := c.Param("id")
	now := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	query := `SELECT "orgId", "userName", "skillId", active, "createdAt", "updatedAt", "createdBy", "updatedBy" 
	FROM public.um_user_with_skills WHERE "skillId" = $1 AND "orgId" = $2`

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query))
	rows, err := conn.Query(ctx, query, skillId, orgId)
	var userList []model.UserSkill
	var u model.UserSkill
	var errorMsg string
	rowIndex := 0
	for rows.Next() {
		rowIndex++
		err := rows.Scan(
			&u.OrgID,
			&u.UserName,
			&u.SkillID,
			&u.Active,
			&u.CreatedAt,
			&u.UpdatedAt,
			&u.CreatedBy,
			&u.UpdatedBy,
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
				txtId, id, "um_user", "GetUserWithSkillsBySkillId", "",
				"search", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
			)
			//=======AUDIT_END=====//
			c.JSON(http.StatusInternalServerError, response)
		}
		userList = append(userList, u)
	}
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
			txtId, id, "um_user", "GetUserWithSkillsBySkillId", "",
			"search", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer rows.Close()
	// 	logger.Warn("Scan failed", zap.Error(err))
	// 	response := model.Response{
	// 		Status: "-1",
	// 		Msg:    "Failed",
	// 		Desc:   err.Error(),
	// 	}
	// 	c.JSON(http.StatusInternalServerError, response)
	// 	return

	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Data:   userList,
		Desc:   "",
	}
	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, id, "um_user", "GetUserWithSkillsBySkillId", "",
		"search", 0, now, GetQueryParams(c), response, "GetUserWithSkillsBySkillId.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)

}

// @summary Create User with skill
// @id Create User with skill
// @security ApiKeyAuth
// @tags User
// @accept json
// @produce json
// @param Body body model.UserSkillInsert true "Create Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/users_with_skills/add [post]
func InsertUserWithSkills(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()
	now := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()
	var req model.UserSkillInsert
	if err := c.ShouldBindJSON(&req); err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, "", "um_user", "InsertUserWithSkills", "",
			"create", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusBadRequest, response)
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	// now req is ready to use
	var id int
	query := `
	INSERT INTO public."um_user_with_skills"(
	"orgId", "userName", "skillId", active, "createdAt", "updatedAt", "createdBy", "updatedBy")
	VALUES ($1, $2, $3, $4, $5, $6, $7,$8)
	RETURNING id ;
	`

	err := conn.QueryRow(ctx, query,
		req.OrgID, req.UserName, req.SkillID, req.Active, now,
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
			txtId, "", "um_user", "InsertUserWithSkills", "",
			"create", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
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
		txtId, "", "um_user", "InsertUserWithSkills", "",
		"create", 0, now, GetQueryParams(c), response, "InsertUserWithSkills Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)

}

// @summary Update User with skill
// @id Update User with skill
// @security ApiKeyAuth
// @accept json
// @produce json
// @tags User
// @Param id path int true "id"
// @param Body body model.UserSkillUpdate true "Update data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/users_with_skills/{id} [patch]
func UpdateUserWithSkills(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	id := c.Param("id")
	now := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()

	var req model.UserSkillUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Update failed", zap.Error(err))
		response := model.UpdateCaseResponse{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
			ID:     ToInt(id),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "um_user", "UpdateUserWithSkills", "",
			"update", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusBadRequest, response)
		return
	}

	query := `UPDATE public."um_user_with_skills"
	SET 
    "skillId"=$2,
    active=$3,
	"updatedAt"=$4,
	"updatedBy"=$5
	WHERE id = $1 `
	_, err := conn.Exec(ctx, query,
		id, req.SkillID, req.Active,
		now, username, username,
	)
	logger.Debug("Update Case SQL Args",
		zap.String("query", query),
		zap.Any("Input", []any{
			id, req.SkillID, req.Active,
			now, username,
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
			txtId, id, "um_user", "UpdateUserWithSkills", "",
			"update", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
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
		txtId, id, "um_user", "UpdateUserWithSkills", "",
		"update", 0, now, GetQueryParams(c), response, "UpdateUserWithSkills Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)
}

// @summary Delete User with skill
// @id Delete User with skill
// @security ApiKeyAuth
// @accept json
// @tags User
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/users_with_skills/{id} [delete]
func DeleteUserWithSkills(c *gin.Context) {

	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	id := c.Param("id")
	now := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()
	query := `DELETE FROM public."um_user_with_skills" WHERE id = $1 `
	logger.Debug("Query", zap.String("query", query), zap.Any("id", id))
	_, err := conn.Exec(ctx, query, id)
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
			txtId, id, "um_user", "DeleteUserWithSkill", "",
			"delete", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
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
		txtId, id, "um_user", "DeleteUserWithSkill", "",
		"delete", 0, now, GetQueryParams(c), response, "DeleteUserWithSkills Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)
}

// @summary Get User with contacts
// @tags User
// @security ApiKeyAuth
// @id Get User with contacts
// @accept json
// @produce json
// @Param start query int false "start" default(0)
// @Param length query int false "length" default(10)
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/users_with_contacts [get]
func GetUserWithContacts(c *gin.Context) {
	logger := utils.GetLog()

	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	id := c.Param("id")
	now := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()
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

	query := `SELECT  "orgId", username, "contactName", "contactPhone", "contactAddr", "createdAt", "updatedAt", "createdBy", "updatedBy" 	
	FROM public.um_user_contacts WHERE "orgId"=$1 LIMIT $2 OFFSET $3`

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
			txtId, id, "um_user", "GetUserWithContacts", "",
			"view", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer rows.Close()
	var errorMsg string
	var u model.UserContact
	var userList []model.UserContact
	rowIndex := 0
	for rows.Next() {
		rowIndex++
		err := rows.Scan(
			&u.OrgID,
			&u.Username,
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
			response := model.Response{
				Status: "-1",
				Msg:    "Failed",
				Desc:   errorMsg,
			}
			//=======AUDIT_START=====//
			_ = utils.InsertAuditLogs(
				c, conn, orgId.(string), username.(string),
				txtId, id, "um_user", "GetUserWithContacts", "",
				"view", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
			)
			//=======AUDIT_END=====//
			c.JSON(http.StatusInternalServerError, response)
			return
		}

		userList = append(userList, u)
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
			txtId, id, "um_user", "GetUserWithContacts", "",
			"view", -1, now, GetQueryParams(c), response, "Failed : "+errorMsg,
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
			txtId, id, "um_user", "GetUserWithContacts", "",
			"view", 0, now, GetQueryParams(c), response, "GetUserWithContacts Success.",
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusOK, response)
	}
}

// @summary Get User with contacts by id
// @tags User
// @security ApiKeyAuth
// @id Get User with contacts by id
// @Param id path int true "id"
// @accept json
// @produce json
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/users_with_contacts/{id} [get]
func GetUserWithContactsById(c *gin.Context) {
	logger := utils.GetLog()
	id := c.Param("id")
	now := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)

	query := `SELECT  "orgId", username, "contactName", "contactPhone", "contactAddr", "createdAt", "updatedAt", "createdBy", "updatedBy" 
	FROM public.um_user_contacts WHERE id=$1 AND "orgId"=$2`

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
			txtId, id, "um_user", "GetUserWithContactsById", "",
			"view", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer rows.Close()
	var u model.UserContact

	err = rows.Scan(
		&u.OrgID,
		&u.Username,
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
		response := model.Response{
			Status: "-1",
			Msg:    "Failed",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "um_user", "GetUserWithContactsById", "",
			"search", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
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
		txtId, id, "um_user", "GetUserWithContactsById", "",
		"search", 0, now, GetQueryParams(c), response, "GetUserWithContactsById Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)

}

// @summary Create User with contacts
// @id Create User with contacts
// @security ApiKeyAuth
// @tags User
// @accept json
// @produce json
// @param Body body model.UserContactInsert true "Create Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/users_with_contacts/add [post]
func InsertUserWithContacts(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	now := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()
	var req model.UserContactInsert
	if err := c.ShouldBindJSON(&req); err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, "", "um_user", "InsertUserWithContacts", "",
			"create", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusBadRequest, response)
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	var id int
	query := `
	INSERT INTO public."um_user_contacts"(
	"orgId", username, "contactName", "contactPhone", "contactAddr", "createdAt", "updatedAt", "createdBy", "updatedBy")
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING id ;
	`

	err := conn.QueryRow(ctx, query,
		req.OrgID, req.Username, req.ContactName, req.ContactPhone, req.ContactAddr, now,
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
			txtId, "", "um_user", "InsertUserWithContacts", "",
			"create", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
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
		txtId, "", "um_user", "InsertUserWithContacts", "",
		"create", 0, now, GetQueryParams(c), response, "InsertUserWithContacts Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)

}

// @summary Update User with contacts
// @id Update User with contacts
// @security ApiKeyAuth
// @accept json
// @produce json
// @tags User
// @Param id path int true "id"
// @param Body body model.UserContactInsertUpdate true "Update data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/users_with_contacts/{id} [patch]
func UpdateUserWithContacts(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	id := c.Param("id")
	now := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()

	var req model.UserContactInsertUpdate
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
			txtId, id, "um_user", "UpdateUserWithContacts", "",
			"update", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusBadRequest, response)
		return
	}

	query := `UPDATE public."um_user_contacts"
	SET "contactName"=$2, "contactPhone"=$3, "contactAddr"=$4,"updatedAt"=$5,"updatedBy"=$6
	WHERE id = $1 AND "orgId"=$7`
	_, err := conn.Exec(ctx, query, id,
		req.ContactName, req.ContactPhone, req.ContactAddr, now, username, orgId)

	logger.Debug("Update Case SQL Args",
		zap.String("query", query),
		zap.Any("Input", []any{
			id, req.ContactName, req.ContactPhone, req.ContactAddr, now, username, orgId,
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
			txtId, id, "um_user", "UpdateUserWithContacts", "",
			"update", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
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
		txtId, id, "um_user", "UpdateUserWithContacts", "",
		"update", 0, now, GetQueryParams(c), response, "UpdateUserWithContacts Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)
}

// @summary Delete User with contacts
// @id Delete User with contacts
// @security ApiKeyAuth
// @accept json
// @tags User
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/users_with_contacts/{id} [delete]
func DeleteUserWithContacts(c *gin.Context) {

	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	id := c.Param("id")
	now := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()
	query := `DELETE FROM public."um_user_contacts" WHERE id = $1 AND "orgId"=$2`
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
			txtId, id, "um_user", "DeleteUserWithContacts", "",
			"delete", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
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
		txtId, id, "um_user", "DeleteUserWithContacts", "",
		"delete", -1, now, GetQueryParams(c), response, "DeleteUserWithContacts Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Delete successfully",
	})
}

// @summary Get User with socials
// @tags User
// @security ApiKeyAuth
// @id Get User with socials
// @accept json
// @produce json
// @Param start query int false "start" default(0)
// @Param length query int false "length" default(10)
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/users_with_socials [get]
func GetUserWithSocials(c *gin.Context) {
	logger := utils.GetLog()

	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	id := c.Param("id")
	now := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()
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

	query := `SELECT  "orgId", username, "socialType", "socialId", "socialName", "createdAt", "updatedAt", "createdBy", "updatedBy" 	
	FROM public.um_user_with_socials WHERE "orgId"=$1 
	LIMIT $2 OFFSET $3`

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
			txtId, id, "um_user", "GetUserWithSocials", "",
			"view", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer rows.Close()
	var errorMsg string
	var u model.UserSocial
	var userList []model.UserSocial
	rowIndex := 0
	for rows.Next() {
		rowIndex++
		err := rows.Scan(
			&u.OrgID,
			&u.Username,
			&u.SocialType,
			&u.SocialID,
			&u.SocialName,
			&u.CreatedAt,
			&u.UpdatedAt,
			&u.CreatedBy,
			&u.UpdatedBy,
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
				txtId, id, "um_user", "GetUserWithSocials", "",
				"view", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
			)
			//=======AUDIT_END=====//
			c.JSON(http.StatusInternalServerError, response)
			return
		}

		userList = append(userList, u)
	}
	if errorMsg != "" {
		response := model.Response{
			Status: "-1",
			Msg:    "Failed",
			Desc:   errorMsg,
		} //=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "um_user", "GetUserWithSocials", "",
			"view", -1, now, GetQueryParams(c), response, "Failed : "+errorMsg,
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
			txtId, id, "um_user", "GetUserWithSocials", "",
			"view", 0, now, GetQueryParams(c), response, "GetUserWithSocials Success.",
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusOK, response)
	}
}

// @summary Get User with Socials by id
// @tags User
// @security ApiKeyAuth
// @id Get User with Socials by id
// @Param id path int true "id"
// @accept json
// @produce json
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/users_with_socials/{id} [get]
func GetUserWithSocialsById(c *gin.Context) {
	logger := utils.GetLog()
	id := c.Param("id")
	now := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()

	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	query := `SELECT  "orgId", username, "socialType", "socialId", "socialName", "createdAt", "updatedAt", "createdBy", "updatedBy" 	
	FROM public.um_user_with_socials WHERE id=$1 AND "orgId"=$2`

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query))
	rows, err := conn.Query(ctx, query, id, orgId)
	if err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		logger.Warn("Query failed", zap.Error(err))
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "um_user", "GetUserWithSocialsById", "",
			"search", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer rows.Close()
	var u model.UserSocial
	err = rows.Scan(
		&u.OrgID,
		&u.Username,
		&u.SocialType,
		&u.SocialID,
		&u.SocialName,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.CreatedBy,
		&u.UpdatedBy,
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
			txtId, id, "um_user", "GetUserWithSocialsById", "",
			"search", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
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
		txtId, id, "um_user", "GetUserWithSocialsById", "",
		"search", -1, now, GetQueryParams(c), response, "GetUserWithSocialsById",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)

}

// @summary Create User with socials
// @id Create User with socials
// @security ApiKeyAuth
// @tags User
// @accept json
// @produce json
// @param Body body model.UserSocialInsert true "Create Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/users_with_socials/add [post]
func InsertUserWithSocials(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	var req model.UserSocialInsert

	now := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()
	if err := c.ShouldBindJSON(&req); err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, "", "um_user", "InsertUserWithSocials", "",
			"create", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusBadRequest, response)
		logger.Warn("Insert failed", zap.Error(err))
		return
	}
	var id int
	query := `
	INSERT INTO public."um_user_with_socials"(
	"orgId", username, "socialType", "socialId", "socialName", "createdAt", "updatedAt", "createdBy", "updatedBy")
	VALUES ($1, $2, $3, $4, $5, $6, $7,$8,$9)
	RETURNING id ;
	`

	err := conn.QueryRow(ctx, query,
		req.OrgID, req.Username, req.SocialType, req.SocialID, req.SocialName, now,
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
			txtId, "", "um_user", "InsertUserWithSocials", "",
			"create", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
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
		txtId, "", "um_user", "InsertUserWithSocials", "",
		"create", 0, now, GetQueryParams(c), response, "InsertUserWithSocials Success",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)

}

// @summary Update User with socials
// @id Update User with socials
// @security ApiKeyAuth
// @accept json
// @produce json
// @tags User
// @Param id path int true "id"
// @param Body body model.UserSocialUpdate true "Update data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/users_with_socials/{id} [patch]
func UpdateUserWithSocials(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	id := c.Param("id")
	now := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()

	var req model.UserSocialUpdate
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
			txtId, id, "um_user", "UpdateUserWithSocials", "",
			"update", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusBadRequest, response)
		return
	}

	query := `UPDATE public."um_user_with_socials"
	SET "orgId"=$2, username=$3, "socialType"=$4, "socialId"=$5, "socialName"=$6, "updatedAt"=$7, "updatedBy"=$8
	WHERE id = $1 AND "orgId"=$9`
	_, err := conn.Exec(ctx, query,
		id, req.OrgID, req.Username, req.SocialType, req.SocialID, req.SocialName,
		now, username, orgId,
	)
	logger.Debug("Update Case SQL Args",
		zap.String("query", query),
		zap.Any("Input", []any{
			id,
			req.OrgID, req.Username, req.SocialType, req.SocialID, req.SocialName,
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
			txtId, id, "um_user", "UpdateUserWithSocials", "",
			"update", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
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
		txtId, id, "um_user", "UpdateUserWithSocials", "",
		"update", 0, now, GetQueryParams(c), response, "UpdateUserWithSocials Success",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)
}

// @summary Delete User with socials
// @id Delete User with socials
// @security ApiKeyAuth
// @accept json
// @tags User
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/users_with_socials/{id} [delete]
func DeleteUserWithSocials(c *gin.Context) {

	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()
	id := c.Param("id")
	now := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()
	query := `DELETE FROM public."um_user_with_socials" WHERE id = $1 AND "orgId"=$2`
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
			txtId, id, "um_user", "DeleteUserWithSocials", "",
			"delete", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
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
		txtId, id, "um_user", "DeleteUserWithSocials", "",
		"delete", 0, now, GetQueryParams(c), response, "DeleteUserWithSocials Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)
}

// @summary Reset User Password
// @tags User
// @id Reset User Password
// @accept json
// @produce json
// @param Body body model.ResetPasswordRequest true "Reset Password Data (username, email, newPassword)"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/users/reset_password [post]
func ResetUserPassword(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	id := c.Param("id")
	now := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()
	var req model.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "um_user", "ResetUserPassword", "",
			"update", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusBadRequest, response)
		logger.Warn("Reset password failed", zap.Error(err))
		return
	}

	// ตรวจสอบว่า user มีอยู่จริงโดยใช้ username และ email
	var userID string
	checkQuery := `SELECT id FROM public.um_users WHERE username=$1 AND email=$2 AND active=true`
	err := conn.QueryRow(ctx, checkQuery, req.Username, req.Email).Scan(&userID)
	if err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   "User not found or inactive",
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "um_user", "ResetUserPassword", "",
			"update", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusNotFound, response)
		logger.Warn("User not found", zap.Error(err))
		return
	}

	// Encrypt new password (ไม่มีการตรวจสอบเงื่อนไข)
	encPassword, err := encrypt(req.NewPassword)
	if err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   "Password encryption failed",
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "um_user", "ResetUserPassword", "",
			"update", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		logger.Warn("Password encryption failed", zap.Error(err))
		return
	}

	query := `UPDATE public.um_users SET password=$1, "updatedAt"=$2, "updatedBy"=$3 WHERE username=$4 AND email=$5`

	logger.Debug(`Query`, zap.String("query", query))
	logger.Debug(`Username`, zap.String("username", req.Username))
	logger.Debug(`Email`, zap.String("email", req.Email))

	_, err = conn.Exec(ctx, query, encPassword, now, "system", req.Username, req.Email)
	if err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "um_user", "ResetUserPassword", "",
			"update", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		logger.Warn("Reset password failed", zap.Error(err))
		return
	}
	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Password reset successfully",
	}
	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, id, "um_user", "ResetUserPassword", "",
		"update", 0, now, GetQueryParams(c), response, "ResetUserPassword Success.",
	)
	//=======AUDIT_END=====//

	c.JSON(http.StatusOK, model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Password reset successfully",
	})
}

// @summary Change User Password
// @tags User
// @security ApiKeyAuth
// @id Change User Password
// @accept json
// @produce json
// @param id path string true "User ID"
// @param Body body model.ChangePasswordRequest true "Change Password Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/users/change_password/{id} [patch]
func ChangeUserPassword(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)

	var req model.ChangePasswordRequest
	id := c.Param("id")
	now := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()
	if err := c.ShouldBindJSON(&req); err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "um_user", "changeUserPassword", "",
			"update", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusBadRequest, response)
		logger.Warn("Change password failed", zap.Error(err))
		return
	}

	// First, verify current password
	var currentPassword string
	query := `SELECT password FROM public.um_users WHERE id=$1 AND "orgId"=$2`
	err := conn.QueryRow(ctx, query, id, orgId).Scan(&currentPassword)
	if err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   "User not found",
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "um_user", "changeUserPassword", "",
			"update", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusNotFound, response)
		logger.Warn("User not found", zap.Error(err))
		return
	}

	// Decrypt current password and verify
	decryptedPassword, err := decrypt(currentPassword)
	if err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   "Password verification failed",
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "um_user", "changeUserPassword", "",
			"update", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   "Password verification failed",
		})
		logger.Warn("Password decryption failed", zap.Error(err))
		return
	}

	// Compare current password
	if subtle.ConstantTimeCompare([]byte(decryptedPassword), []byte(req.CurrentPassword)) != 1 {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   "Current password is incorrect",
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "um_user", "changeUserPassword", "",
			"update", -1, now, GetQueryParams(c), response, "Current password is incorrect.",
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	// Encrypt new password
	encPassword, err := encrypt(req.NewPassword)
	if err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   "Password encryption failed",
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "um_user", "changeUserPassword", "",
			"update", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		logger.Warn("Password encryption failed", zap.Error(err))
		return
	}

	updateQuery := `UPDATE public.um_users SET password=$1, "updatedAt"=$2, "updatedBy"=$3 WHERE id=$4 AND "orgId"=$5`

	logger.Debug(`Query`, zap.String("query", updateQuery))
	logger.Debug(`User ID`, zap.String("id", id))

	_, err = conn.Exec(ctx, updateQuery, encPassword, now, username, id, orgId)
	if err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "um_user", "changeUserPassword", "",
			"update", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		logger.Warn("Change password failed", zap.Error(err))
		return
	}

	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Password changed successfully",
	}
	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, id, "um_user", "changeUserPassword", "",
		"update", 0, now, GetQueryParams(c), response, "ChangeUserPassword Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)
}

// @summary Get User Groups
// @tags UserGroup
// @security ApiKeyAuth
// @id GetUserGroups
// @accept json
// @produce json
// @Param start query int false "start" default(0)
// @Param length query int false "length" default(10)
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/user_groups/all [get]
func GetUmGroupList(c *gin.Context) {
	logger := utils.GetLog()
	id := c.Param("id")
	now := time.Now()
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

	query := `
	SELECT "id", "orgId", "grpId", "en", "th", "active", "createdAt", "updatedAt", "createdBy", "updatedBy"
	FROM public.um_groups
	WHERE "orgId"=$1
	LIMIT $2 OFFSET $3
	`
	logger.Debug(`Query`, zap.String("query", query))

	rows, err := conn.Query(ctx, query, orgId, length, start)
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
			txtId, id, "um_user", "changeUserPassword", "",
			"view", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer rows.Close()

	var group model.UmGroup
	var groupList []model.UmGroup
	var errorMsg string

	for rows.Next() {
		err := rows.Scan(
			&group.ID,
			&group.OrgID,
			&group.GrpID,
			&group.En,
			&group.Th,
			&group.Active,
			&group.CreatedAt,
			&group.UpdatedAt,
			&group.CreatedBy,
			&group.UpdatedBy,
		)
		if err != nil {
			logger.Warn("Scan failed", zap.Error(err))
			errorMsg = err.Error()
			break
		}
		groupList = append(groupList, group)
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
			txtId, id, "um_user", "changeUserPassword", "",
			"view", -1, now, GetQueryParams(c), response, "Failed : "+errorMsg,
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
	} else {
		response := model.Response{
			Status: "0",
			Msg:    "Success",
			Data:   groupList,
			Desc:   "",
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "um_user", "changeUserPassword", "",
			"view", -1, now, GetQueryParams(c), response, "GetUmGroupList",
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusOK, response)
	}
}
