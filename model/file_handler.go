package model

import "time"

type DeleteFileRequest struct {
	Path     string `uri:"path" binding:"required" example:"profile"`
	Filename string `form:"filename" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000.jpg"`
	CaseId   string `form:"caseId" example:"CASE-000123"`
	AttId    string `form:"attId" example:"123e4567-e89b-12d3-a456-426614174000"`
}

type TixCaseAttachment struct {
	Id        int       `json:"id"`
	OrgId     string    `json:"orgId"`
	CaseId    string    `json:"caseId"`
	Type      string    `json:"type"`
	AttId     string    `json:"attId"`
	AttName   string    `json:"attName"`
	AttUrl    string    `json:"attUrl"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
	CreatedBy string    `json:"createdBy"`
	UpdatedBy string    `json:"updatedBy"`
}
