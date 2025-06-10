package handler

import (
	"context"
	"crypto/subtle"
	"fmt"
	"log"
	"mainPackage/model"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
)

var secretKey = []byte("secret-key")

type AuthHandler struct {
	DB *pgx.Conn
}

// Login godoc
// @summary Login
// @description Login to the system and get an access token
// @tags authAPI
// @security ApiKeyAuth
// @id Login
// @accept json
// @produce json
// @param Token body model.InputTokenModel true "Token data"
// @response 200 {object} model.OutputTokenModel "OK - Request successful"
// @response 201 {object} model.OutputTokenModel "Created - Resource created successfully"
// @response 400 {object} model.OutputTokenModel "Bad Request - Invalid request parameters"
// @response 401 {object} model.OutputTokenModel "Unauthorized - Invalid or missing authentication"
// @response 403 {object} model.OutputTokenModel "Forbidden - Insufficient permissions"
// @response 404 {object} model.OutputTokenModel "Not Found - Resource doesn't exist"
// @response 422 {object} model.OutputTokenModel "Bad Request and Not Found (temporary)"
// @response 429 {object} model.OutputTokenModel "Too Many Requests - Rate limit exceeded"
// @response 500 {object} model.OutputTokenModel "Internal Server Error"
// @Router /api/v1/authAPI/token [post]
func (h *AuthHandler) GetToken(c *gin.Context) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var u model.InputTokenModel

	// Bind incoming JSON to struct
	if err := c.ShouldBindJSON(&u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	var password string
	err := h.DB.QueryRow(ctx, `SELECT password FROM public."GoLangDemo" WHERE name = $1`, u.Username).Scan(&password)
	if err != nil {
		log.Printf("Query failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"customerName": u.Username,
			"message":      err.Error(),
		})
		return
	}
	if subtle.ConstantTimeCompare([]byte(password), []byte(u.Password)) == 1 {
		tokenString, err := CreateToken(u.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Token creation failed",
				"message": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"accessToken": tokenString,
			"token_type":  "bearer",
		})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
	}

	fmt.Printf("The user request value: %+v\n", u)

	// Dummy user check
	// if u.Username == "Chek" && u.Password == "123456" {
	// 	tokenString, err := CreateToken(u.Username)
	// 	if err != nil {
	// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token creation failed",
	// 			"message": err.Error()})
	// 		return
	// 	}

	// 	c.JSON(http.StatusOK, gin.H{
	// 		"accessToken": tokenString,
	// 		"token_type":  "bearer",
	// 	})
	// } else {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
	// }
}

func CreateToken(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": username,
			"exp":      time.Now().Add(time.Hour * 24).Unix(),
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
	fmt.Print(c.GetHeader("Authorization"))
	authHeader := c.GetHeader("Authorization")
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

	token := authHeader[len(prefix):]

	// Verify token
	err := verifyToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		c.Abort()
		return
	}

	c.Next()
}
