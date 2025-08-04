package model

import (
	"time"
)

type Customer struct {
	ID          string                 `json:"id"`
	OrgID       string                 `json:"orgId"`
	DisplayName string                 `json:"displayName"`
	Title       string                 `json:"title"`
	FirstName   string                 `json:"firstName"`
	MiddleName  string                 `json:"middleName"`
	LastName    string                 `json:"lastName"`
	CitizenID   string                 `json:"citizenId"`
	DOB         time.Time              `json:"dob"`
	Blood       string                 `json:"blood"`
	Gender      string                 `json:"gender"`
	MobileNo    string                 `json:"mobileNo"`
	Address     map[string]interface{} `json:"address"`
	Photo       string                 `json:"photo"`
	Email       string                 `json:"email"`
	UserType    string                 `json:"userType"`
	Active      bool                   `json:"active"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
	CreatedBy   string                 `json:"createdBy"`
	UpdatedBy   string                 `json:"updatedBy"`
}

type CustomerInsert struct {
	DisplayName string                 `json:"displayName"`
	Title       string                 `json:"title"`
	FirstName   string                 `json:"firstName"`
	MiddleName  string                 `json:"middleName"`
	LastName    string                 `json:"lastName"`
	CitizenID   string                 `json:"citizenId"`
	DOB         time.Time              `json:"dob"`
	Blood       string                 `json:"blood"`
	Gender      string                 `json:"gender"`
	MobileNo    string                 `json:"mobileNo"`
	Address     map[string]interface{} `json:"address"`
	Photo       string                 `json:"photo"`
	Email       string                 `json:"email"`
	UserType    string                 `json:"userType"`
	Active      bool                   `json:"active"`
}

type CustomerUpdate struct {
	DisplayName string                 `json:"displayName"`
	Title       string                 `json:"title"`
	FirstName   string                 `json:"firstName"`
	MiddleName  string                 `json:"middleName"`
	LastName    string                 `json:"lastName"`
	CitizenID   string                 `json:"citizenId"`
	DOB         time.Time              `json:"dob"`
	Blood       string                 `json:"blood"`
	Gender      string                 `json:"gender"`
	MobileNo    string                 `json:"mobileNo"`
	Address     map[string]interface{} `json:"address"`
	Photo       string                 `json:"photo"`
	Email       string                 `json:"email"`
	UserType    string                 `json:"userType"`
	Active      bool                   `json:"active"`
}

type CustomerSocial struct {
	ID         string    `json:"id"`
	OrgID      string    `json:"orgId"`
	CustID     string    `json:"custId"`
	SocialType string    `json:"socialType"`
	SocialID   string    `json:"socialId"`
	SocialName string    `json:"socialName"`
	ImgURL     string    `json:"imgUrl"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	CreatedBy  string    `json:"createdBy"`
	UpdatedBy  string    `json:"updatedBy"`
}

type CustomerSocialInsert struct {
	CustID     string `json:"custId"`
	SocialType string `json:"socialType"`
	SocialID   string `json:"socialId"`
	SocialName string `json:"socialName"`
	ImgURL     string `json:"imgUrl"`
}

type CustomerSocialUpdate struct {
	CustID     string `json:"custId"`
	SocialType string `json:"socialType"`
	SocialID   string `json:"socialId"`
	SocialName string `json:"socialName"`
	ImgURL     string `json:"imgUrl"`
}

type CustomerContact struct {
	ID           string                 `json:"id"`
	OrgID        string                 `json:"orgId"`
	CustID       int                    `json:"custId"`
	ContactName  string                 `json:"contactName"`
	ContactPhone string                 `json:"contactPhone"`
	ContactAddr  map[string]interface{} `json:"contactAddr"`
	CreatedAt    time.Time              `json:"createdAt"`
	UpdatedAt    time.Time              `json:"updatedAt"`
	CreatedBy    string                 `json:"createdBy"`
	UpdatedBy    string                 `json:"updatedBy"`
}

type CustomerContactInsert struct {
	CustID       int                    `json:"custId"`
	ContactName  string                 `json:"contactName"`
	ContactPhone string                 `json:"contactPhone"`
	ContactAddr  map[string]interface{} `json:"contactAddr"`
}

type CustomerContactUpdate struct {
	CustID       int                    `json:"custId"`
	ContactName  string                 `json:"contactName"`
	ContactPhone string                 `json:"contactPhone"`
	ContactAddr  map[string]interface{} `json:"contactAddr"`
}
