package model

import (
	"time"
)

type DeviceIoT struct {
	OrgID       string    `json:"orgId"`
	DeviceID    string    `json:"deviceId"`
	DeviceType  string    `json:"deviceType"`
	Model       string    `json:"model"`
	FirmwareVer string    `json:"firmwareVer"`
	Latitude    string    `json:"latitude"`
	Longitude   string    `json:"longitude"`
	IPAddress   string    `json:"ipAddress"`
	MacAddress  string    `json:"macAddress"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	CreatedBy   string    `json:"createdBy"`
	UpdatedBy   string    `json:"updatedBy"`
}
