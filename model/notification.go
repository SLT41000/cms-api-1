package model

import "time"

type Notification struct {
	ID          string    `json:"id"`
	CaseID      string    `json:"caseId"`
	CaseType    string    `json:"caseType"`
	CaseDetail  string    `json:"caseDetail"`
	Recipient   string    `json:"recipient"`
	Sender      string    `json:"sender"` // Note: 'sender' is intentionally lowercase to avoid conflict with the field in NotificationRecipient
	Message     string    `json:"message"`
	EventType   string    `json:"eventType"`
	CreatedAt   time.Time `json:"createdAt"`
	Read        bool      `json:"read"`
	RedirectURL string    `json:"redirectURL"`
}

// Table: notification_recipients
// type NotificationRecipient struct {
// 	ID                string     `json:"id"`                // รหัสรายการการแจ้งเตือนแต่ละรายการ (PK)
// 	NotificationID    string     `json:"notificationId"`    // อ้างอิงไปยัง Notification ที่ส่ง (FK ไปยัง notifications)
// 	UserID            string     `json:"userId"`            // รหัสของผู้ใช้ที่เป็นเป้าหมายของการแจ้งเตือนนี้
// 	Status            string     `json:"status"`            // สถานะของการแจ้งเตือน: unread, read, acknowledged
// 	AcknowledgedAt    *time.Time `json:"acknowledgedAt"`    // เวลาที่ผู้ใช้ทำการยืนยันรับรู้ (nullable)
// 	DeliveredChannels []string   `json:"deliveredChannels"` // ช่องทางที่ส่งสำเร็จแล้ว เช่น ["email", "web", "line"]
// }

// Table: notification_templates
// type NotificationTemplate struct {
// 	ID              string `json:"id"`              // รหัสของ Template (PK)
// 	OrgID           string `json:"orgId"`           // รหัสองค์กร (รองรับ multi-tenant platform)
// 	EventType       string `json:"eventType"`       // ประเภทของ Event ที่ Template นี้รองรับ เช่น "open_case"
// 	Language        string `json:"language"`        // ภาษา เช่น "en" หรือ "th"
// 	TitleTemplate   string `json:"titleTemplate"`   // Template ของหัวข้อ เช่น "มีเคสใหม่ #{{caseId}}"
// 	MessageTemplate string `json:"messageTemplate"` // Template ของข้อความแบบเต็ม เช่น "เกิดเหตุที่ {{location}}, รหัสเคส: {{caseId}}"
// }

// Table: user_preferences
// type UserPreference struct {
// 	UserID            string   `json:"userId"`            // รหัสผู้ใช้ (FK ไปยัง users)
// 	PreferredChannels []string `json:"preferredChannels"` // ช่องทางที่ผู้ใช้ต้องการรับการแจ้งเตือน เช่น ["web", "email", "line"]
// 	DoNotDisturb      bool     `json:"doNotDisturb"`      // ถ้า true จะระงับการแจ้งเตือนในช่วงที่ตั้งไว้
// }
