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

func GetVariableFromToken(c *gin.Context, varname string) interface{} {
	var_return, exists := c.Get(varname)
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return nil
	}
	return var_return
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

			c.Next()
			return
		}
	}

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
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/auth/login [get]
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

	query = `SELECT id,"orgId", "userId", "displayName", "fullName", "phoneNumber", email, username,
	 "passwordHash", "lastLogin", "roleId", active, "areaId", "deviceId", "pushToken", "currentLat",
	  "currentLon", "createdAt", "updatedAt", "createdBy", "updatedBy" FROM public.users WHERE username = $1 AND active = true`
	logger.Debug(`Query`, zap.String("query", query))
	logger.Debug(`request input`, zap.Any("username", username))
	var UserOpt model.User
	err = conn.QueryRow(ctx, query, username).Scan(&UserOpt.ID,
		&UserOpt.OrgID, &UserOpt.UserID, &UserOpt.DisplayName, &UserOpt.FullName, &UserOpt.PhoneNumber, &UserOpt.Email,
		&UserOpt.Username, &UserOpt.PasswordHash, &UserOpt.LastLogin, &UserOpt.RoleID, &UserOpt.Active, &UserOpt.AreaID,
		&UserOpt.DeviceID, &UserOpt.PushToken, &UserOpt.CurrentLat, &UserOpt.CurrentLon, &UserOpt.CreatedAt, &UserOpt.UpdatedAt,
		&UserOpt.CreatedBy, &UserOpt.UpdatedBy)
	if err != nil {
		logger.Debug(err.Error())
		c.JSON(http.StatusUnauthorized, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		return
	}
	var dec string
	dec, err = decrypt(UserOpt.PasswordHash)
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
		c.JSON(http.StatusOK, model.Response{
			Status: "0",
			Msg:    "Success",
			Desc:   "",
			Data: gin.H{
				"accessToken": tokenString,
				"token_type":  "bearer",
				"user":        UserOpt,
			},
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

// Login godoc
// @summary Create User
// @tags Authentication
// @security ApiKeyAuth
// @id Create User
// @accept json
// @produce json
// @param Case body model.Response true "User to be created"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/auth/add [post]
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
		c.JSON(http.StatusBadRequest, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	// now req is ready to use

	var enc string
	var err error
	var id int
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
	).Scan(&id)

	if err != nil {
		// log.Printf("Insert failed: %v", err)
		c.JSON(http.StatusUnauthorized, model.Response{
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
