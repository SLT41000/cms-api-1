package model

import (
	"time"
)

type Role struct {
	ID        string    `json:"id" gorm:"primaryKey;autoIncrement"`
	OrgID     string    `json:"orgId" gorm:"column:orgId"`
	RoleName  string    `json:"roleName" gorm:"column:roleName"`
	Active    bool      `json:"active" gorm:"column:active"`
	CreatedAt time.Time `json:"createdAt" gorm:"column:createdAt"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"column:updatedAt"`
	CreatedBy string    `json:"createdBy" gorm:"column:createdBy"`
	UpdatedBy string    `json:"updatedBy" gorm:"column:updatedBy"`
}

type RoleInsert struct {
	RoleName string `json:"roleName" gorm:"column:roleName"`
	Active   bool   `json:"active" gorm:"column:active"`
}

type RoleUpdate struct {
	RoleName string `json:"roleName" gorm:"column:roleName"`
	Active   bool   `json:"active" gorm:"column:active"`
}
