package handler

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mainPackage/config"
	"mainPackage/model"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

func ToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int, int8, int16, int32, int64:
		return strconv.FormatInt(reflect.ValueOf(v).Int(), 10)
	case uint, uint8, uint16, uint32, uint64:
		return strconv.FormatUint(reflect.ValueOf(v).Uint(), 10)
	case float32, float64:
		return strconv.FormatFloat(reflect.ValueOf(v).Float(), 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", value)
	}
}

func ToInt(value interface{}) int {
	switch v := value.(type) {
	case int:
		return v
	case int8, int16, int32, int64:
		return int(reflect.ValueOf(v).Int())
	case uint, uint8, uint16, uint32, uint64:
		return int(reflect.ValueOf(v).Uint())
	case float32, float64:
		return int(reflect.ValueOf(v).Float())
	case string:
		num, _ := strconv.Atoi(strings.TrimSpace(v))
		return num // Returns 0 if string isn't a number
	case bool:
		if v {
			return 1
		}
		return 0
	default:
		return 0
	}
}

func deriveKey(passphrase string) []byte {
	hash := sha256.Sum256([]byte(passphrase))
	return hash[:] // 32 bytes
}

func encrypt(plaintext string) (string, error) {
	key := deriveKey(os.Getenv("SECRET_KEY"))

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)

	// base64-encode so it's printable
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decrypt(ciphertextBase64 string) (string, error) {
	// logger := config.GetLog()
	key := deriveKey(os.Getenv("SECRET_KEY"))

	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {

		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func unmarshalToMap(data []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func unmarshalToSliceOfMaps(data []byte) ([]map[string]interface{}, error) {
	var result []map[string]interface{}
	err := json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func CaseCurrentStageInsert(conn *pgx.Conn, ctx context.Context, c *gin.Context, req model.CustomCaseCurrentStage) error {
	logger := config.GetLog()

	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	now := time.Now()

	query := `SELECT t1.id, t1."orgId", t1."wfId", t1."nodeId", t1.versions, t1.type, t1.section, t1.data, t1.pic, t1."group", t1."formId",t1. "createdAt", t1."updatedAt", t1."createdBy", t1."updatedBy" FROM public.wf_nodes t1 JOIN public.wf_definitions t2 ON t1."versions"=t2."versions" AND t1."wfId"=t2."wfId"
	WHERE t2."wfId"=$1 AND t2."nodeId"=$2 AND t2."orgId"=$3`

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query),
		zap.Any("Input", []any{req.WfID, req.NodeID, orgId}))
	rows, err := conn.Query(ctx, query, orgId, req.WfID, req.NodeID, orgId)
	if err != nil {
		logger.Warn("Query failed", zap.Error(err))
		return err
	}
	defer rows.Close()
	logger.Debug(`Query`, zap.String("query", query))
	var Workflow model.WfNode
	err = rows.Scan(&Workflow.ID, &Workflow.OrgID, &Workflow.WfID, &Workflow.NodeID, &Workflow.Versions, &Workflow.Type, &Workflow.Section,
		&Workflow.Data, &Workflow.Pic, &Workflow.Group, &Workflow.FormID, &Workflow.CreatedAt, &Workflow.UpdatedAt, &Workflow.CreatedBy, &Workflow.UpdatedBy)
	if err != nil {
		logger.Warn("Query failed", zap.Error(err))
		return err
	}
	query = `
	INSERT INTO public.tix_case_current_stage(
	"orgId", "caseId", "wfId", "nodeId", versions, type, section, data, pic, "group", "formId", "createdAt", "updatedAt", "createdBy", "updatedBy")
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	logger.Debug(`Query`, zap.String("query", query), zap.Any("req", req))
	_, err = conn.Exec(ctx, query,
		&Workflow.OrgID, req.CaseID, &Workflow.WfID, req.NodeID, &Workflow.Versions, &Workflow.Type, &Workflow.Section,
		&Workflow.Data, &Workflow.Pic, &Workflow.Group, &Workflow.FormID, now, now, username, username,
	)

	if err != nil {
		return err
	}
	return nil
}
