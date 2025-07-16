package handler

import (
	"mainPackage/config"
	"mainPackage/model"
	"net/http"

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

	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	query := `SELECT "orgId", "displayName", title, "firstName", "middleName", "lastName", "citizenId", bod, blood, gender, "mobileNo", address, photo, username, password, email, "roleId", "userType", "empId", "deptId", "commId", "stnId", active, "activationToken", "lastActivationRequest", "lostPasswordRequest", "signupStamp", islogin, "lastLogin", "createdAt", "updatedAt", "createdBy", "updatedBy" FROM public.um_users`

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query))
	rows, err := conn.Query(ctx, query)
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
	defer cancel()
	defer conn.Close(ctx)
	query := `SELECT "orgId", "displayName", title, "firstName", "middleName", "lastName", "citizenId", bod, blood, gender, "mobileNo", address, photo, username, password, email, "roleId", "userType", "empId", "deptId", "commId", "stnId", active, "activationToken", "lastActivationRequest", "lostPasswordRequest", "signupStamp", islogin, "lastLogin", "createdAt", "updatedAt", "createdBy", "updatedBy" 
	FROM public.um_users WHERE id=$1`

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query))
	rows, err := conn.Query(ctx, query, id)
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

	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	query := `SELECT "orgId", "userName", "skillId", active, "createdAt", "updatedAt", "createdBy", "updatedBy" FROM public.um_user_with_skills`

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query))
	rows, err := conn.Query(ctx, query)
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
