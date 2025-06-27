package model

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type CaseTransactionResponse struct {
	Status string                `json:"status"`
	Msg    string                `json:"msg"`
	Data   []CaseTransactionData `json:"data,omitempty"`
	Desc   string                `json:"desc"`
}

type CaseTransactionData struct {
	ID             int         `json:"id"`
	CaseID         string      `json:"caseId"`
	UserCode       *string     `json:"userCode"`
	UserName       string      `json:"userName"`
	ReceiveDate    *time.Time  `json:"receiveDate"`
	ArriveDate     *time.Time  `json:"arriveDate"`
	CloseDate      *time.Time  `json:"closeDate"`
	CancelDate     *time.Time  `json:"cancelDate"`
	Duration       pgtype.Int8 `json:"duration" swaggerignore:"true"`
	SuggestRoute   *string     `json:"suggestRoute"`
	UserSLA        pgtype.Int8 `json:"userSla" swaggerignore:"true"`
	Viewed         pgtype.Int8 `json:"viewed" swaggerignore:"true"`
	CaseStatusCode string      `json:"caseStatusCode"`
	NotiStage      *string     `json:"notiStage"`
	UserClosedJob  *string     `json:"userClosedJob"`
	ResultCode     *string     `json:"resultCode"`
	ResultDetail   *string     `json:"resultDetail"`
	CreatedDate    time.Time   `json:"createdDate"`
	CreatedModify  time.Time   `json:"createdModify"`
	Owner          string      `json:"owner"`
	UpdatedAccount string      `json:"updatedAccount"`
	VehicleCode    *string     `json:"vehicleCode"`
	ActionCarType  *string     `json:"actionCarType"`
	TimeToArrive   *string     `json:"timeToArrive"`
	Lat            *string     `json:"lat"`
	Lon            *string     `json:"lon"`
}

type CaseNote struct {
	ID           pgtype.Int8 `json:"id"`
	CaseID       string      `json:"caseId"`
	Detail       string      `json:"detail"`
	CreatedDate  time.Time   `json:"createdDate"`
	ModifiedDate time.Time   `json:"modifiedDate"`
	UserCreate   string      `json:"userCreate"`
	UserModify   string      `json:"userModify"`
}

type CaseNoteResponse struct {
	Status string     `json:"status"`
	Msg    string     `json:"msg"`
	Data   []CaseNote `json:"data,omitempty"`
	Desc   string     `json:"desc"`
}

type CaseTransactionModelInput struct {
	CaseID         string     `json:"caseId"`
	UserCode       string     `json:"userCode"`
	UserName       string     `json:"userName"`
	ReceiveDate    *time.Time `json:"receiveDate"`
	ArriveDate     *time.Time `json:"arriveDate"`
	CloseDate      *time.Time `json:"closeDate"`
	CancelDate     *time.Time `json:"cancelDate"`
	Duration       int        `json:"duration"`
	SuggestRoute   string     `json:"suggestRoute"`
	UserSLA        int        `json:"userSla"`
	Viewed         int        `json:"viewed"`
	CaseStatusCode string     `json:"caseStatusCode"`
	NotiStage      string     `json:"notiStage"`
	UserClosedJob  string     `json:"userClosedJob"`
	ResultCode     string     `json:"resultCode"`
	ResultDetail   string     `json:"resultDetail"`
	CreatedDate    time.Time  `json:"createdDate"`
	CreatedModify  time.Time  `json:"createdModify"`
	Owner          string     `json:"owner"`
	UpdatedAccount string     `json:"updatedAccount"`
	VehicleCode    string     `json:"vehicleCode"`
	ActionCarType  string     `json:"actionCarType"`
	TimeToArrive   string     `json:"timeToArrive"` // If it’s a string, keep as-is; else change to time.Duration or time.Time
	Lat            string     `json:"lat"`          // Consider float64 if GPS
	Lon            string     `json:"lon"`
	CommandedDate  *time.Time `json:"commandedDate"`
	ReceivedDate   *time.Time `json:"receivedDate"`
	ArrivedDate    *time.Time `json:"arrivedDate"`
	ClosedDate     *time.Time `json:"closedDate"`
	ModifiedDate   *time.Time `json:"modifiedDate"`
	UserCommand    string     `json:"userCommand"`
	UserReceive    string     `json:"userReceive"`
	UserArrive     string     `json:"userArrive"`
	UserClose      string     `json:"userClose"`
	UserModify     string     `json:"userModify"`
	Token          string     `json:"token"`
}

type CaseTransactionCRUDResponse struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
	Desc   string `json:"desc"`
	ID     int    `json:"id,omitempty"`
	CaseID string `json:"caseId,omitempty"`
}

type CaseNoteInput struct {
	CaseID       string    `json:"caseId"`
	Detail       string    `json:"detail"`
	CreatedDate  time.Time `json:"createdDate"`
	ModifiedDate time.Time `json:"modifiedDate"`
	UserCreate   string    `json:"userCreate"`
	UserModify   string    `json:"userModify"`
}

type CaseNoteInputResponse struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
	Desc   string `json:"desc"`
	ID     int    `json:"id,omitempty"`
	CaseID string `json:"caseId,omitempty"`
}

type CaseTransactionUpdateInput struct {
	CaseID         string     `json:"caseId"`
	UserCode       string     `json:"userCode"`
	UserName       string     `json:"userName"`
	ReceiveDate    *time.Time `json:"receiveDate"`
	ArriveDate     *time.Time `json:"arriveDate"`
	CloseDate      *time.Time `json:"closeDate"`
	CancelDate     *time.Time `json:"cancelDate"`
	Duration       int        `json:"duration"`
	SuggestRoute   string     `json:"suggestRoute"`
	UserSLA        int        `json:"userSla"`
	Viewed         int        `json:"viewed"`
	CaseStatusCode string     `json:"caseStatusCode"`
	NotiStage      string     `json:"notiStage"`
	UserClosedJob  string     `json:"userClosedJob"`
	ResultCode     string     `json:"resultCode"`
	ResultDetail   string     `json:"resultDetail"`
	CreatedDate    time.Time  `json:"createdDate"`
	CreatedModify  time.Time  `json:"createdModify"`
	Owner          string     `json:"owner"`
	UpdatedAccount string     `json:"updatedAccount"`
	VehicleCode    string     `json:"vehicleCode"`
	ActionCarType  string     `json:"actionCarType"`
	TimeToArrive   string     `json:"timeToArrive"` // If it’s a string, keep as-is; else change to time.Duration or time.Time
	Lat            string     `json:"lat"`          // Consider float64 if GPS
	Lon            string     `json:"lon"`
	CommandedDate  *time.Time `json:"commandedDate"`
	ReceivedDate   *time.Time `json:"receivedDate"`
	ArrivedDate    *time.Time `json:"arrivedDate"`
	ClosedDate     *time.Time `json:"closedDate"`
	ModifiedDate   *time.Time `json:"modifiedDate"`
	UserCommand    string     `json:"userCommand"`
	UserReceive    string     `json:"userReceive"`
	UserArrive     string     `json:"userArrive"`
	UserClose      string     `json:"userClose"`
	UserModify     string     `json:"userModify"`
	Token          string     `json:"token"`
}
