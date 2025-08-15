package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"mainPackage/config"
	"mainPackage/model"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype" // Import pgtype to handle nullable fields
)

// --- Helper Functions ---

// RandomString generates a random string of a fixed length.
func RandomString(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

// --- API Handlers ---

// CreateNotifications godoc
// @Summary Create one or more new notifications
// @Description Creates a batch of notifications, saves them to the database in a single transaction, and broadcasts each one to relevant online users. The input should be a JSON array of notification objects.
// @Tags Notifications
// @security ApiKeyAuth
// @Accept json
// @Produce json
// @Param notifications body []model.Notification true "A JSON array of notification objects. The `data` field should be an array of key-value objects, e.g., `\"data\": [{\"key\": \"caseId\", \"value\": \"C1122\"}]`"
// @Success 201 {array} model.Notification "Notifications created successfully"
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 500 {object} map[string]string "Internal server error (e.g., database transaction failure)"
// @Router /api/v1/notifications [post]
func CreateNotifications(c *gin.Context) {
	var inputs []model.Notification
	if err := c.ShouldBindJSON(&inputs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "detail": err.Error()})
		return
	}

	createdNotifications, err := CoreNotifications(c.Request.Context(), inputs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdNotifications)
}

// UpdateNotification godoc
// @Summary Update an existing notification
// @Description Updates the content of a specific notification by its ID.
// @Tags Notifications
// @security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path integer true "Notification ID"
// @Param notification body model.Notification true "Notification object with fields to update."
// @Success 200 {object} model.Notification "Notification updated successfully"
// @Failure 400 {object} map[string]string "Invalid request body or ID"
// @Failure 404 {object} map[string]string "Notification not found"
// @Failure 500 {object} map[string]string "Database error"
// @Router /api/v1/notifications/{id} [put]
func UpdateNotification(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification ID format"})
		return
	}

	var input model.Notification
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "detail": err.Error()})
		return
	}

	recipientsJSON, err := json.Marshal(input.Recipients)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process recipients", "detail": err.Error()})
		return
	}
	dataJSON, err := json.Marshal(input.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process custom data", "detail": err.Error()})
		return
	}

	conn, ctx, cancel := config.ConnectDB()
	defer cancel()
	defer conn.Close(ctx)

	// UPDATED: SQL UPDATE statement to include new updatable fields like expiredAt
	tag, err := conn.Exec(ctx, `
        UPDATE notifications
        SET "message" = $1, "eventType" = $2, "redirectUrl" = $3, "recipients" = $4, "data" = $5, "expiredAt" = $6
        WHERE "id" = $7
    `, input.Message, input.EventType, input.RedirectUrl, recipientsJSON, dataJSON, input.ExpiredAt, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database update failed", "detail": err.Error()})
		return
	}

	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "notification not found with the given ID"})
		return
	}

	var updatedNoti model.Notification
	var recipientsStr, dataStr []byte
	// FIX: Use pgtype for nullable fields to prevent scan errors
	var senderPhoto, createdBy pgtype.Text
	var expiredAt pgtype.Timestamptz

	// UPDATED: SQL SELECT to retrieve all fields from the new model
	err = conn.QueryRow(ctx, `
		SELECT "id", "orgId", "senderType", "sender", "senderPhoto", "message", "eventType", "redirectUrl", "createdAt", "createdBy", "expiredAt", "recipients", "data" 
		FROM notifications WHERE "id" = $1
	`, id).Scan(
		&updatedNoti.ID, &updatedNoti.OrgID, &updatedNoti.SenderType, &updatedNoti.Sender, &senderPhoto, &updatedNoti.Message,
		&updatedNoti.EventType, &updatedNoti.RedirectUrl, &updatedNoti.CreatedAt, &createdBy, &expiredAt, &recipientsStr, &dataStr,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve updated notification", "detail": err.Error()})
		return
	}

	// FIX: Assign values from pgtype variables to the struct if they are valid (not NULL)
	if senderPhoto.Valid {
		updatedNoti.SenderPhoto = senderPhoto.String
	}
	if createdBy.Valid {
		updatedNoti.CreatedBy = createdBy.String
	}
	if expiredAt.Valid {
		updatedNoti.ExpiredAt = expiredAt.Time
	}

	json.Unmarshal(recipientsStr, &updatedNoti.Recipients)
	json.Unmarshal(dataStr, &updatedNoti.Data)

	c.JSON(http.StatusOK, updatedNoti)
}

// DeleteNotification godoc
// @Summary Delete a notification
// @Description Deletes a notification by its ID.
// @Tags Notifications
// @security ApiKeyAuth
// @Produce json
// @Param id path integer true "Notification ID"
// @Success 200 {object} map[string]string "Message confirming deletion"
// @Failure 404 {object} map[string]string "Notification not found"
// @Failure 500 {object} map[string]string "Database error"
// @Router /api/v1/notifications/{id} [delete]
func DeleteNotification(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification ID format"})
		return
	}

	conn, ctx, cancel := config.ConnectDB()
	defer cancel()
	defer conn.Close(ctx)

	tag, err := conn.Exec(ctx, `DELETE FROM notifications WHERE "id" = $1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database delete failed", "detail": err.Error()})
		return
	}

	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "notification not found with the given ID"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "notification deleted successfully", "id": id})
}

// GetNotificationsForUser godoc
// @Summary Get all notifications for a specific user
// @Description Retrieves all notifications intended for a user by their username and organization ID.
// @Tags Notifications
// @security ApiKeyAuth
// @Produce json
// @Param orgId path string true "Organization ID of the user"
// @Param username path string true "Username to fetch notifications for"
// @Success 200 {array} model.Notification
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Database error"
// @Router /api/v1/notifications/{orgId}/{username} [get]
func GetNotificationsForUser(c *gin.Context) {
	orgId := c.Param("orgId")
	username := c.Param("username")
	log.Printf("Fetching notifications for username: %s in org: %s", username, orgId)

	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not connect to the database"})
		return
	}
	defer cancel()
	defer conn.Close(ctx)

	// Step 1: Get the user's full profile using their username and orgId.
	var userProfile model.Um_User
	err := conn.QueryRow(ctx, `SELECT "empId", "orgId", "roleId", "deptId", "stnId", "commId" FROM um_users WHERE "username" = $1 AND "orgId" = $2 AND "active" = true`, username, orgId).Scan(
		&userProfile.EmpID, &userProfile.OrgID, &userProfile.RoleID, &userProfile.DeptID, &userProfile.StnID, &userProfile.CommID,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found in the specified organization"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user profile", "detail": err.Error()})
		return
	}

	// Step 2: Build query conditions and arguments safely.
	conditions := []string{}
	args := []interface{}{userProfile.OrgID} // $1 is always orgId

	// Helper function to create a JSON string for a recipient rule.
	createRecipientJSON := func(recipientType, value string) (string, error) {
		recipient := []model.Recipient{{Type: recipientType, Value: value}}
		jsonBytes, err := json.Marshal(recipient)
		return string(jsonBytes), err
	}

	// Condition for EMP_ID
	if empIdJson, err := createRecipientJSON("empId", userProfile.EmpID); err == nil {
		args = append(args, empIdJson)
		conditions = append(conditions, fmt.Sprintf(`"recipients"::jsonb @> $%d::jsonb`, len(args)))
	}

	// Condition for ROLE_ID
	if userProfile.RoleID != "" {
		if roleIdJson, err := createRecipientJSON("roleId", userProfile.RoleID); err == nil {
			args = append(args, roleIdJson)
			conditions = append(conditions, fmt.Sprintf(`"recipients"::jsonb @> $%d::jsonb`, len(args)))
		}
	}

	// Condition for DEPARTMENT_ID
	if userProfile.DeptID != "" {
		if deptIdJson, err := createRecipientJSON("deptId", userProfile.DeptID); err == nil {
			args = append(args, deptIdJson)
			conditions = append(conditions, fmt.Sprintf(`"recipients"::jsonb @> $%d::jsonb`, len(args)))
		}
	}

	// Condition for STATION_ID
	if userProfile.StnID != "" {
		if stnIdJson, err := createRecipientJSON("stnId", userProfile.StnID); err == nil {
			args = append(args, stnIdJson)
			conditions = append(conditions, fmt.Sprintf(`"recipients"::jsonb @> $%d::jsonb`, len(args)))
		}
	}

	// Condition for COMMUNITY_ID
	if userProfile.CommID != "" {
		if commIdJson, err := createRecipientJSON("commId", userProfile.CommID); err == nil {
			args = append(args, commIdJson)
			conditions = append(conditions, fmt.Sprintf(`"recipients"::jsonb @> $%d::jsonb`, len(args)))
		}
	}

	// Condition for ORG_ID (everyone in the organization)
	if orgIdJson, err := createRecipientJSON("orgId", userProfile.OrgID); err == nil {
		args = append(args, orgIdJson)
		conditions = append(conditions, fmt.Sprintf(`"recipients"::jsonb @> $%d::jsonb`, len(args)))
	}

	// Condition for USERNAME (userName)
	if usernameJson, err := createRecipientJSON("username", username); err == nil {
		args = append(args, usernameJson)
		conditions = append(conditions, fmt.Sprintf(`"recipients"::jsonb @> $%d::jsonb`, len(args)))
	}

	fullCondition := strings.Join(conditions, " OR ")

	// UPDATED: SQL SELECT to retrieve all fields from the new model
	query := fmt.Sprintf(`
        SELECT "id", "orgId", "senderType", "sender", "senderPhoto", "message", "eventType", "redirectUrl", "createdAt", "createdBy", "expiredAt", "recipients", "data" 
        FROM notifications
        WHERE "orgId" = $1 AND (%s)
        ORDER BY "createdAt" DESC
    `, fullCondition)

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query notifications", "detail": err.Error(), "query": query})
		return
	}
	defer rows.Close()

	var notifications []model.Notification
	for rows.Next() {
		var n model.Notification
		var recipientsStr, dataStr []byte
		// FIX: Use pgtype for nullable fields to prevent scan errors
		var senderPhoto, createdBy pgtype.Text
		var expiredAt pgtype.Timestamptz

		// UPDATED: Scan to accommodate all fields from the new model

		var redirectUrl pgtype.Text // เพิ่มตัวแปรสำหรับรองรับ nullable

		if err := rows.Scan(
			&n.ID, &n.OrgID, &n.SenderType, &n.Sender, &senderPhoto, &n.Message,
			&n.EventType, &redirectUrl, &n.CreatedAt, &createdBy, &expiredAt, &recipientsStr, &dataStr,
		); err != nil {
			log.Printf("Error scanning notification row: %v", err)
			continue
		}

		// assign redirectUrl ถ้ามีค่า
		if redirectUrl.Valid {
			n.RedirectUrl = redirectUrl.String
		}

		if senderPhoto.Valid {
			n.SenderPhoto = senderPhoto.String
		}
		if createdBy.Valid {
			n.CreatedBy = createdBy.String
		}
		if expiredAt.Valid {
			n.ExpiredAt = expiredAt.Time
		}

		json.Unmarshal(recipientsStr, &n.Recipients)
		json.Unmarshal(dataStr, &n.Data)
		notifications = append(notifications, n)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error iterating notification rows", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, notifications)
}

// --- Background Job for Auto-Deletion ---

// DeleteExpiredNotifications connects to the database and deletes all notifications
// where the 'expiredAt' timestamp is in the past.
func DeleteExpiredNotifications() {
	log.Println("Scheduler: Running job to delete expired notifications...")

	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		log.Println("Scheduler Error: could not connect to the database")
		return
	}
	defer cancel()
	defer conn.Close(ctx)

	// Delete notifications where expiredAt is not NULL and is older than the current time
	tag, err := conn.Exec(ctx, `DELETE FROM notifications WHERE "expiredAt" IS NOT NULL AND "expiredAt" < NOW()`)
	if err != nil {
		log.Printf("Scheduler Error: database delete failed: %v", err)
		return
	}

	if tag.RowsAffected() > 0 {
		log.Printf("Scheduler: Successfully deleted %d expired notifications.", tag.RowsAffected())
	} else {
		log.Println("Scheduler: No expired notifications to delete.")
	}
}

// StartAutoDeleteScheduler starts a ticker that runs the DeleteExpiredNotifications
// function at a regular interval (e.g., every hour).
func StartAutoDeleteScheduler() {
	log.Println("Starting background scheduler for auto-deleting notifications...")
	// Run the cleanup job every 1 hour.
	ticker := time.NewTicker(1 * time.Hour)

	// Run forever in the background
	go func() {
		for {
			// Wait for the ticker to fire
			<-ticker.C
			// Run the deletion job
			DeleteExpiredNotifications()
		}
	}()
}
