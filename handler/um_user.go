package handler

import (
	"mainPackage/config"
	"mainPackage/model"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// Stations godoc
// @summary Get User
// @tags User
// @security ApiKeyAuth
// @id Get User
// @accept json
// @produce json
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/users [get]
func GetUmUserList(c *gin.Context) {
	logger := config.GetLog()
	orgId := GetVariableFromToken(c, "orgId")
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	query := `SELECT "orgId", "displayName", title, "firstName", "middleName", "lastName", "citizenId", bod, blood, gender, "mobileNo", address, photo, username, password, email, "roleId", "userType", "empId", "deptId", "commId", "stnId", active, "activationToken", "lastActivationRequest", "lostPasswordRequest", "signupStamp", islogin, "lastLogin", "createdAt", "updatedAt", "createdBy", "updatedBy" 
	FROM public.um_users WHERE "orgId"=$1 LIMIT 1000`

	var rows pgx.Rows
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
	var errorMsg string
	var u model.Um_User
	var userList []model.Um_User
	rowIndex := 0
	for rows.Next() {
		rowIndex++
		err := rows.Scan(
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
			response := model.Response{
				Status: "-1",
				Msg:    "Failed",
				Desc:   errorMsg,
			}
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
		c.JSON(http.StatusInternalServerError, response)
	} else {
		response := model.Response{
			Status: "0",
			Msg:    "Success",
			Data:   userList,
			Desc:   "",
		}
		c.JSON(http.StatusOK, response)
	}
}

// Stations godoc
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
	logger := config.GetLog()
	username := c.Param("username")
	orgId := GetVariableFromToken(c, "orgId")
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	query := `SELECT "orgId", "displayName", title, "firstName", "middleName", "lastName", "citizenId", bod, blood, gender, "mobileNo", address, photo, username, password, email, "roleId", "userType", "empId", "deptId", "commId", "stnId", active, "activationToken", "lastActivationRequest", "lostPasswordRequest", "signupStamp", islogin, "lastLogin", "createdAt", "updatedAt", "createdBy", "updatedBy" 
	FROM public.um_users WHERE username=$1 AND "orgId"=$2`

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query),
		zap.Any("Input", []any{
			username, orgId,
		}))
	rows, err := conn.Query(ctx, query, username, orgId)
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
	var u model.Um_User

	err = rows.Scan(
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
		errorMsg = err.Error()
		logger.Warn("Scan failed", zap.Error(err))
		response := model.Response{
			Status: "-1",
			Msg:    "Failed",
			Desc:   errorMsg,
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	if errorMsg != "" {
		response := model.Response{
			Status: "-1",
			Msg:    "Failed",
			Desc:   errorMsg,
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	} else {
		response := model.Response{
			Status: "0",
			Msg:    "Success",
			Data:   u,
			Desc:   "",
		}
		c.JSON(http.StatusOK, response)
		return
	}
}

// Stations godoc
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
	logger := config.GetLog()
	id := c.Param("id")
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	orgId := GetVariableFromToken(c, "orgId")
	defer cancel()
	defer conn.Close(ctx)
	query := `SELECT "orgId", "displayName", title, "firstName", "middleName", "lastName", "citizenId", bod, blood, gender, "mobileNo", address, photo, username, password, email, "roleId", "userType", "empId", "deptId", "commId", "stnId", active, "activationToken", "lastActivationRequest", "lostPasswordRequest", "signupStamp", islogin, "lastLogin", "createdAt", "updatedAt", "createdBy", "updatedBy" 
	FROM public.um_users WHERE id=$1 AND "orgId"=$2`

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query))
	rows, err := conn.Query(ctx, query, id, orgId)
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
	var u model.Um_User
	err = rows.Scan(
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
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	if errorMsg != "" {
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
			Data:   u,
			Desc:   "",
		}
		c.JSON(http.StatusOK, response)
	}
}

// Stations godoc
// @summary Get User with skills
// @tags User
// @security ApiKeyAuth
// @id Get User with skills
// @accept json
// @produce json
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/users_with_skills [get]
func GetUserWithSkills(c *gin.Context) {
	logger := config.GetLog()
	orgId := GetVariableFromToken(c, "orgId")
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	query := `SELECT "orgId", "userName", "skillId", active, "createdAt", "updatedAt", "createdBy", "updatedBy" 
	FROM public.um_user_with_skills WHERE "orgId"=$1`

	var rows pgx.Rows
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
		c.JSON(http.StatusInternalServerError, response)
	} else {
		response := model.Response{
			Status: "0",
			Msg:    "Success",
			Data:   userList,
			Desc:   "",
		}
		c.JSON(http.StatusOK, response)
	}
}

// Stations godoc
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
	logger := config.GetLog()
	id := c.Param("id")
	orgId := GetVariableFromToken(c, "orgId")
	conn, ctx, cancel := config.ConnectDB()
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
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
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
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Data:   u,
		Desc:   "",
	}
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
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()
	username := GetVariableFromToken(c, "username")
	var req model.UserSkillInsert
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	// now req is ready to use
	now := time.Now()
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
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	id := c.Param("id")

	var req model.UserSkillUpdate
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
	username := GetVariableFromToken(c, "username")
	now := time.Now()
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

	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	id := c.Param("id")
	query := `DELETE FROM public."um_user_with_skills" WHERE id = $1 `
	logger.Debug("Query", zap.String("query", query), zap.Any("id", id))
	_, err := conn.Exec(ctx, query, id)
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

// Stations godoc
// @summary Get User with contacts
// @tags User
// @security ApiKeyAuth
// @id Get User with contacts
// @accept json
// @produce json
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/users_with_contacts [get]
func GetUserWithContacts(c *gin.Context) {
	logger := config.GetLog()

	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	orgId := GetVariableFromToken(c, "orgId")
	defer cancel()
	defer conn.Close(ctx)
	query := `SELECT  "orgId", username, "contactName", "contactPhone", "contactAddr", "createdAt", "updatedAt", "createdBy", "updatedBy" 	
	FROM public.um_user_contacts WHERE "orgId"=$1 LIMIT 1000`

	var rows pgx.Rows
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
		c.JSON(http.StatusInternalServerError, response)
	} else {
		response := model.Response{
			Status: "0",
			Msg:    "Success",
			Data:   userList,
			Desc:   "",
		}
		c.JSON(http.StatusOK, response)
	}
}

// Stations godoc
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
	logger := config.GetLog()
	id := c.Param("id")
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	orgId := GetVariableFromToken(c, "orgId")
	query := `SELECT  "orgId", username, "contactName", "contactPhone", "contactAddr", "createdAt", "updatedAt", "createdBy", "updatedBy" 
	FROM public.um_user_contacts WHERE id=$1 AND "orgId"=$2`

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query))
	rows, err := conn.Query(ctx, query, id, orgId)
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
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Data:   u,
		Desc:   "",
	}
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
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	var req model.UserContactInsert
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	now := time.Now()
	username := GetVariableFromToken(c, "username")
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
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	id := c.Param("id")

	var req model.UserContactInsertUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Update failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		return
	}
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	now := time.Now()
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

	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	id := c.Param("id")
	orgId := GetVariableFromToken(c, "orgId")
	query := `DELETE FROM public."um_user_contacts" WHERE id = $1 AND "orgId"=$2`
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

// Stations godoc
// @summary Get User with socials
// @tags User
// @security ApiKeyAuth
// @id Get User with socials
// @accept json
// @produce json
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/users_with_socials [get]
func GetUserWithSocials(c *gin.Context) {
	logger := config.GetLog()

	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	orgId := GetVariableFromToken(c, "orgId")
	query := `SELECT  "orgId", username, "socialType", "socialId", "socialName", "createdAt", "updatedAt", "createdBy", "updatedBy" 	
	FROM public.um_user_with_socials AND "orgId"=$1 LIMIT 1000`

	var rows pgx.Rows
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
		c.JSON(http.StatusInternalServerError, response)
	} else {
		response := model.Response{
			Status: "0",
			Msg:    "Success",
			Data:   userList,
			Desc:   "",
		}
		c.JSON(http.StatusOK, response)
	}
}

// Stations godoc
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
	logger := config.GetLog()
	id := c.Param("id")
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	orgId := GetVariableFromToken(c, "orgId")
	query := `SELECT  "orgId", username, "socialType", "socialId", "socialName", "createdAt", "updatedAt", "createdBy", "updatedBy" 	
	FROM public.um_user_with_socials WHERE id=$1 AND "orgId"=$2`

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query))
	rows, err := conn.Query(ctx, query, id, orgId)
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
		c.JSON(http.StatusInternalServerError, response)
	}

	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Data:   u,
		Desc:   "",
	}
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
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	var req model.UserSocialInsert
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
	// now req is ready to use
	now := time.Now()
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
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	id := c.Param("id")

	var req model.UserSocialUpdate
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
	query := `DELETE FROM public."um_user_with_socials" WHERE id = $1 AND "orgId"=$2`
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
