package model

import (
	"encoding/json"
	"time"
)

type Um_User struct {
	OrgID                 string          `json:"orgId"`
	DisplayName           string          `json:"displayName"`
	Title                 string          `json:"title"`
	FirstName             string          `json:"firstName"`
	MiddleName            string          `json:"middleName"`
	LastName              string          `json:"lastName"`
	CitizenID             string          `json:"citizenId"`
	Bod                   time.Time       `json:"bod"`   // Date of birth
	Blood                 string          `json:"blood"` // Blood type
	Gender                string          `json:"gender"`
	MobileNo              string          `json:"mobileNo"` // Primary mobile
	Address               json.RawMessage `json:"address"`
	Photo                 *string         `json:"photo"` // URL or path
	Username              string          `json:"username"`
	Password              string          `json:"password"` // Store a hash, never plain text
	Email                 string          `json:"email"`
	RoleID                string          `json:"roleId"`
	UserType              string          `json:"userType"`
	EmpID                 string          `json:"empId"` // Employee code / HR number
	DeptID                string          `json:"deptId"`
	CommID                string          `json:"commId"`
	StnID                 string          `json:"stnId"`
	Active                bool            `json:"active"`
	ActivationToken       *string         `json:"activationToken"`
	LastActivationRequest *time.Time      `json:"lastActivationRequest"` // nullable
	LostPasswordRequest   *time.Time      `json:"lostPasswordRequest"`   // nullable
	SignupStamp           *time.Time      `json:"signupStamp"`
	IsLogin               bool            `json:"islogin"`
	Skill                 []string        `json:"skills"`
	Contacts              []string        `json:"contacts"`
	Socials               []string        `json:"socials"`
	LastLogin             *time.Time      `json:"lastLogin"` // nullable
	CreatedAt             time.Time       `json:"createdAt"`
	UpdatedAt             time.Time       `json:"updatedAt"`
	CreatedBy             string          `json:"createdBy"`
	UpdatedBy             string          `json:"updatedBy"`
}
