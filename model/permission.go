package model

import "time"

type Permission struct {
	ID        string    `json:"id"`
	GroupName string    `json:"groupName"`
	PermID    string    `json:"permId"`
	PermName  string    `json:"permName"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	CreatedBy string    `json:"createdBy"`
	UpdatedBy string    `json:"updatedBy"`
}

type PermissionInsert struct {
	GroupName string `json:"groupName"`
	PermName  string `json:"permName"`
	Active    bool   `json:"active"`
}

type PermissionUpdate struct {
	GroupName string `json:"groupName"`
	PermName  string `json:"permName"`
	Active    bool   `json:"active"`
}
