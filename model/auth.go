package model

import "time"

type InputTokenModel struct {
	GrantType    *string `json:"grantType"`
	Scope        *string `json:"scope"`
	Username     string  `json:"username"`
	Password     string  `json:"password"`
	ClientId     *string `json:"clientId"`
	ClientSecret *string `json:"clientSecret"`
}

type OutputTokenModel struct {
	AccessToken string `json:"accessToken"`
	TokenType   string `json:"tokenType"`
}

type UserInputModel struct {
	OrgID        string `json:"orgId"`
	UserID       string `json:"userId"`
	DisplayName  string `json:"displayName"`
	FullName     string `json:"fullName"`
	PhoneNumber  string `json:"phoneNumber"`
	Email        string `json:"email"`
	Username     string `json:"username"`
	PasswordHash string `json:"passwordHash"`
	LastLogin    string `json:"lastLogin"`
	RoleID       string `json:"roleId"`
	Active       bool   `json:"active"`
	AreaID       string `json:"areaId"`
	DeviceID     string `json:"deviceId"`
	PushToken    string `json:"pushToken"`
	CurrentLat   string `json:"currentLat"`
	CurrentLon   string `json:"currentLon"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
	CreatedBy    string `json:"createdBy"`
	UpdatedBy    string `json:"updatedBy"`
}

type User struct {
	ID           string    `json:"id"` // Primary key
	OrgID        string    `json:"orgId"`
	UserID       string    `json:"userId"`
	DisplayName  string    `json:"displayName"`
	FullName     string    `json:"fullName"`
	PhoneNumber  string    `json:"phoneNumber"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"passwordHash"`
	LastLogin    time.Time `json:"lastLogin"`
	RoleID       string    `json:"roleId"`
	Active       bool      `json:"active"`
	AreaID       string    `json:"areaId"`
	DeviceID     string    `json:"deviceId"`
	PushToken    string    `json:"pushToken"`
	CurrentLat   string    `json:"currentLat"`
	CurrentLon   string    `json:"currentLon"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	CreatedBy    string    `json:"createdBy"`
	UpdatedBy    string    `json:"updatedBy"`
}
