package model

import "time"

type RolePermission struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	OrgID     string    `json:"orgId" gorm:"column:orgId"`
	RoleID    string    `json:"roleId" gorm:"column:roleId"`
	PermID    string    `json:"permId" gorm:"column:permId"`
	Active    bool      `json:"active" gorm:"column:active"`
	CreatedAt time.Time `json:"createdAt" gorm:"column:createdAt"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"column:updatedAt"`
	CreatedBy string    `json:"createdBy" gorm:"column:createdBy"`
	UpdatedBy string    `json:"updatedBy" gorm:"column:updatedBy"`
}

type RolePermissionBody struct {
	PermID string `json:"permId" gorm:"column:permId"`
	Active bool   `json:"active" gorm:"column:active"`
}

type RolePermissionInsert struct {
	RoleID string               `json:"roleId" gorm:"column:roleId"`
	PermID []RolePermissionBody `json:"permId" gorm:"column:permId"`
}

type RolePermissionUpdate struct {
	PermID []RolePermissionBody `json:"permId" gorm:"column:permId"`
}
