package model

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type WorkOrder struct {
	WorkOrderNumber       string        `json:"work_order_number"`
	ParentWorkOrderNumber string        `json:"parent_work_order_number"`
	WorkOrderRefNumber    string        `json:"work_order_ref_number"`
	IncidentNumber        string        `json:"incident_number"`
	WorkOrderType         string        `json:"work_order_type"`
	WorkOrderMetadata     WorkOrderMeta `json:"work_order_metadata"`
	UserMetadata          UserMeta      `json:"user_metadata"`
	DeviceMetadata        DeviceMeta    `json:"device_metadata"`
	SopMetadata           SopMeta       `json:"sop_metadata"`
	Status                string        `json:"status"`
	WorkDate              string        `json:"work_date"`
	Workspace             string        `json:"workspace"`
	Namespace             string        `json:"namespace"`
	Source                string        `json:"source"`
}

type WorkOrderMeta struct {
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Severity    string      `json:"severity"`
	Location    GeoLocation `json:"location"`
	Images      []string    `json:"images"`
}

type UserMeta struct {
	AssignedEmployeeCode  AssignedEmployee `json:"assigned_employee_code"`
	AssociateEmployeeCode []string         `json:"associate_employee_code"`
}

type DeviceMeta struct {
	DeviceID           string      `json:"device_id"`
	DeviceName         string      `json:"device_name"`
	DeviceType         string      `json:"device_type"`
	DeviceSerialNumber string      `json:"device_serial_number"`
	DeviceModel        string      `json:"device_model"`
	DeviceBrand        string      `json:"device_brand"`
	DeviceLocation     GeoLocation `json:"device_location"`
}

type SopMeta struct {
	// ใส่ field เพิ่มเติมถ้ามี
}

type GeoLocation struct {
	Latitude  StringOrNumber `json:"latitude"`
	Longitude StringOrNumber `json:"longitude"`
}

type WorkflowBySubType struct {
	Id            string    `json:"id"`
	TypeID        string    `json:"typeId"`
	STypeID       string    `json:"sTypeId"`
	STypeCode     string    `json:"sTypeCode"`
	OrgID         string    `json:"orgId"`
	EN            string    `json:"en"`
	TH            string    `json:"th"`
	WFID          string    `json:"wfId"`
	CaseSLA       string    `json:"caseSla"`
	Priority      string    `json:"priority"`
	UserSkillList []string  `json:"userSkillList"`
	UnitPropLists []string  `json:"unitPropLists"`
	Active        bool      `json:"active"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	CreatedBy     string    `json:"createdBy"`
	UpdatedBy     string    `json:"updatedBy"`

	// extra from wf_definitions
	WfTitle    string `json:"wfTitle" db:"wfTitle"`
	WfDesc     string `json:"wfDesc" db:"wfDesc"`
	WfVersions string `json:"wfVersions" db:"wfVersions"`
	WfSection  string `json:"wfSection" db:"wfSection"`
	WfData     string `json:"wfData" db:"wfData"`
	WfNodeId   string `json:"wfNodeId" db:"wfNodeId"`
}

type AssignedEmployee struct {
	UserEmployeeCode string `json:"user_employee_code"`
	UserFirstname    string `json:"user_firstname"`
	UserLastname     string `json:"user_lastname"`
	UserAvatar       string `json:"user_avatar"`
	UserPhone        string `json:"user_phone"`
}

type StringOrNumber string

func (s *StringOrNumber) UnmarshalJSON(data []byte) error {
	dataStr := strings.TrimSpace(string(data))

	// null -> empty
	if dataStr == "null" {
		*s = ""
		return nil
	}

	// "" -> empty
	if dataStr == `""` {
		*s = ""
		return nil
	}

	// number -> convert to string
	if dataStr != "" && (dataStr[0] >= '0' && dataStr[0] <= '9' || dataStr[0] == '-') {
		*s = StringOrNumber(dataStr)
		return nil
	}

	// quoted string
	var tmp string
	if err := json.Unmarshal(data, &tmp); err == nil {
		*s = StringOrNumber(tmp)
		return nil
	}

	return fmt.Errorf("StringOrNumber: invalid value %s", dataStr)
}

func (a *AssignedEmployee) UnmarshalJSON(data []byte) error {
	// ถ้าเป็น null
	if string(data) == "null" {
		*a = AssignedEmployee{}
		return nil
	}

	// ถ้าเป็น string ว่าง ""
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		if s == "" {
			*a = AssignedEmployee{}
			return nil
		}
		return fmt.Errorf("expected object or null, got string: %s", s)
	}

	// ปกติเป็น object
	type Alias AssignedEmployee
	var tmp Alias
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	*a = AssignedEmployee(tmp)
	return nil
}
