package model

import "time"

type AuditLog struct {
	ID int `json:"id" db:"id"`
	// OrgID     string    `json:"orgId" db:"orgId"`
	Username string `json:"username" db:"username"`
	// TxID      string    `json:"txId" db:"txId"`
	UniqueId  string    `json:"uniqueId" db:"uniqueId"`
	MainFunc  string    `json:"mainFunc" db:"mainFunc"`
	SubFunc   string    `json:"subFunc" db:"subFunc"`
	NameFunc  string    `json:"nameFunc" db:"nameFunc"`
	Action    string    `json:"action" db:"action"`
	Status    int       `json:"status" db:"status"`
	Duration  float64   `json:"duration" db:"duration"`
	NewData   string    `json:"newData" db:"newData"`
	OldData   string    `json:"oldData" db:"oldData"`
	ResData   string    `json:"resData" db:"resData"`
	Message   string    `json:"message" db:"message"`
	CreatedAt time.Time `json:"createdAt" db:"createdAt"`
}
