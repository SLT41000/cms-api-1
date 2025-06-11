package model

import "time"

type OutputCase struct {
	Id                    string     `json:"id"`
	CaseId                *string    `json:"caseId"`
	Casetype_code         *string    `json:"CasetypeCode"`
	Priority              *int       `json:"priority"`
	Phone_number          *string    `json:"phoneNumber"`
	Case_status_code      *string    `json:"caseStatusCode"`
	Case_status_name      *string    `json:"caseStatusName"`
	Case_detail           *string    `json:"caseDetail"`
	Police_station_name   *string    `json:"stationName"`
	Case_location_address *string    `json:"caseLocationAddress"`
	Case_location_detail  *string    `json:"caseLocationDetail"`
	Special_emergency     *int       `json:"specialEmergency"`
	Urgent_amount         *string    `json:"urgentAmount"`
	Created_date          *time.Time `json:"createdDate"`
	Casetype_name         *string    `json:"casetypeName"`
	Media_type            *int       `json:"mediaType"`
	VOwner                *int       `json:"vOwner"`
	VVin                  *string    `json:"vVin"`
}

type CaseListData struct {
	Draw            int          `json:"draw"`
	RecordsTotal    int          `json:"recordsTotal"`
	RecordsFiltered int          `json:"recordsFiltered"`
	Data            []OutputCase `json:"data"`
	Error           string       `json:"error"`
}

type CaseDetailData struct {
	ID                  int        `json:"id"`
	CaseID              string     `json:"caseId"`
	ReferCaseID         *string    `json:"referCaseId"`
	CasetypeCode        string     `json:"casetypeCode"`
	Priority            int        `json:"priority"`
	Ways                int        `json:"ways"`
	PhoneNumber         string     `json:"phoneNumber"`
	CaseStatusCode      string     `json:"caseStatusCode"`
	CaseStatusName      string     `json:"caseStatusName"`
	CaseDetail          string     `json:"caseDetail"`
	DepartmentName      *string    `json:"departmentName"`
	CommandCode         *string    `json:"commandCode"`
	CommandName         *string    `json:"commandName"`
	PoliceStationCode   *string    `json:"policeStationCode"`
	PoliceStationName   *string    `json:"policeStationName"`
	CaseLocationAddress *string    `json:"caseLocationAddress"`
	CaseLocationDetail  *string    `json:"caseLocationDetail"`
	CaseLat             *string    `json:"caseLat"`
	CaseLon             *string    `json:"caseLon"`
	TransImg            string     `json:"transImg"` // Was a string in your example, not an array
	CitizenCode         string     `json:"citizenCode"`
	ExtensionReceive    string     `json:"extensionReceive"`
	SpecialEmergency    int        `json:"specialEmergency"`
	UrgentAmount        *string    `json:"urgentAmount"`
	OpenedDate          time.Time  `json:"openedDate"`
	SavedDate           *time.Time `json:"savedDate"`
	CreatedDate         time.Time  `json:"createdDate"`
	StartedDate         time.Time  `json:"startedDate"`
	ModifiedDate        time.Time  `json:"modifiedDate"`
	UserCreate          string     `json:"userCreate"`
	UserCreateID        *string    `json:"userCreateId"`
	UserModify          string     `json:"userModify"`
	Responsible         *string    `json:"responsible"`
	ApprovedStatus      *int       `json:"approvedStatus"`
	CasetypeName        string     `json:"casetypeName"`
	MediaType           int        `json:"mediaType"`
	VOwner              int        `json:"vOwner"`
	VVin                string     `json:"vVin"`
	DestLocationAddress *string    `json:"destLocationAddress"`
	DestLocationDetail  *string    `json:"destLocationDetail"`
	DestLat             *string    `json:"destLat"`
	DestLon             *string    `json:"destLon"`
	CitizenFullname     string     `json:"citizenFullname"`
}

type CaseDetailResponse struct {
	Status string          `json:"status"`
	Msg    string          `json:"msg"`
	Data   *CaseDetailData `json:"data,omitempty"`
	Desc   string          `json:"desc"`
}

type CaseListResponse struct {
	Status string       `json:"status"`
	Msg    string       `json:"msg"`
	Data   CaseListData `json:"data,omitempty"`
	Desc   string       `json:"desc"`
}

type CaseForCreate struct {
	ReferCaseID         string    `json:"referCaseId" example:"1"`
	CasetypeCode        string    `json:"casetypeCode" example:"1"`
	Priority            int       `json:"priority" example:"0"`
	Ways                int       `json:"ways" example:"0"`
	PhoneNumber         string    `json:"phoneNumber" example:"1"`
	PhoneNumberHide     int       `json:"phoneNumberHide" example:"0"`
	Duration            int       `json:"duration" example:"0"`
	CaseStatusCode      string    `json:"caseStatusCode" example:"1"`
	CaseCondition       int       `json:"caseCondition" example:"0"`
	CaseDetail          string    `json:"caseDetail" example:"1"`
	CaseLocationType    string    `json:"caseLocationType" example:"1"`
	CommandCode         string    `json:"commandCode" example:"1"`
	PoliceStationCode   string    `json:"policeStationCode" example:"1"`
	CaseLocationAddress string    `json:"caseLocationAddress" example:"1"`
	CaseLocationDetail  string    `json:"caseLocationDetail" example:"1"`
	CaseRoute           string    `json:"caseRoute" example:"1"`
	CaseLat             string    `json:"caseLat" example:"1"`
	CaseLon             string    `json:"caseLon" example:"1"`
	CaseDirection       string    `json:"caseDirection" example:"1"`
	CasePhoto           []Media   `json:"casePhoto" `
	TransImg            []Media   `json:"transImg" `
	Home                int       `json:"home" example:"0"`
	CitizenCode         string    `json:"citizenCode" example:"1"`
	ExtensionReceive    string    `json:"extensionReceive" example:"1"`
	CaseSLA             int       `json:"caseSla" example:"0"`
	ActionProCode       string    `json:"actionProCode" example:"1"`
	SpecialEmergency    int       `json:"specialEmergency" example:"1"`
	MediaCode           string    `json:"mediaCode" example:"1"`
	MediaType           int       `json:"mediaType" example:"0"`
	OpenedDate          time.Time `json:"openedDate"`
	CreatedDate         time.Time `json:"createdDate"`
	StartedDate         time.Time `json:"startedDate"`
	ClosedDate          time.Time `json:"closedDate"`
	ModifiedDate        time.Time `json:"modifiedDate" `
	UserCreate          string    `json:"userCreate" example:"1"`
	UserClose           string    `json:"userClose" example:"1"`
	UserModify          string    `json:"userModify" example:"1"`
	NeedAmbulance       int       `json:"needAmbulance" example:"0"`
	Backdated           int       `json:"backdated" example:"0"`
	EscapeRoute         string    `json:"escapeRoute" example:"1"`
	VOwner              int       `json:"vOwner" example:"0"`
	VVin                string    `json:"vVin" example:"1"`
	DestLocationAddress string    `json:"destLocationAddress" example:"1"`
	DestLocationDetail  string    `json:"destLocationDetail" example:"1"`
	DestLat             string    `json:"destLat" example:"1"`
	DestLon             string    `json:"destLon" example:"1"`
	Token               string    `json:"token" example:"1"`
}

func (c *CaseForCreate) SetDefaults() {
	now := time.Now()
	if c.OpenedDate.IsZero() {
		c.OpenedDate = now
	}
	if c.CreatedDate.IsZero() {
		c.CreatedDate = now
	}
	if c.StartedDate.IsZero() {
		c.StartedDate = now
	}
	if c.ClosedDate.IsZero() {
		c.ClosedDate = now
	}
	if c.ModifiedDate.IsZero() {
		c.ModifiedDate = now
	}
}

type CaseForUpdate struct {
	ReferCaseID         string    `json:"referCaseId"`
	CasetypeCode        string    `json:"casetypeCode"`
	Priority            int       `json:"priority"`
	Ways                int       `json:"ways"`
	PhoneNumber         string    `json:"phoneNumber"`
	PhoneNumberHide     int       `json:"phoneNumberHide"`
	Duration            int       `json:"duration"`
	CaseStatusCode      string    `json:"caseStatusCode"`
	CaseCondition       int       `json:"caseCondition"`
	CaseDetail          string    `json:"caseDetail"`
	CaseLocationType    string    `json:"caseLocationType"`
	CommandCode         string    `json:"commandCode"`
	PoliceStationCode   string    `json:"policeStationCode"`
	CaseLocationAddress string    `json:"caseLocationAddress"`
	CaseLocationDetail  string    `json:"caseLocationDetail"`
	CaseRoute           string    `json:"caseRoute"`
	CaseLat             string    `json:"caseLat"`
	CaseLon             string    `json:"caseLon"`
	CaseDirection       string    `json:"caseDirection"`
	CasePhoto           []Media   `json:"casePhoto"`
	TransImg            []Media   `json:"transImg"`
	Home                int       `json:"home"`
	CitizenCode         string    `json:"citizenCode"`
	ExtensionReceive    string    `json:"extensionReceive"`
	CaseSLA             int       `json:"caseSla"`
	ActionProCode       string    `json:"actionProCode"`
	SpecialEmergency    int       `json:"specialEmergency"`
	MediaCode           string    `json:"mediaCode"`
	MediaType           int       `json:"mediaType"`
	OpenedDate          time.Time `json:"openedDate"`
	CreatedDate         time.Time `json:"createdDate"`
	StartedDate         time.Time `json:"startedDate"`
	ClosedDate          time.Time `json:"closedDate"`
	ModifiedDate        time.Time `json:"modifiedDate"`
	UserCreate          string    `json:"userCreate"`
	UserClose           string    `json:"userClose"`
	UserModify          string    `json:"userModify"`
	NeedAmbulance       int       `json:"needAmbulance"`
	Backdated           int       `json:"backdated"`
	EscapeRoute         string    `json:"escapeRoute"`
	VOwner              int       `json:"vOwner"`
	VVin                string    `json:"vVin"`
	DestLocationAddress string    `json:"destLocationAddress"`
	DestLocationDetail  string    `json:"destLocationDetail"`
	DestLat             string    `json:"destLat"`
	DestLon             string    `json:"destLon"`
}
type Media struct {
	URL string `json:"url"`
}

type CaseCloseInput struct {
	CaseStatusCode string    `json:"caseStatusCode"`
	ResultCode     string    `json:"resultCode"`
	ResultDetail   string    `json:"resultDetail"`
	TransImg       []Media   `json:"transImg"`
	ClosedDate     time.Time `json:"closedDate"`
	ModifiedDate   time.Time `json:"modifiedDate"`
	UserClose      string    `json:"userClose"`
	UserModify     string    `json:"userModify"`
}

type CreateCaseResponse struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
	Desc   string `json:"desc"`
	ID     string `json:"id"`
	CaseID string `json:"caseId"`
}

type UpdateCaseResponse struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
	Desc   string `json:"desc"`
	ID     int    `json:"id"`
}

type DeleteCaseResponse struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
	Desc   string `json:"desc"`
}
