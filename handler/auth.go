package handler

import (
	"crypto/subtle"
	"fmt"
	"mainPackage/model"
	"mainPackage/utils"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func GetVariableFromToken(c *gin.Context, varname string) interface{} {
	var_return, exists := c.Get(varname)
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return nil
	}
	return var_return
}

func CreateToken(username string, orgId string) (string, string, error) {

	var secretKey = []byte(os.Getenv("TOKEN_SECRET_KEY"))
	var refreshKey = []byte(os.Getenv("REFRESH_TOKEN_KEY"))
	TimeoutStr := os.Getenv("TOKEN_TIMEOUT")
	timeoutInt, _ := strconv.Atoi(TimeoutStr)
	var TIMEOUT = time.Minute * time.Duration(timeoutInt)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": username,
			"orgId":    orgId,
			"exp":      time.Now().Add(TIMEOUT).Unix(),
		})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", "", err
	}
	TimeoutStr = os.Getenv("REFRESH_TOKEN_TIMEOUT")
	timeoutInt, _ = strconv.Atoi(TimeoutStr)
	TIMEOUT = time.Minute * time.Duration(timeoutInt)
	refreshtoken := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": username,
			"orgId":    orgId,
			"exp":      time.Now().Add(TIMEOUT).Unix(),
		})

	refreshtokenString, err := refreshtoken.SignedString(refreshKey)
	if err != nil {
		return "", "", err
	}

	return tokenString, refreshtokenString, nil
}

func verifyToken(tokenString string) (*jwt.Token, error) {
	var secretKey = []byte(os.Getenv("TOKEN_SECRET_KEY"))
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return token, nil
}

func verifyRefreshToken(tokenString string) (*jwt.Token, error) {
	var secretKey = []byte(os.Getenv("REFRESH_TOKEN_KEY"))
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return token, nil
}

func ProtectedHandler(c *gin.Context) {
	logger := utils.GetLog()
	authHeader := c.GetHeader("Authorization")
	logger.Debug("Authorization header", zap.String("Authorization", authHeader))

	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, model.Response{
			Status: "-1",
			Msg:    "Failed",
			Desc:   "Missing authorization header",
		})
		c.Abort()
		return
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		c.JSON(http.StatusUnauthorized, model.Response{
			Status: "-1",
			Msg:    "Failed",
			Desc:   "Invalid authorization header format",
		})
		c.Abort()
		return
	}

	tokenString := strings.TrimPrefix(authHeader, prefix)
	logger.Debug("Token: " + tokenString)

	parsedToken, err := verifyToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.Response{
			Status: "-1",
			Msg:    "Failed",
			Desc:   err.Error(),
		})
		c.Abort()
		return
	}

	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
		username, uOK := claims["username"].(string)
		orgId, orgOK := claims["orgId"].(string)

		if uOK && orgOK {
			logger.Debug("Verified user",
				zap.String("username", username),
				zap.String("orgId", orgId),
			)
			c.Set("username", username)
			c.Set("orgId", orgId)
			c.Set("tokenString", tokenString)
			c.Next()
			return
		}
	}

}

// @summary Login
// @tags Authentication
// @security ApiKeyAuth
// @id Login User
// @accept json
// @produce json
// @Param username query string true "username"
// @Param password query string true "password"
// @Param organization query string true "organization"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/auth/login [get]
func UserLogin(c *gin.Context) {
	logger := utils.GetLog()
	username := c.Query("username")
	password := c.Query("password")
	organization := c.Query("organization")
	txtId := uuid.New().String()
	start_time := time.Now()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	var id string
	query := `SELECT id FROM public.organizations WHERE name = $1`
	logger.Debug(`Query`, zap.String("query", query))
	logger.Debug(`request input`, zap.Any("organization", organization))
	err := conn.QueryRow(ctx, query, organization).Scan(&id)
	if err != nil {
		logger.Debug(err.Error())
		c.JSON(http.StatusUnauthorized, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		return
	}

	query = `SELECT 
    u.id,
    u."orgId",
    u."displayName",
    u.title,
    u."firstName",
    u."middleName",
    u."lastName",
    u."citizenId",
    u.bod,
    u.blood,
    u.gender,
    u."mobileNo",
    u.address,
    u.photo,
    u.username,
    u.password,
    u.email,
    u."roleId",
    u."userType",
    u."empId",
    u."deptId",
    u."commId",
    u."stnId",
    u.active,
    u."activationToken",
    u."lastActivationRequest",
    u."lostPasswordRequest",
    u."signupStamp",
    u.islogin,
    u."lastLogin",
    u."createdAt",
    u."updatedAt",
    u."createdBy",
    u."updatedBy",
    COALESCE(a."distIdLists", '[]'::jsonb) AS "distIdLists"
FROM public.um_users u
LEFT JOIN public.um_user_with_area_response a
    ON a."username" = u.username AND a."orgId" = u."orgId"
WHERE u.username = $1
  AND u.active = TRUE;
`
	logger.Debug(`Query`, zap.String("query", query))
	logger.Debug(`request input`, zap.Any("username", username))
	var UserOpt model.Um_User_Login
	err = conn.QueryRow(ctx, query, username).Scan(&UserOpt.ID,
		&UserOpt.OrgID, &UserOpt.DisplayName, &UserOpt.Title, &UserOpt.FirstName, &UserOpt.MiddleName, &UserOpt.LastName,
		&UserOpt.CitizenID, &UserOpt.Bod, &UserOpt.Blood, &UserOpt.Gender, &UserOpt.MobileNo, &UserOpt.Address,
		&UserOpt.Photo, &UserOpt.Username, &UserOpt.Password, &UserOpt.Email, &UserOpt.RoleID, &UserOpt.UserType,
		&UserOpt.EmpID, &UserOpt.DeptID, &UserOpt.CommID, &UserOpt.StnID, &UserOpt.Active, &UserOpt.ActivationToken,
		&UserOpt.LastActivationRequest, &UserOpt.LostPasswordRequest, &UserOpt.SignupStamp, &UserOpt.IsLogin, &UserOpt.LastLogin,
		&UserOpt.CreatedAt, &UserOpt.UpdatedAt, &UserOpt.CreatedBy, &UserOpt.UpdatedBy, &UserOpt.DistIdLists)
	if err != nil {
		logger.Debug(err.Error())
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, UserOpt.OrgID, username,
			txtId, "", "Auth", "UserLogin", "",
			"search", -1, start_time, GetQueryParams(c), response, "Failure = "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusUnauthorized, response)
		return
	}
	var dec string
	dec, err = decrypt(UserOpt.Password)
	if err != nil {
		logger.Debug(err.Error())
		return
	}

	if subtle.ConstantTimeCompare([]byte(dec), []byte(password)) == 1 {
		tokenString, refreshtoken, err := CreateToken(username, id)
		if err != nil {
			response := gin.H{
				"error":   "Token creation failed",
				"message": err.Error(),
			}
			//=======AUDIT_START=====//
			_ = utils.InsertAuditLogs(
				c, conn, UserOpt.OrgID, username,
				txtId, "", "Auth", "UserLogin", "",
				"search", -1, start_time, GetQueryParams(c), response, "Token creation failed = "+err.Error(),
			)
			//=======AUDIT_END=====//
			c.JSON(http.StatusUnauthorized, response)
			logger.Debug(err.Error())
			return
		}
		response := model.Response{
			Status: "0",
			Msg:    "Success",
			Desc:   "",
			Data: gin.H{
				"accessToken":  tokenString,
				"refreshToken": refreshtoken,
				"token_type":   "bearer",
				"user":         UserOpt,
			},
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, UserOpt.OrgID, username,
			txtId, "", "Auth", "UserLogin", "",
			"search", 0, start_time, GetQueryParams(c), response, "Login successfully",
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusOK, response)
	} else {
		response := model.Response{
			Status: "-1",
			Msg:    "",
			Desc:   "Invalid credentials",
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, UserOpt.OrgID, username,
			txtId, "", "Auth", "UserLogin", "",
			"search", -1, start_time, GetQueryParams(c), response, "Invalid credentials",
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusUnauthorized, response)
	}
}

// @summary Login User Post
// @tags Authentication
// @security ApiKeyAuth
// @id Login User Post
// @accept json
// @produce json
// @param Body body model.Login true "Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/auth/login [post]
func UserLoginPost(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	start_time := time.Now()
	txtId := uuid.New().String()
	var req model.Login
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Update failed", zap.Error(err))
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, "", "",
			txtId, "", "Auth", "UserLoginPost", "",
			"search", -1, start_time, GetQueryParams(c), response, "Failure : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusBadRequest, response)
		return
	}

	organization := req.Organization
	username := req.Username
	password := req.Password
	var id string
	query := `SELECT id FROM public.organizations WHERE name = $1`
	logger.Debug(`Query`, zap.String("query", query))
	logger.Debug(`request input`, zap.Any("organization", organization))
	err := conn.QueryRow(ctx, query, organization).Scan(&id)
	if err != nil {
		logger.Debug(err.Error())
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, "", username,
			txtId, "", "Auth", "UserLoginPost", "",
			"search", -1, start_time, GetQueryParams(c), response, "Failure : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	query = `SELECT 
    u.id,
    u."orgId",
    u."displayName",
    u.title,
    u."firstName",
    u."middleName",
    u."lastName",
    u."citizenId",
    u.bod,
    u.blood,
    u.gender,
    u."mobileNo",
    u.address,
    u.photo,
    u.username,
    u.password,
    u.email,
    u."roleId",
    u."userType",
    u."empId",
    u."deptId",
    u."commId",
    u."stnId",
    u.active,
    u."activationToken",
    u."lastActivationRequest",
    u."lostPasswordRequest",
    u."signupStamp",
    u.islogin,
    u."lastLogin",
    u."createdAt",
    u."updatedAt",
    u."createdBy",
    u."updatedBy",
    COALESCE(a."distIdLists", '[]'::jsonb) AS "distIdLists"
FROM public.um_users u
LEFT JOIN public.um_user_with_area_response a
    ON a."username" = u.username AND a."orgId" = u."orgId"
WHERE u.username = $1
  AND u.active = TRUE;
`
	logger.Debug(`Query`, zap.String("query", query))
	logger.Debug(`request input`, zap.Any("username", username))
	var UserOpt model.Um_User_Login
	err = conn.QueryRow(ctx, query, username).Scan(&UserOpt.ID,
		&UserOpt.OrgID, &UserOpt.DisplayName, &UserOpt.Title, &UserOpt.FirstName, &UserOpt.MiddleName, &UserOpt.LastName,
		&UserOpt.CitizenID, &UserOpt.Bod, &UserOpt.Blood, &UserOpt.Gender, &UserOpt.MobileNo, &UserOpt.Address,
		&UserOpt.Photo, &UserOpt.Username, &UserOpt.Password, &UserOpt.Email, &UserOpt.RoleID, &UserOpt.UserType,
		&UserOpt.EmpID, &UserOpt.DeptID, &UserOpt.CommID, &UserOpt.StnID, &UserOpt.Active, &UserOpt.ActivationToken,
		&UserOpt.LastActivationRequest, &UserOpt.LostPasswordRequest, &UserOpt.SignupStamp, &UserOpt.IsLogin, &UserOpt.LastLogin,
		&UserOpt.CreatedAt, &UserOpt.UpdatedAt, &UserOpt.CreatedBy, &UserOpt.UpdatedBy, &UserOpt.DistIdLists)
	if err != nil {
		logger.Debug(err.Error())
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, "", username,
			txtId, "", "Auth", "UserLoginPost", "",
			"search", -1, start_time, GetQueryParams(c), response, "Failure : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusUnauthorized, response)
		return
	}
	var dec string
	dec, err = decrypt(UserOpt.Password)
	if err != nil {
		response := zap.Error(err)
		logger.Warn("Decryption failed", response) // Use Warn for visibility
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, UserOpt.OrgID, username,
			txtId, "", "Auth", "UserLoginPost", "",
			"search", -1, start_time, GetQueryParams(c), response, "Decryption failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		return
	}
	// logger.Debug(UserOpt.Password)
	// logger.Debug(dec)
	if subtle.ConstantTimeCompare([]byte(dec), []byte(password)) == 1 {
		tokenString, refreshtoken, err := CreateToken(username, id)
		if err != nil {
			response := gin.H{
				"error":   "Token creation failed",
				"message": err.Error(),
			}
			//=======AUDIT_START=====//
			_ = utils.InsertAuditLogs(
				c, conn, UserOpt.OrgID, username,
				txtId, "", "Auth", "UserLoginPost", "",
				"search", -1, start_time, GetQueryParams(c), response, "Token creation failed : "+err.Error(),
			)
			//=======AUDIT_END=====//
			c.JSON(http.StatusUnauthorized, response)
			logger.Debug(err.Error())
			return
		}

		query = `SELECT "permId" FROM public.um_role_with_permissions 
				WHERE "orgId"=$1 AND "roleId"=$2 AND active = true`
		logger.Debug(`Query`, zap.String("query", query), zap.Any("input", []any{UserOpt.OrgID, UserOpt.RoleID}))
		var permId string
		var RolePermissionList []string
		rows, err := conn.Query(ctx, query, UserOpt.OrgID, UserOpt.RoleID)
		if err != nil {
			logger.Warn("Query failed", zap.Error(err))
			response := model.Response{
				Status: "-1",
				Msg:    "Failure",
				Desc:   err.Error(),
			}
			//=======AUDIT_START=====//
			_ = utils.InsertAuditLogs(
				c, conn, UserOpt.OrgID, username,
				txtId, "", "Auth", "UserLoginPost", "",
				"search", -1, start_time, GetQueryParams(c), response, "Failure : "+err.Error(),
			)
			//=======AUDIT_END=====//
			c.JSON(http.StatusInternalServerError, response)
			return
		}
		for rows.Next() {
			err = rows.Scan(&permId)
			if err != nil {
				logger.Warn("Scan failed", zap.Error(err))
				response := model.Response{
					Status: "-1",
					Msg:    "Failed",
					Desc:   err.Error(),
				}
				//=======AUDIT_START=====//
				_ = utils.InsertAuditLogs(
					c, conn, UserOpt.OrgID, username,
					txtId, "", "Auth", "UserLoginPost", "",
					"search", -1, start_time, GetQueryParams(c), response, "Scan failed : "+err.Error(),
				)
				//=======AUDIT_END=====//
				c.JSON(http.StatusInternalServerError, response)
			}
			RolePermissionList = append(RolePermissionList, permId)

		}
		UserOpt.Permission = RolePermissionList
		response := model.Response{
			Status: "0",
			Msg:    "Success",
			Desc:   "",
			Data: gin.H{
				"accessToken":  tokenString,
				"refreshToken": refreshtoken,
				"token_type":   "bearer",
				"user":         UserOpt,
				// "permission":   RolePermissionList,
			},
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, "", username,
			txtId, UserOpt.OrgID, "Auth", "UserLoginPost", "",
			"search", 0, start_time, GetQueryParams(c), response, "Successfully",
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusOK, response)
		return
	} else {
		response := model.Response{
			Status: "-1",
			Msg:    "",
			Desc:   "Invalid credentials",
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, UserOpt.OrgID, username,
			txtId, "", "Auth", "UserLoginPost", "",
			"search", -1, start_time, GetQueryParams(c), response, "Invalid credentials",
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusUnauthorized, response)
		return
	}

}

// @summary Create User Auth
// @tags Authentication
// @security ApiKeyAuth
// @id Create User Auth
// @accept json
// @produce json
// @param Case body model.UserAdminInput true "User to be created"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/auth/add [post]
func UserAddAuth(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()

	var req model.UserAdminInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, "", "Auth", "UserAddAuth", "",
			"search", -1, start_time, GetQueryParams(c), response, "Failure : "+err.Error(),
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
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, "", "Auth", "UserAddAuth", "",
			"search", -1, start_time, GetQueryParams(c), "", "encrypt fail : "+err.Error(),
		)
		//=======AUDIT_END=====//
		return
	}
	now := time.Now()
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
		req.OrgID, req.DisplayName, req.Title, req.FirstName, req.MiddleName,
		req.LastName, req.CitizenID, req.Bod, req.Blood,
		req.Gender, req.MobileNo, req.Address, req.Photo, req.Username,
		enc, req.Email, req.RoleID, req.UserType, req.EmpID, req.DeptID, req.CommID, req.StnID,
		req.Active, req.ActivationToken, req.LastActivationRequest, req.LostPasswordRequest, req.SignupStamp,
		req.IsLogin, req.LastLogin, now, now, "system", "system",
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
			txtId, "", "Auth", "UserAddAuth", "",
			"create", -1, start_time, GetQueryParams(c), response, "Failure : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusUnauthorized, response)
		logger.Warn("Insert failed", zap.Error(err))
		return
	}
	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Create successfully",
	}
	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, "", "Auth", "UserAddAuth", "",
		"create", 0, start_time, GetQueryParams(c), response, "Create User success.",
	)
	//=======AUDIT_END=====//
	// Continue logic...
	c.JSON(http.StatusOK, response)
}

// @summary Refresh Token
// @tags Authentication
// @security ApiKeyAuth
// @id Refresh Token
// @accept json
// @produce json
// @param Case body model.RefreshInput true "Body"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/auth/refresh [post]
func RefreshToken(c *gin.Context) {
	logger := utils.GetLog()
	var req model.RefreshInput
	id := c.Param("id")
	start_time := time.Now()
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
			c, nil, orgId.(string), username.(string),
			txtId, id, "Auth", "RefreshToken", "",
			"update", -1, start_time, GetQueryParams(c), response, "Failure : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusBadRequest, response)
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	parsedToken, err := verifyRefreshToken(req.RefreshToken)
	if err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failed",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, nil, orgId.(string), username.(string),
			txtId, id, "Auth", "RefreshToken", "",
			"update", -1, start_time, GetQueryParams(c), response, "Failure : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusUnauthorized, response)
		c.Abort()
		return
	}

	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
		username, uOK := claims["username"].(string)
		orgId, orgOK := claims["orgId"].(string)

		if uOK && orgOK {
			tokenString, _, err := CreateToken(username, orgId)
			if err != nil {
				response := gin.H{
					"error":   "Token creation failed",
					"message": err.Error(),
				}
				//=======AUDIT_START=====//
				_ = utils.InsertAuditLogs(
					c, nil, orgId, username,
					txtId, id, "Auth", "RefreshToken", "",
					"update", -1, start_time, GetQueryParams(c), response, "Token creation failed : "+err.Error(),
				)
				//=======AUDIT_END=====//
				c.JSON(http.StatusUnauthorized, response)
				logger.Debug(err.Error())
				return
			}
			response := model.Response{
				Status: "0",
				Msg:    "Success",
				Desc:   "",
				Data: gin.H{
					"accessToken": tokenString,
					// "refreshToken": refreshtoken,
					"token_type": "bearer",
				},
			}
			//=======AUDIT_START=====//
			_ = utils.InsertAuditLogs(
				c, nil, orgId, username,
				txtId, id, "Auth", "RefreshToken", "",
				"update", 0, start_time, GetQueryParams(c), response, "RefreshToken Success",
			)
			//=======AUDIT_END=====//
			c.JSON(http.StatusOK, response)

		}
	}
}
