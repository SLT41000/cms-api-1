package handler

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"log"
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
	"github.com/jackc/pgx/v5"
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
	secretKey := []byte(os.Getenv("REFRESH_TOKEN_KEY"))

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
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
func UserLoginGet(c *gin.Context) {
	// Parse from Query
	req := model.Login{
		Username:     c.Query("username"),
		Password:     c.Query("password"),
		Organization: c.Query("organization"),
	}

	// DB Connect
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1", Msg: "Failure", Desc: "DB connection error",
		})
		return
	}
	defer cancel()
	defer conn.Close(ctx)

	// Call Shared Logic
	resp, err := LoginUser(ctx, c, conn, req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	c.JSON(http.StatusOK, resp)
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
	var req model.Login
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Status: "-1", Msg: err.Error()})
		return
	}

	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		c.JSON(http.StatusInternalServerError, model.Response{Status: "-1", Msg: "DB error"})
		return
	}
	defer cancel()
	defer conn.Close(ctx)

	resp, err := LoginUser(ctx, c, conn, req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	c.JSON(http.StatusOK, resp)
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
			txtId, "", "Authentication", "UserAddAuth", "",
			"create", -1, start_time, GetQueryParams(c), response, "Failure : "+err.Error(),
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
			txtId, "", "Authentication", "UserAddAuth", "",
			"login", -1, start_time, GetQueryParams(c), "", "encrypt fail : "+err.Error(),
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
			txtId, "", "Authentication", "UserAddAuth", "",
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
		txtId, "", "Authentication", "UserAddAuth", "",
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
	var input model.RefreshInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// verify refresh token
	token, err := verifyRefreshToken(input.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	claims := token.Claims.(jwt.MapClaims)
	username := claims["username"].(string)
	orgId := claims["orgId"].(string)

	accessToken, refreshToken, err := CreateToken(username, orgId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	resp := model.Response{
		Status: "0",
		Msg:    "Success",
		Data: gin.H{
			"accessToken":  accessToken,
			"refreshToken": refreshToken,
			"token_type":   "bearer",
		},
	}

	c.JSON(http.StatusOK, resp)
}

// @summary Logout
// @tags Authentication
// @security ApiKeyAuth
// @id Logout
// @accept json
// @produce json
// @response 200 {object} model.Response "OK - Logout successful"
// @Router /api/v1/logout [post]
func UserLogout(c *gin.Context) {

	logger := utils.GetLog()
	startTime := time.Now()
	txtId := uuid.New().String()

	orgId := GetVariableFromToken(c, "orgId")
	username := GetVariableFromToken(c, "username")

	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1", Msg: "Failure", Desc: "Cannot connect to database",
		})
		return
	}
	defer cancel()
	defer conn.Close(ctx)

	query := `
		UPDATE public.um_users
		SET "islogin" = FALSE, "updatedAt" = NOW(), "updatedBy" = $3
		WHERE "orgId" = $1 AND "username" = $2
	`
	log.Print(username)
	cmdTag, err := conn.Exec(ctx, query, orgId, username, username)
	if err != nil {
		logger.Warn("Logout update failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1", Msg: "Failure", Desc: "Logout failed: " + err.Error(),
		})
		return
	}

	if cmdTag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, model.Response{
			Status: "-1", Msg: "Failure", Desc: "User not found or already logged out",
		})
		return
	}

	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Logout successful",
	}

	// Optionally record audit log
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, "", "Authentication", "Logout", "",
		"logout", 0, startTime, nil, response, "User logged out successfully",
	)

	c.JSON(http.StatusOK, response)
}

// @summary Verify Token
// @tags Authentication
// @security ApiKeyAuth
// @id VerifyToken
// @accept application/x-www-form-urlencoded
// @produce json
// @param token formData string true "Access Token to verify"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/auth/verify [post]
func VerifyTokenHandler(c *gin.Context) {
	var input model.VerifyTokenInput

	// This works with:
	// - form-data
	// - x-www-form-urlencoded
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Print("===VerifyTokenHandler===")
	log.Print(input.Token)

	result, err := VerifyToken(input.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	log.Print(result.Active)
	if result.Active {

		//Process login without password

		conn, ctx, cancel := utils.ConnectDB()
		if conn == nil {
			c.JSON(http.StatusInternalServerError, model.Response{
				Status: "-1", Msg: "Failure", Desc: "Cannot connect to database",
			})
			return
		}
		defer cancel()
		defer conn.Close(ctx)

		orgId := os.Getenv("INTEGRATION_ORG_ID")

		user, err := utils.GetUserByUsername(c, conn, orgId, result.Username)
		if err != nil {
			log.Printf("Error: %v", err)
		}

		if user == nil {
			log.Printf("User not found")
		} else {
			pass_, _ := decrypt(user.Password)
			loginReq := model.Login{
				Username:       user.Username,
				Password:       pass_, // ⚠️ user.password should be encrypted
				OrganizationId: &orgId,
			}

			resp, err := LoginUser(ctx, c, conn, loginReq)
			if err != nil {
				c.JSON(http.StatusUnauthorized, resp)
				return
			}

			c.JSON(http.StatusOK, resp)
		}
	} else {
		resp := model.Response{Status: "-1", Desc: fmt.Sprintf("%t", result.Active)}
		c.JSON(http.StatusOK, resp)
	}

}

func VerifyToken(token string) (*model.TokenIntrospectResponse, error) {
	url := os.Getenv("AUTH_SERVER")

	form := map[string]string{
		"client_id":     os.Getenv("AUTH_CLIENT_ID"),
		"client_secret": os.Getenv("AUTH_CLETNT_SECRET"),
		"token":         token,
	}
	log.Print("==VerifyToken==")

	resp, err := callVerify(url, form)
	if err != nil {
		return nil, err
	}

	var out model.TokenIntrospectResponse
	if err := json.Unmarshal([]byte(resp), &out); err != nil {
		return nil, err
	}

	return &out, nil
}

func LoginUser(
	ctx context.Context,
	c *gin.Context,
	conn *pgx.Conn,
	req model.Login,
) (model.Response, error) {

	//logger := utils.GetLog()
	start_time := time.Now()
	txtId := uuid.New().String()
	log.Print("===LoginUser==")
	log.Print(req)
	// ========= Find Organization =========
	var orgId string
	if req.OrganizationId != nil {
		orgId = *req.OrganizationId
	} else {
		err := conn.QueryRow(ctx,
			`SELECT id FROM public.organizations WHERE name=$1`,
			req.Organization,
		).Scan(&orgId)

		if err != nil {
			resp := model.Response{Status: "-1", Msg: "Failure", Desc: err.Error()}
			_ = utils.InsertAuditLogs(
				c, conn, "", req.Username,
				txtId, "", "Authentication", "UserLoginPost", "",
				"login", -1, start_time, GetQueryParams(c), resp, "org not found",
			)
			return resp, err
		}
	}

	// ========= Find User =========
	var user model.Um_User_Login
	err := conn.QueryRow(ctx, `
SELECT u.id, u."orgId", u."displayName", u.title,
       u."firstName", u."middleName", u."lastName",
       u."citizenId", u.bod, u.blood, u.gender,
       u."mobileNo", u.address, u.photo, u.username,
       u.password, u.email, u."roleId", u."userType",
       u."empId", u."deptId", u."commId", u."stnId",
       u.active, u."activationToken", u."lastActivationRequest",
       u."lostPasswordRequest", u."signupStamp",
       u.islogin, u."lastLogin", u."createdAt",
       u."updatedAt", u."createdBy", u."updatedBy",
       COALESCE(a."distIdLists", '[]'::jsonb)
FROM public.um_users u
LEFT JOIN public.um_user_with_area_response a
    ON a.username = u.username
   AND a."orgId" = u."orgId"
WHERE u.username=$1 AND u."orgId"=$2 AND u.active=true;
`, req.Username, orgId).Scan(
		&user.ID, &user.OrgID, &user.DisplayName, &user.Title,
		&user.FirstName, &user.MiddleName, &user.LastName,
		&user.CitizenID, &user.Bod, &user.Blood, &user.Gender,
		&user.MobileNo, &user.Address, &user.Photo, &user.Username,
		&user.Password, &user.Email, &user.RoleID, &user.UserType,
		&user.EmpID, &user.DeptID, &user.CommID, &user.StnID,
		&user.Active, &user.ActivationToken, &user.LastActivationRequest,
		&user.LostPasswordRequest, &user.SignupStamp,
		&user.IsLogin, &user.LastLogin, &user.CreatedAt,
		&user.UpdatedAt, &user.CreatedBy, &user.UpdatedBy,
		&user.DistIdLists,
	)

	if err != nil {
		resp := model.Response{Status: "-1", Msg: "Failure", Desc: err.Error()}
		_ = utils.InsertAuditLogs(
			c, conn, orgId, req.Username,
			txtId, "", "Authentication", "UserLoginPost", "",
			"login", -1, start_time, GetQueryParams(c), resp, "user not found",
		)
		return resp, err
	}

	// ========= Decrypt Password =========
	dec, err := decrypt(user.Password)
	if err != nil || subtle.ConstantTimeCompare([]byte(dec), []byte(req.Password)) != 1 {
		resp := model.Response{Status: "-1", Desc: "Invalid credentials"}
		_ = utils.InsertAuditLogs(
			c, conn, orgId, req.Username,
			txtId, "", "Authentication", "UserLoginPost", "",
			"login", -1, start_time, GetQueryParams(c), resp, "invalid pass",
		)
		return resp, err
	}

	// ========= Create Token =========
	accessToken, refreshToken, err := CreateToken(user.Username, orgId)
	if err != nil {
		resp := model.Response{Status: "-1", Desc: "Token creation failed"}
		return resp, err
	}

	// ========= Load Permissions =========
	rows, err := conn.Query(ctx,
		`SELECT "permId" FROM public.um_role_with_permissions WHERE "orgId"=$1 AND "roleId"=$2 AND active=true`,
		orgId, user.RoleID,
	)
	if err != nil {
		resp := model.Response{Status: "-1", Desc: "Permission failed"}
		return resp, err
	}

	var perm []string
	for rows.Next() {
		var pid string
		_ = rows.Scan(&pid)
		perm = append(perm, pid)
	}
	user.Permission = perm

	// ========= Cache area =========
	utils.GetAreaByUsernameOrLoad(ctx, conn, orgId, user.Username)

	resp := model.Response{
		Status: "0",
		Msg:    "Success",
		Data: map[string]any{
			"accessToken":  accessToken,
			"refreshToken": refreshToken,
			"token_type":   "bearer",
			"user":         user,
		},
	}

	_ = utils.InsertAuditLogs(
		c, conn, orgId, req.Username,
		txtId, "", "Authentication", "UserLoginPost", "",
		"login", 0, start_time, GetQueryParams(c), resp, "success",
	)

	return resp, nil
}
