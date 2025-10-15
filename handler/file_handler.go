package handler

import (
	"context"
	"fmt"
	"mainPackage/model"
	"mainPackage/utils"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
)

// @summary Upload file
// @description Upload a file to MinIO under specified path (e.g. /upload/profile)
// @tags Files
// @security ApiKeyAuth
// @accept multipart/form-data
// @produce json
// @Param path path string true "path"
// @param file formData file true "File to upload"
// @Param caseId formData string false "Case ID to link file (optional)"
// @response 200 {object} model.Response "OK - Request successful"
// @router /api/v1/upload/{path} [post]
func UploadFile(c *gin.Context) {
	logger := utils.GetLog()
	orgId := GetVariableFromToken(c, "orgId")
	username := GetVariableFromToken(c, "username")

	path := c.Param("path") // e.g. "profile"
	caseId := c.DefaultPostForm("caseId", "")

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Status: "-1",
			Msg:    "❌ File is required",
			Desc:   err.Error(),
		})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1",
			Msg:    "❌ Failed to open file",
			Desc:   err.Error(),
		})
		return
	}
	defer src.Close()

	// Generate unique filename
	attId := uuid.New().String()
	ext := filepath.Ext(file.Filename)
	uuidFilename := attId + ext
	objectName := filepath.Join(path, uuidFilename)
	bucket := os.Getenv("MINIO_BUCKET")

	// Upload to MinIO
	_, err = utils.MinioClient.PutObject(
		context.Background(),
		bucket,
		objectName,
		src,
		file.Size,
		minio.PutObjectOptions{ContentType: file.Header.Get("Content-Type")},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1",
			Msg:    "❌ Failed to upload file",
			Desc:   err.Error(),
		})
		return
	}

	// Generate public URL
	url := fmt.Sprintf("https://%s/%s/%s", os.Getenv("MINIO_API"), bucket, objectName)

	// Build attachment struct
	attachment := model.TixCaseAttachment{
		CaseId:  caseId,
		Type:    path,
		AttId:   attId,
		AttName: uuidFilename,
		AttUrl:  url,
	}

	// Insert into DB if caseId provided
	if caseId != "" {
		conn, ctx, cancel := utils.ConnectDB()
		if conn == nil {
			c.JSON(http.StatusInternalServerError, model.Response{
				Status: "-1",
				Msg:    "❌ Database connection failed",
			})
			return
		}
		defer cancel()
		defer conn.Close(ctx)

		query := `
			INSERT INTO tix_case_attachments 
				("orgId", "caseId", "type", "attId", "attName", "attUrl", "createdBy", "updatedBy")
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING "id", "createdAt", "updatedAt"
		`
		err := conn.QueryRow(ctx, query,
			orgId.(string), attachment.CaseId, attachment.Type, attachment.AttId,
			attachment.AttName, attachment.AttUrl, username.(string), username.(string),
		).Scan(&attachment.Id, &attachment.CreatedAt, &attachment.UpdatedAt)

		if err != nil {
			logger.Error("Insert attachment failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.Response{
				Status: "-1",
				Msg:    "❌ Failed to insert attachment record",
				Desc:   err.Error(),
			})
			return
		}
	}

	// ✅ Success
	c.JSON(http.StatusOK, model.Response{
		Status: "0",
		Msg:    "✅ File uploaded successfully",
		Desc:   fmt.Sprintf("path=%s filename=%s", path, uuidFilename),
		Data:   attachment,
	})
}

// @Summary Delete file
// @Description Delete a file from MinIO and optionally from Database
// @Tags Files
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param Body body model.DeleteFileRequest true "Delete file"
// @Success 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/delete [delete]
func DeleteFile(c *gin.Context) {
	logger := utils.GetLog()
	orgId := GetVariableFromToken(c, "orgId")

	var req model.DeleteFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Status: "-1",
			Msg:    "Invalid request",
			Desc:   err.Error(),
		})
		return
	}

	objectName := filepath.Join(req.Path, req.Filename)
	bucket := os.Getenv("MINIO_BUCKET")

	// 🔹 1) Delete from MinIO
	err := utils.MinioClient.RemoveObject(context.Background(), bucket, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		logger.Error("MinIO delete failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1",
			Msg:    "Failed to delete file from storage",
			Desc:   err.Error(),
		})
		return
	}

	// 🔹 2) If caseId and attId provided → delete from DB
	if req.CaseId != "" && req.AttId != "" {
		conn, ctx, cancel := utils.ConnectDB()
		if conn == nil {
			c.JSON(http.StatusInternalServerError, model.Response{
				Status: "-1",
				Msg:    "DB connection failed",
			})
			return
		}
		defer cancel()
		defer conn.Close(ctx)

		query := `
			DELETE FROM tix_case_attachments
			WHERE "orgId" = $1 AND "caseId" = $2 AND "attId" = $3
		`
		cmdTag, err := conn.Exec(ctx, query, orgId, req.CaseId, req.AttId)
		if err != nil {
			logger.Error("DB delete failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.Response{
				Status: "-1",
				Msg:    "Failed to delete DB record",
				Desc:   err.Error(),
			})
			return
		}

		if cmdTag.RowsAffected() == 0 {
			logger.Warn("No DB record found for attachment", zap.String("attId", req.AttId))
		}
	}

	// 🔹 3) Success response
	c.JSON(http.StatusOK, model.Response{
		Status: "0",
		Msg:    "🗑️ File deleted successfully",
		Desc:   fmt.Sprintf("path=%s filename=%s", req.Path, req.Filename),
	})
}

func InsertCaseAttachments(ctx context.Context, conn *pgx.Conn, orgId, caseId, username string, attachments []model.TixCaseAttachmentInput, logger *zap.Logger) error {
	if len(attachments) == 0 {
		return nil
	}

	for _, att := range attachments {
		query := `
			INSERT INTO tix_case_attachments
				("orgId", "caseId", "type", "attId", "attName", "attUrl", "createdBy", "updatedBy")
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`
		_, err := conn.Exec(ctx, query,
			orgId, caseId, att.Type, att.AttId, att.AttName, att.AttUrl, username, username,
		)
		if err != nil {
			logger.Error("Insert attachment failed", zap.Error(err))
			return err
		}
	}

	return nil
}

// GetCaseAttachments returns a list of attachments for a given orgId and caseId
func GetCaseAttachments(ctx context.Context, conn *pgx.Conn, orgId string, caseId string) ([]model.TixCaseAttachment, error) {
	if orgId == "" || caseId == "" {
		return nil, fmt.Errorf("orgId and caseId are required")
	}

	query := `
		SELECT "type", "attId", "attName", "attUrl"
		FROM tix_case_attachments
		WHERE "orgId" = $1 AND "caseId" = $2
	`

	rows, err := conn.Query(ctx, query, orgId, caseId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attachments []model.TixCaseAttachment
	for rows.Next() {
		var att model.TixCaseAttachment
		if err := rows.Scan(
			&att.Type, &att.AttId, &att.AttName, &att.AttUrl,
		); err != nil {
			return nil, err
		}
		attachments = append(attachments, att)
	}

	return attachments, nil
}
