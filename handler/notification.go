package handler

import (
	"log"
	"mainPackage/config"
	"mainPackage/model"
	"math/rand"
	"net/http"
	"net/smtp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

// genNotificationID สร้าง ID แบบ timestamp + nanoseconds
func genNotificationID() string {
	currentTime := time.Now()
	year := currentTime.Format("06")
	month := currentTime.Format("01") // ใช้ Format("01") ง่ายกว่า int()
	day := currentTime.Format("02")
	hour := currentTime.Format("15")
	minute := currentTime.Format("04")
	second := currentTime.Format("05")
	millisecond := currentTime.Format("0000000") // nanoseconds -> microseconds (อาจต้องปรับถ้าใช้ Format)

	timestamp := "D" + year + month + day + hour + minute + second + millisecond
	return timestamp
}

func randomFromSlice(arr []string) string {
	return arr[rand.Intn(len(arr))]
}

func RandomString(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyz")
	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

// sendEmailNotification ส่งอีเมลแจ้งเตือน
func sendEmailNotification(to, subject, body string) error {
	from := "your_email@gmail.com"    // เปลี่ยนเป็นอีเมลผู้ส่ง
	password := "your_email_password" // เปลี่ยนเป็นรหัสผ่านอีเมล
	host := "smtp.gmail.com"
	port := "587"
	addr := host + ":" + port

	auth := smtp.PlainAuth("", from, password, host)
	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/plain; charset=\"UTF-8\"\r\n" +
		"\r\n" +
		body + "\r\n")

	return smtp.SendMail(addr, auth, from, []string{to}, msg)
}

// GetNotificationByID godoc
// @Summary Get notification by ID from database
// @Tags Notifications
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} model.Notification
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/notifications/noti/{id} [get]
func GetNotificationByID(c *gin.Context) {
	id := c.Param("id")
	conn, ctx, cancel := config.ConnectDB()
	defer cancel()
	defer conn.Close(ctx)

	var notification model.Notification

	err := conn.QueryRow(ctx, `
	SELECT id, "caseId", "caseType", "caseDetail", recipient, sender, message, "eventType", "createdAt", read, "redirectUrl"
	FROM notifications
	WHERE id = $1
	`, id).Scan(
		&notification.ID,
		&notification.CaseID,
		&notification.CaseType,
		&notification.CaseDetail,
		&notification.Recipient,
		&notification.Sender,
		&notification.Message,
		&notification.EventType,
		&notification.CreatedAt,
		&notification.Read,
		&notification.RedirectURL,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "notification not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		}
		return
	}

	c.JSON(http.StatusOK, notification)
}

// GetNotificationsByRecipient godoc
// @Summary Get notifications received by username
// @Tags Notifications
// @Produce json
// @Param username path string true "Username of the recipient"
// @Success 200 {array} model.Notification
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/notifications/recipient/{username} [get]
func GetNotificationsByRecipient(c *gin.Context) {
	username := c.Param("username")
	log.Println("Recipient param:", username)

	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username cannot be empty"})
		return
	}

	conn, ctx, cancel := config.ConnectDB()
	defer cancel()
	defer conn.Close(ctx)

	rows, err := conn.Query(ctx, `
		SELECT id, "caseId", "caseType", "caseDetail", recipient, sender, message, "eventType", "createdAt", read, "redirectUrl"
		FROM notifications
		WHERE recipient = $1
		ORDER BY "createdAt" DESC
	`, username)
	if err != nil {
		log.Println("DB query error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error: " + err.Error()})
		return
	}
	defer rows.Close()

	var notifications []model.Notification
	for rows.Next() {
		var n model.Notification
		err := rows.Scan(
			&n.ID,
			&n.CaseID,
			&n.CaseType,
			&n.CaseDetail,
			&n.Recipient,
			&n.Sender,
			&n.Message,
			&n.EventType,
			&n.CreatedAt,
			&n.Read,
			&n.RedirectURL,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "scan error: " + err.Error()})
			return
		}
		notifications = append(notifications, n)
	}

	c.JSON(http.StatusOK, notifications)
}

// PostNotificationCustom godoc

// @Summary Create notification (partial input)
// @Description Create a notification by providing only partial fields. The remaining fields (e.g., ID, caseId, createdAt) will be generated automatically.
// @Tags Notifications
// @Accept json
// @Produce json
// @Param notification body model.Notification true "Partial Notification Input (Do not include: id, caseId, createdAt)"
// @Success 200 {object} model.Notification
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/notifications/new [post]
func PostNotificationCustom(c *gin.Context) {
	var input model.Notification
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "detail": err.Error()})
		return
	}

	noti := model.Notification{
		ID:          genNotificationID(),
		CaseID:      RandomString(10),
		CaseType:    input.CaseType,
		CaseDetail:  input.CaseDetail,
		Recipient:   input.Recipient,
		Sender:      input.Sender,
		Message:     input.Message,
		EventType:   input.EventType,
		CreatedAt:   time.Now(), // Generate automatically
		Read:        false,
		RedirectURL: input.RedirectURL,
	}

	conn, ctx, cancel := config.ConnectDB()
	defer cancel()
	defer conn.Close(ctx)

	_, err := conn.Exec(ctx, `
		INSERT INTO notifications 
		(id, "caseId", "caseType", "caseDetail", recipient, sender, message, "eventType", "createdAt", read, "redirectUrl")
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`, noti.ID, noti.CaseID, noti.CaseType, noti.CaseDetail,
		noti.Recipient, noti.Sender, noti.Message, noti.EventType,
		noti.CreatedAt, noti.Read, noti.RedirectURL)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "insert error", "detail": err.Error()})
		return
	}

	// 🔔 Send to WebSocket recipient in real-time
	SendNotificationToRecipient(noti)

	// 📨 ส่งอีเมลแจ้งเตือน (สมมติ recipient เป็นอีเมล)
	if noti.Recipient != "" {
		emailSubject := "New Notification: " + noti.EventType
		emailBody := noti.Message
		err := sendEmailNotification(noti.Recipient, emailSubject, emailBody)
		if err != nil {
			log.Println("Failed to send email:", err)
		}
	}

	c.JSON(http.StatusOK, noti)
}

// @Summary edit notification (partial input)
// @Description เเก้ไข Notification
// @Tags Notifications
// @Accept json
// @Produce json
// @Param notification body model.Notification true "Partial Notification Input"
// @Success 200 {object} model.Notification
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/notifications/edit/{id} [put]
func UpdateNotificationByID(c *gin.Context) {
	id := c.Param("id")
	var input model.Notification

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON", "detail": err.Error()})
		return
	}

	conn, ctx, cancel := config.ConnectDB()
	defer cancel()
	defer conn.Close(ctx)

	var exists bool
	err := conn.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM notifications WHERE id = $1)`, id).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "notification not found"})
		return
	}

	_, err = conn.Exec(ctx, `
       	UPDATE notifications 
	SET recipient = $1, sender = $2, message = $3, "eventType" = $4, read = $5,
		"redirectUrl" = $6, "caseType" = $7, "caseDetail" = $8
	WHERE id = $9
    `, input.Recipient, input.Sender, input.Message, input.EventType,
		input.Read, input.RedirectURL, input.CaseType, input.CaseDetail, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed", "detail": err.Error()})
		return
	}

	var updated model.Notification
	err = conn.QueryRow(ctx, `
       	SELECT id, "caseId", "caseType", "caseDetail", recipient, sender, message, "eventType", "createdAt", read, "redirectUrl"
	FROM notifications
	WHERE id = $1
    `, id).Scan(
		&updated.ID,
		&updated.CaseID,
		&updated.CaseType,
		&updated.CaseDetail,
		&updated.Recipient,
		&updated.Sender,
		&updated.Message,
		&updated.EventType,
		&updated.CreatedAt,
		&updated.Read,
		&updated.RedirectURL,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve updated", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

// @Summary delete notification by id
// @Description ลบ Notification ตาม id
// @Tags Notifications
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/notifications/delete/{id} [delete]
func DeleteNotificationByID(c *gin.Context) {
	id := c.Param("id")

	conn, ctx, cancel := config.ConnectDB()
	defer cancel()
	defer conn.Close(ctx)

	// ตรวจสอบว่า notification มีอยู่จริงก่อน
	var exists bool
	err := conn.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM notifications WHERE id = $1)`, id).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error", "detail": err.Error()})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "notification not found"})
		return
	}

	// ลบ notification
	_, err = conn.Exec(ctx, `DELETE FROM notifications WHERE id = $1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete error", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "notification deleted successfully", "id": id})
}
