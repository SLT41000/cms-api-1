package model

import (
	"time"
)

type CaseHistory struct {
	ID        int       `json:"id" db:"id"`
	OrgID     string    `json:"orgId" db:"orgId"`
	CaseID    string    `json:"caseId" db:"caseId"`
	Username  string    `json:"username" db:"username"`
	Type      string    `json:"type" db:"type"`
	FullMsg   string    `json:"fullMsg" db:"fullMsg"`
	JSONData  *string   `json:"jsonData" db:"jsonData"` // assuming jsonb in DB
	CreatedAt time.Time `json:"createdAt" db:"createdAt"`
	CreatedBy string    `json:"createdBy" db:"createdBy"`
}

type CaseHistoryInsert struct {
	CaseID   string `json:"caseId" db:"caseId"`
	Username string `json:"username" db:"username"`
	Type     string `json:"type" db:"type"`
	FullMsg  string `json:"fullMsg" db:"fullMsg"`
	JSONData string `json:"jsonData" db:"jsonData"`
}

type CaseHistoryUpdate struct {
	// CaseID   string          `json:"caseId" db:"caseId"`
	// Username string          `json:"username" db:"username"`
	Type     string `json:"type" db:"type"`
	FullMsg  string `json:"fullMsg" db:"fullMsg"`
	JSONData string `json:"jsonData" db:"jsonData"`
}
