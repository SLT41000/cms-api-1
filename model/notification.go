package model

import (
	"time"

	"github.com/gorilla/websocket"
)

type UserConnectionInfo struct {
	Conn        *websocket.Conn `json:"-"`
	ID          string          `json:"id"`
	EmpID       string          `json:"empId"`
	RoleID      string          `json:"roleId"`
	OrgID       string          `json:"orgID"`
	StnID       string          `json:"stnID"`
	DeptID      string          `json:"deptId"`
	CommID      string          `json:"commId"`
	Username    string          `json:"username"`
	GrpID       []string        `json:"grpId"`
	DistIdLists []string        `json:"distIdLists"`
	Ip          string          `json:"ip"`
}
type RegistrationMessage struct {
	OrgID    string `json:"orgId"`    // องค์กร
	Username string `json:"username"` // ชื่อผู้ใช้
	EmpID    string `json:"empId"`    // รหัสพนักงาน
	RoleID   string `json:"roleId"`   // ตำแหน่ง
	DeptID   string `json:"deptId"`   // แผนก
	CommID   string `json:"commId"`   // หน่วยงานย่อย/สายงาน
	StnID    string `json:"stnId"`    // สถานี/สาขา
	GrpID    string `json:"grpId"`    // กลุ่มผู้ใช้
}

// Notification คือข้อมูลการแจ้งเตือนหลัก
type Notification struct {
	Event       *string      `json:"EVENT"`
	ID          int          `json:"id,omitempty"`
	OrgID       string       `json:"orgId,omitempty"`
	SenderType  string       `json:"senderType,omitempty"`  // "SYSTEM" or "USER"
	SenderPhoto string       `json:"senderPhoto,omitempty"` // เพิ่มใหม่
	Sender      string       `json:"sender,omitempty"`
	Message     string       `json:"message,omitempty"`
	EventType   string       `json:"eventType"`
	RedirectUrl string       `json:"redirectUrl,omitempty"`
	CreatedAt   *time.Time   `json:"createdAt,omitempty"`
	CreatedBy   string       `json:"createdBy,omitempty"` // เพิ่มใหม่
	ExpiredAt   *time.Time   `json:"expiredAt,omitempty"` // เพิ่มใหม่
	Data        *[]Data      `json:"data,omitempty"`
	Recipients  *[]Recipient `json:"recipients,omitempty"` // ใช้ตอนสร้างเท่านั้น
	Additional  interface{}  `json:"additionalJson,omitempty" swaggertype:"object"`
}

// ใช้สำหรับการตอบกลับสถานะ connection
type SubscribeResponse struct {
	EVENT    string `json:"EVENT"`
	Msg      string `json:"msg"`
	OrgId    string `json:"orgId"`
	Username string `json:"username"`
}

// Recipient คือเป้าหมายผู้รับ
type Recipient struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}
type Data struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// NotificationCreateRequest สำหรับ request body ในการสร้าง notification ใหม่
// ไม่รวม ID และ CreatedAt เพราะจะถูกสร้างโดยระบบ
type NotificationCreateRequest struct {
	Event       *string      `json:"EVENT"`
	ID          int          `json:"id,omitempty"`
	OrgID       string       `json:"orgId,omitempty"`
	SenderType  string       `json:"senderType,omitempty"`  // "SYSTEM" or "USER"
	SenderPhoto string       `json:"senderPhoto,omitempty"` // เพิ่มใหม่
	Sender      string       `json:"sender,omitempty"`
	Message     string       `json:"message,omitempty"`
	EventType   string       `json:"eventType"`
	RedirectUrl string       `json:"redirectUrl,omitempty"`
	CreatedAt   *time.Time   `json:"createdAt,omitempty"`
	CreatedBy   string       `json:"createdBy,omitempty"` // เพิ่มใหม่
	ExpiredAt   *time.Time   `json:"expiredAt,omitempty"` // เพิ่มใหม่
	Data        *[]Data      `json:"data,omitempty"`
	Recipients  *[]Recipient `json:"recipients,omitempty"` // ใช้ตอนสร้างเท่านั้น
	Additional  interface{}  `json:"additionalJson,omitempty" swaggertype:"object"`
}

type HiddenNotification struct {
	Event      *string     `json:"EVENT,omitempty"` // optional, original event name
	ID         int         `json:"id,omitempty"`
	EventType  string      `json:"eventType"`      // should be "hidden"
	Additional interface{} `json:"additionalJson"` // dashboard or summary JSON

}
