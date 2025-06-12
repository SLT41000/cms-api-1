package handler

import (
	"crypto/subtle"
	"fmt"
	"mainPackage/config"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

var secretKey = []byte("secret-key")
var TIMEOUT = time.Hour

// Login godoc
// @summary Login
// @description Login to the system and get an access token
// @tags authAPI
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
// @response 201 {object} model.OutputTokenModel "Created - Resource created successfully"
// @response 400 {object} model.OutputTokenModel "Bad Request - Invalid request parameters"
// @response 401 {object} model.OutputTokenModel "Unauthorized - Invalid or missing authentication"
// @response 403 {object} model.OutputTokenModel "Forbidden - Insufficient permissions"
// @response 404 {object} model.OutputTokenModel "Not Found - Resource doesn't exist"
// @response 422 {object} model.OutputTokenModel "Bad Request and Not Found (temporary)"
// @response 429 {object} model.OutputTokenModel "Too Many Requests - Rate limit exceeded"
// @response 500 {object} model.OutputTokenModel "Internal Server Error"
// @Router /api/v1/AuthAPI/token [post]
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"customerName": username,
			"message":      err.Error(),
		})
		return
	}
	if subtle.ConstantTimeCompare([]byte(dbPassword), []byte(password)) == 1 {
		tokenString, err := CreateToken(username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
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

func CreateToken(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": username,
			"exp":      time.Now().Add(TIMEOUT).Unix(),
		})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func verifyToken(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return err
	}

	if !token.Valid {
		return fmt.Errorf("invalid token")
	}

	return nil
}

func ProtectedHandler(c *gin.Context) {
	logger := config.GetLog()
	logger.Debug("Authorization header", zap.String("Authorization", c.GetHeader("Authorization")))
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization header"})
		logger.Debug("Missing authorization header")
		c.Abort()
		return
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
		logger.Debug("Invalid authorization header format")
		c.Abort()
		return
	}

	token := authHeader[len(prefix):]

	// Verify token
	err := verifyToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token", "message": err.Error()})
		logger.Debug("Invalid token")
		c.Abort()
		return
	}

	c.Next()
}
