package handler

import (
	"crypto/subtle"
	"fmt"
	"mainPackage/config"
	"mainPackage/model"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// @summary Get Token
// @tags Authentication
// @security ApiKeyAuth
// @id Login
// @accept json
// @produce json
// @Param grantType query string false "grantType"
// @Param username query string true "username"
// @Param password query string true "password"
// @Param scope query string false "scope"
// @Param clientId query string false "clientId"
// @Param clientSecret query string false "clientSecret"
// @response 200 {object} model.OutputTokenModel "OK - Request successful"
// @Router /api/v1/AuthAPI/token [get]
func GetToken(c *gin.Context) {
	logger := config.GetLog()
	username := c.Query("username")
	password := c.Query("password")
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	var dbPassword string
	err := conn.QueryRow(ctx, `SELECT password FROM public."uc_users" WHERE username = $1`, username).Scan(&dbPassword)
	if err != nil {
		logger.Debug(err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{
			"customerName": username,
			"message":      err.Error(),
		})
		return
	}
	if subtle.ConstantTimeCompare([]byte(dbPassword), []byte(password)) == 1 {
		tokenString, err := CreateToken(username, "")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Token creation failed",
				"message": err.Error(),
			})
			logger.Debug(err.Error())
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"accessToken": tokenString,
			"token_type":  "bearer",
		})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
	}

	logger.Debug("User : " + username)
}

// Login godoc
// @summary Login
// @tags Authentication
// @security ApiKeyAuth
// @id Login User
// @accept json
// @produce json
// @Param username query string true "username"
// @Param password query string true "password"
// @Param organization query string true "organization"
// @response 200 {object} model.OutputTokenModel "OK - Request successful"
// @response 201 {object} model.OutputTokenModel "Created - Resource created successfully"
// @response 400 {object} model.OutputTokenModel "Bad Request - Invalid request parameters"
// @response 401 {object} model.OutputTokenModel "Unauthorized - Invalid or missing authentication"
// @response 403 {object} model.OutputTokenModel "Forbidden - Insufficient permissions"
// @response 404 {object} model.OutputTokenModel "Not Found - Resource doesn't exist"
// @response 422 {object} model.OutputTokenModel "Bad Request and Not Found (temporary)"
// @response 429 {object} model.OutputTokenModel "Too Many Requests - Rate limit exceeded"
// @response 500 {object} model.OutputTokenModel "Internal Server Error"
// @Router /api/v1/AuthAPI/login [get]
func UserLogin(c *gin.Context) {
	logger := config.GetLog()
	username := c.Query("username")
	password := c.Query("password")
	organization := c.Query("organization")
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	var dbPassword string
	var id string
	query := `SELECT id FROM public.organizations WHERE name = $1`
	logger.Debug(`Query`, zap.String("query", query))
	logger.Debug(`request input`, zap.Any("organization", organization))
	err := conn.QueryRow(ctx, query, organization).Scan(&id)
	if err != nil {
		logger.Debug(err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{
			"organization": organization,
			"message":      err.Error(),
		})
		return
	}
	if id == "" {
		logger.Debug("organization not found")
		c.JSON(http.StatusUnauthorized, gin.H{
			"organization": organization,
			"message":      "organization not found",
		})
		return
	}
	query = `SELECT "passwordHash" FROM public.users WHERE username = $1 AND active = true`
	logger.Debug(`Query`, zap.String("query", query))
	logger.Debug(`request input`, zap.Any("username", username))
	err = conn.QueryRow(ctx, query, username).Scan(&dbPassword)
	if err != nil {
		logger.Debug(err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{
			"customerName": username,
			"message":      err.Error(),
		})
		return
	}
	var dec string
	dec, err = decrypt(dbPassword)
	if err != nil {
		logger.Debug(err.Error())
		return
	}

	if subtle.ConstantTimeCompare([]byte(dec), []byte(password)) == 1 {
		tokenString, err := CreateToken(username, id)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Token creation failed",
				"message": err.Error(),
			})
			logger.Debug(err.Error())
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"accessToken": tokenString,
			"token_type":  "bearer",
		})
	} else {
		c.JSON(http.StatusUnauthorized, model.Response{
			Status: "-1",
			Msg:    "",
			Desc:   "Invalid credentials",
		})
	}

	logger.Debug("User : " + username)
}

func CreateToken(username string, orgId string) (string, error) {

	var secretKey = []byte(os.Getenv("TOKEN_SECRET_KEY"))
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
		return "", err
	}

	return tokenString, nil
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

func ProtectedHandler(c *gin.Context) {
	logger := config.GetLog()
	authHeader := c.GetHeader("Authorization")
	logger.Debug("Authorization header", zap.String("Authorization", authHeader))

	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization header"})
		c.Abort()
		return
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
		c.Abort()
		return
	}

	tokenString := strings.TrimPrefix(authHeader, prefix)
	logger.Debug("Token: " + tokenString)

	parsedToken, err := verifyToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token", "message": err.Error()})
		c.Abort()
		return
	}

	// Extract claims
	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
		if username, ok := claims["username"].(string); ok {
			logger.Debug("Verified username", zap.String("username", username))
			c.Set("username", username)
			c.Next()
			return
		}
	}
}

// Login godoc
// @summary Create User
// @tags Authentication
// @security ApiKeyAuth
// @id Create User
// @accept json
// @produce json
// @param Case body model.UserInputModel true "User to be created"
// @response 200 {object} model.OutputTokenModel "OK - Request successful"
// @Router /api/v1/AuthAPI/add [post]
func UserAdd(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	var req model.UserInputModel
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
	var enc string
	var err error
	enc, err = encrypt(req.PasswordHash)
	if err != nil {
		return
	}

	query := `
		INSERT INTO public.users(
		"orgId", "userId", "displayName", "fullName", "phoneNumber", email, username, "passwordHash"
		, "lastLogin", "roleId", active, "areaId", "deviceId", "pushToken", "currentLat", "currentLon"
		, "createdAt", "updatedAt", "createdBy", "updatedBy"
			)
	VALUES (
		$1, $2, $3, $4, $5, $6, $7,
		$8, $9, $10, $11,
		$12, $13, $14, $15, $16,
		$17, $18, $19, $20
	)
	RETURNING id;
	`
	logger.Debug(`Query`, zap.String("query", query))
	logger.Debug(`request input`, zap.Any("Input", []any{req}))
	logger.Debug(`Encrypt Password :` + enc)
	err = conn.QueryRow(ctx, query,
		req.OrgID, req.UserID, req.DisplayName, req.FullName, req.PhoneNumber,
		req.Email, req.Username, enc, req.LastLogin, req.RoleID,
		req.Active, req.AreaID, req.DeviceID, req.PushToken, req.CurrentLat,
		req.CurrentLon, req.CreatedAt, req.UpdatedAt, req.CreatedBy, req.UpdatedBy,
	).Scan(&cust.ID)

	if err != nil {
		// log.Printf("Insert failed: %v", err)
		c.JSON(http.StatusUnauthorized, model.CaseTransactionCRUDResponse{
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
		Desc:   "Create user successfully",
		ID:     cust.ID,
	})
}
