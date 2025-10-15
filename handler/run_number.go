package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"mainPackage/model"
	"mainPackage/utils"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

func GenerateCaseID(ctx context.Context, conn *pgx.Conn, prefix string) (string, error) {
	today := time.Now().Format("060102") // YYMMDD
	logger := utils.GetLog()
	var lastNumber int
	err := conn.QueryRow(ctx, `
        INSERT INTO case_id_sequences (prefix, date_code, last_number)
        VALUES ($1, $2, 1)
        ON CONFLICT (prefix, date_code)
        DO UPDATE SET last_number = case_id_sequences.last_number + 1, updatedAt = NOW()
        RETURNING last_number
    `, prefix, today).Scan(&lastNumber)

	if err != nil {
		logger.Warn("Query failed", zap.Error(err))
		return "", err
	}

	// format: D250904-00001
	caseID := fmt.Sprintf("%s%s-%05d", prefix, today, lastNumber)
	logger.Debug("caseID: " + caseID)
	return caseID, nil
}

// @summary Generate Case ID
// @tags Public
// @in header
// @name X-API-KEY
// @id Generate Case ID
// @accept json
// @produce json
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/generate_caseid [get]
func GenerateCaseIDHandler(c *gin.Context) {
	log.Print(GenerateSecureAPIKey(16))
	logger := utils.GetLog()

	// ✅ ตรวจสอบ API Key
	apiKey := c.GetHeader("X-API-KEY")
	apiKeyConf := os.Getenv("API_KEY")
	if apiKey != apiKeyConf && apiKeyConf != "" { // ตรงนี้คุณเปลี่ยน "xxx" เป็นค่าที่คุณเก็บใน config ได้
		logger.Warn("X-API-KEY Error : " + apiKey)
		c.JSON(http.StatusUnauthorized, model.Response{
			Status: "-1",
			Msg:    "Unauthorized",
			Desc:   "Invalid API key",
		})
		return
	}
	log.Print("=======GEN====")
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)

	caseID, err := GenerateCaseID(ctx, conn, "I")
	if err != nil {
		logger.Warn("GenerateCaseID : ", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Status: "0",
		Msg:    "Success",
		Data:   map[string]string{"caseId": caseID},
		Desc:   "",
	})
}

func GenerateSecureAPIKey(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
