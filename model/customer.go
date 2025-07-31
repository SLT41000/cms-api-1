package model

import (
	"encoding/json"
	"time"
)

type Customer struct {
	ID          string          `json:"id"`
	OrgID       string          `json:"orgId"`
	DisplayName string          `json:"displayName"`
	Title       string          `json:"title"`
	FirstName   string          `json:"firstName"`
	MiddleName  string          `json:"middleName"`
	LastName    string          `json:"lastName"`
	CitizenID   string          `json:"citizenId"`
	DOB         time.Time       `json:"dob"`
	Blood       string          `json:"blood"`
	Gender      string          `json:"gender"`
	MobileNo    string          `json:"mobileNo"`
	Address     json.RawMessage `json:"address"`
	Photo       string          `json:"photo"`
	Email       string          `json:"email"`
	UserType    string          `json:"userType"`
	Active      bool            `json:"active"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
	CreatedBy   string          `json:"createdBy"`
	UpdatedBy   string          `json:"updatedBy"`
}

type CustomerInsert struct {
	DisplayName string          `json:"displayName"`
	Title       string          `json:"title"`
	FirstName   string          `json:"firstName"`
	MiddleName  string          `json:"middleName"`
	LastName    string          `json:"lastName"`
	CitizenID   string          `json:"citizenId"`
	DOB         time.Time       `json:"dob"`
	Blood       string          `json:"blood"`
	Gender      string          `json:"gender"`
	MobileNo    string          `json:"mobileNo"`
	Address     json.RawMessage `json:"address"`
	Photo       string          `json:"photo"`
	Email       string          `json:"email"`
	UserType    string          `json:"userType"`
	Active      bool            `json:"active"`
}

type CustomerUpdate struct {
	DisplayName string          `json:"displayName"`
	Title       string          `json:"title"`
	FirstName   string          `json:"firstName"`
	MiddleName  string          `json:"middleName"`
	LastName    string          `json:"lastName"`
	CitizenID   string          `json:"citizenId"`
	DOB         time.Time       `json:"dob"`
	Blood       string          `json:"blood"`
	Gender      string          `json:"gender"`
	MobileNo    string          `json:"mobileNo"`
	Address     json.RawMessage `json:"address"`
	Photo       string          `json:"photo"`
	Email       string          `json:"email"`
	UserType    string          `json:"userType"`
	Active      bool            `json:"active"`
}
