package model

import (
	"encoding/json"
	"time"
)

type User struct {
	ID                    string          `json:"id"`
	OrgID                 string          `json:"orgId"`
	DisplayName           string          `json:"displayName"`
	Title                 string          `json:"title"`
	FirstName             string          `json:"firstName"`
	MiddleName            string          `json:"middleName"`
	LastName              string          `json:"lastName"`
	CitizenID             string          `json:"citizenId"`
	Bod                   time.Time       `json:"bod"`
	Blood                 string          `json:"blood"`
	Gender                string          `json:"gender"`
	MobileNo              string          `json:"mobileNo"`
	Address               json.RawMessage `json:"address"` // Assumes JSON column
	Photo                 *string         `json:"photo"`
	Username              string          `json:"username"`
	Password              string          `json:"password"`
	Email                 string          `json:"email"`
	RoleID                string          `json:"roleId"`
	UserType              string          `json:"userType"`
	EmpID                 string          `json:"empId"`
	DeptID                string          `json:"deptId"`
	CommID                string          `json:"commId"`
	StnID                 string          `json:"stnId"`
	Active                bool            `json:"active"`
	ActivationToken       *string         `json:"activationToken"`
	LastActivationRequest *time.Time      `json:"lastActivationRequest"`
	LostPasswordRequest   *time.Time      `json:"lostPasswordRequest"`
	SignupStamp           *time.Time      `json:"signupStamp"`
	IsLogin               bool            `json:"islogin"`
	LastLogin             *time.Time      `json:"lastLogin"`
	CreatedAt             time.Time       `json:"createdAt"`
	UpdatedAt             time.Time       `json:"updatedAt"`
	CreatedBy             string          `json:"createdBy"`
	UpdatedBy             string          `json:"updatedBy"`
}

type UserAdminInput struct {
	OrgID                 string     `json:"orgId"`
	DisplayName           string     `json:"displayName"`
	Title                 string     `json:"title"`
	FirstName             string     `json:"firstName"`
	MiddleName            string     `json:"middleName"`
	LastName              string     `json:"lastName"`
	CitizenID             string     `json:"citizenId"`
	Bod                   string     `json:"bod"`
	Blood                 string     `json:"blood"`
	Gender                int        `json:"gender"`
	MobileNo              string     `json:"mobileNo"`
	Address               string     `json:"address"` // Assumes JSON column
	Photo                 *string    `json:"photo"`
	Username              string     `json:"username"`
	Password              string     `json:"password"`
	Email                 string     `json:"email"`
	RoleID                string     `json:"roleId"`
	UserType              int        `json:"userType"`
	EmpID                 string     `json:"empId"`
	DeptID                string     `json:"deptId"`
	CommID                string     `json:"commId"`
	StnID                 string     `json:"stnId"`
	Active                bool       `json:"active"`
	ActivationToken       string     `json:"activationToken"`
	LastActivationRequest int        `json:"lastActivationRequest"`
	LostPasswordRequest   int        `json:"lostPasswordRequest"`
	SignupStamp           int        `json:"signupStamp"`
	IsLogin               bool       `json:"islogin"`
	LastLogin             *time.Time `json:"lastLogin"`
}

type RefreshInput struct {
	RefreshToken string `json:"refreshToken"`
}

type Login struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	Organization string `json:"organization"`
}
