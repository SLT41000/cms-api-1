package model

// type MinimalCaseInsert struct {
// 	CaseId string `json:"caseId" `
// 	//CaseVersion     *string    `json:"caseVersion"`
// 	//ReferCaseID     *string    `json:"referCaseId"`
// 	CaseTypeID  string `json:"caseTypeId"`
// 	CaseSTypeID string `json:"caseSTypeId"`
// 	//Priority        int        `json:"priority"`
// 	WfID string `json:"wfId"`
// 	//WfVersions      *string    `json:"versions"`
// 	NodeID   string  `json:"nodeId" `
// 	Source   string  `json:"source"`
// 	DeviceID *string `json:"deviceId"`
// 	PhoneNo  *string `json:"phoneNo"`
// 	//PhoneNoHide     bool       `json:"phoneNoHide"`
// 	CaseDetail *string `json:"caseDetail"`
// 	//ExtReceive      *string    `json:"extReceive"`
// 	StatusID string `json:"statusId"`
// 	//CaseLat         *string    `json:"caseLat"`
// 	//CaseLon         *string    `json:"caseLon"`
// 	//CaseLocAddr     *string    `json:"caselocAddr"`
// 	//CaseLocAddrDecs *string    `json:"caselocAddrDecs"`
// 	CountryID string `json:"countryId"`
// 	ProvID    string `json:"provId"`
// 	DistID    string `json:"distId"`
// 	//CaseDuration    int        `json:"caseDuration"`
// 	//CreatedDate     *time.Time `json:"createdDate"`
// 	//StartedDate     *time.Time `json:"startedDate"`
// 	//CommandedDate   *time.Time `json:"commandedDate"`
// 	//ReceivedDate    *time.Time `json:"receivedDate"`
// 	//ArrivedDate     *time.Time `json:"arrivedDate"`
// 	//ClosedDate      *time.Time `json:"closedDate"`
// 	//UserCreate      *string    `json:"usercreate"`
// 	//UserCommand     *string    `json:"usercommand"`
// 	//UserReceive     *string    `json:"userreceive"`
// 	//UserArrive      *string    `json:"userarrive"`
// 	//UserClose       *string    `json:"userclose"`
// 	//ResID           *string    `json:"resId"`
// 	//ResDetail       *string    `json:"resDetail"`
// 	//ScheduleFlag    *bool      `json:"scheduleFlag"`
// 	//ScheduleDate    *time.Time `json:"scheduleDate"`
// }

type IotInfo struct {
	DeviceID    string  `json:"deviceId"`
	DeviceType  string  `json:"deviceType"`
	Model       *string `json:"model"`
	FirmwareVer *string `json:"firmwareVer"`
	Latitude    *string `json:"latitude"`
	Longitude   *string `json:"longitude"`
	IPAddress   *string `json:"ipAddress"`
	MacAddress  *string `json:"macAddress"`
	CountryID   string  `json:"countryId"`
	ProvID      string  `json:"provId"`
	DistID      string  `json:"distId"`
}

type MinimalCaseInsert struct {
	CaseId      string   `json:"caseId"`
	CaseTypeID  string   `json:"caseTypeId"`
	CaseSTypeID string   `json:"caseSTypeId"`
	WfID        string   `json:"wfId"`
	NodeID      string   `json:"nodeId"`
	Source      string   `json:"source"`
	PhoneNo     *string  `json:"phoneNo"`
	CaseDetail  *string  `json:"caseDetail"`
	StatusID    string   `json:"statusId"`
	IotInfo     *IotInfo `json:"iotInfo"`
}
