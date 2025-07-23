package model

import "time"

type RolePermission struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	OrgID     int64     `json:"orgId" gorm:"column:orgId"`
	RoleID    int64     `json:"roleId" gorm:"column:roleId"`
	PermID    int64     `json:"permId" gorm:"column:permId"`
	Active    bool      `json:"active" gorm:"column:active"`
	CreatedAt time.Time `json:"createdAt" gorm:"column:createdAt"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"column:updatedAt"`
	CreatedBy int64     `json:"createdBy" gorm:"column:createdBy"`
	UpdatedBy int64     `json:"updatedBy" gorm:"column:updatedBy"`
}

type RolePermissionBody struct {
	PermID string `json:"permId" gorm:"column:permId"`
	Active bool   `json:"active" gorm:"column:active"`
}

type RolePermissionInsert struct {
	PermID []RolePermissionBody `json:"permId" gorm:"column:permId"`
}

type RolePermissionUpdate struct {
	PermID []RolePermissionBody `json:"permId" gorm:"column:permId"`
}
