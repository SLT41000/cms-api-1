package model

type GenerateCaseIDRequest struct {
	Prefix string `json:"prefix" binding:"required"`
}
