package model

import (
	"time"
)

type Um_User_Login struct {
	ID                    string     `json:"id"`
	OrgID                 string     `json:"orgId"`
	OrgName               string     `json:"orgName"`
	DisplayName           string     `json:"displayName"`
	Title                 string     `json:"title"`
	FirstName             string     `json:"firstName"`
	MiddleName            *string    `json:"middleName"`
	LastName              string     `json:"lastName"`
	CitizenID             string     `json:"citizenId"`
	Bod                   time.Time  `json:"bod"`
	Blood                 string     `json:"blood"`
	Gender                string     `json:"gender"`
	MobileNo              *string    `json:"mobileNo"`
	Address               *string    `json:"address"`
	Photo                 *string    `json:"photo"`
	Username              string     `json:"username"`
	Password              string     `json:"password"`
	Email                 *string    `json:"email"`
	RoleID                string     `json:"roleId"`
	Permission            []string   `json:"permission"`
	RoleName              string     `json:"roleName"`
	UserType              string     `json:"userType"`
	EmpID                 string     `json:"empId"`
	DeptID                string     `json:"deptId"`
	CommID                string     `json:"commId"`
	StnID                 string     `json:"stnId"`
	Active                bool       `json:"active"`
	ActivationToken       *string    `json:"activationToken"`
	LastActivationRequest *int64     `json:"lastActivationRequest"`
	LostPasswordRequest   *int64     `json:"lostPasswordRequest"`
	SignupStamp           *int64     `json:"signupStamp"`
	IsLogin               bool       `json:"islogin"`
	LastLogin             *time.Time `json:"lastLogin"`
	CreatedAt             time.Time  `json:"createdAt"`
	UpdatedAt             time.Time  `json:"updatedAt"`
	CreatedBy             string     `json:"createdBy"`
	UpdatedBy             string     `json:"updatedBy"`
	DistIdLists           *[]string  `json:"distIdLists,omitempty"` // ✅ optional JSON array of UUIDs
}

type Um_User struct {
	ID                    string                    `json:"id"`
	OrgID                 string                    `json:"orgId"`
	OrgName               string                    `json:"orgName"`
	DisplayName           string                    `json:"displayName"`
	Title                 string                    `json:"title"`
	FirstName             string                    `json:"firstName"`
	MiddleName            *string                   `json:"middleName"`
	LastName              string                    `json:"lastName"`
	CitizenID             string                    `json:"citizenId"`
	Bod                   time.Time                 `json:"bod"`
	Blood                 string                    `json:"blood"`
	Gender                string                    `json:"gender"`
	MobileNo              *string                   `json:"mobileNo"`
	Address               *string                   `json:"address"`
	Photo                 *string                   `json:"photo"`
	Username              string                    `json:"username"`
	Password              string                    `json:"password"`
	Email                 *string                   `json:"email"`
	RoleID                string                    `json:"roleId"`
	RoleName              string                    `json:"roleName"`
	UserType              string                    `json:"userType"`
	EmpID                 string                    `json:"empId"`
	DeptID                string                    `json:"deptId"`
	CommID                string                    `json:"commId"`
	StnID                 string                    `json:"stnId"`
	Active                bool                      `json:"active"`
	ActivationToken       *string                   `json:"activationToken"`
	LastActivationRequest *int64                    `json:"lastActivationRequest"`
	LostPasswordRequest   *int64                    `json:"lostPasswordRequest"`
	SignupStamp           *int64                    `json:"signupStamp"`
	IsLogin               bool                      `json:"islogin"`
	LastLogin             *time.Time                `json:"lastLogin"`
	CreatedAt             time.Time                 `json:"createdAt"`
	UpdatedAt             time.Time                 `json:"updatedAt"`
	CreatedBy             string                    `json:"createdBy"`
	UpdatedBy             string                    `json:"updatedBy"`
	Skills                *[]map[string]interface{} `json:"skills"`
	Areas                 *[]map[string]interface{} `json:"areas"`
}

type User_UnitInfo struct {
	ID          string                    `json:"id"`
	DisplayName string                    `json:"displayName"`
	Title       string                    `json:"title"`
	FirstName   string                    `json:"firstName"`
	MiddleName  *string                   `json:"middleName"`
	LastName    string                    `json:"lastName"`
	Gender      string                    `json:"gender"`
	MobileNo    *string                   `json:"mobileNo"`
	Address     *string                   `json:"address"`
	Photo       *string                   `json:"photo"`
	Username    string                    `json:"username"`
	Email       *string                   `json:"email"`
	DeptID      string                    `json:"deptId"`
	CommID      string                    `json:"commId"`
	StnID       string                    `json:"stnId"`
	CreatedAt   time.Time                 `json:"createdAt"`
	UpdatedAt   time.Time                 `json:"updatedAt"`
	CreatedBy   string                    `json:"createdBy"`
	UpdatedBy   string                    `json:"updatedBy"`
	Skills      *[]map[string]interface{} `json:"skills"`
	Areas       *[]map[string]interface{} `json:"areas"`
}

type UserInput struct {
	DisplayName           string     `json:"displayName"`
	Title                 string     `json:"title"`
	FirstName             string     `json:"firstName"`
	MiddleName            string     `json:"middleName"`
	LastName              string     `json:"lastName"`
	CitizenID             string     `json:"citizenId"`
	Bod                   time.Time  `json:"bod"`
	Blood                 string     `json:"blood"`
	Gender                *int64     `json:"gender"`
	MobileNo              string     `json:"mobileNo"`
	Address               string     `json:"address"`
	Photo                 *string    `json:"photo"`
	Username              string     `json:"username"`
	Password              string     `json:"password"`
	Email                 string     `json:"email"`
	RoleID                string     `json:"roleId"`
	UserType              *int64     `json:"userType"`
	EmpID                 string     `json:"empId"`
	DeptID                string     `json:"deptId"`
	CommID                string     `json:"commId"`
	StnID                 string     `json:"stnId"`
	Active                bool       `json:"active"`
	LastActivationRequest *int64     `json:"lastActivationRequest"`
	LostPasswordRequest   *int64     `json:"lostPasswordRequest"`
	SignupStamp           *int64     `json:"signupStamp"`
	IsLogin               bool       `json:"islogin"`
	LastLogin             *time.Time `json:"lastLogin"`
}

// Remove Password form UserUpdate by Delta 26/08/2568
type UserUpdate struct {
	DisplayName           string     `json:"displayName"`
	Title                 string     `json:"title"`
	FirstName             string     `json:"firstName"`
	MiddleName            string     `json:"middleName"`
	LastName              string     `json:"lastName"`
	CitizenID             string     `json:"citizenId"`
	Bod                   string     `json:"bod"`
	Blood                 string     `json:"blood"`
	Gender                *int64     `json:"gender"`
	MobileNo              string     `json:"mobileNo"`
	Address               string     `json:"address"`
	Photo                 *string    `json:"photo"`
	Username              string     `json:"username"`
	Email                 string     `json:"email"`
	RoleID                string     `json:"roleId"`
	UserType              *int64     `json:"userType"`
	EmpID                 string     `json:"empId"`
	DeptID                string     `json:"deptId"`
	CommID                string     `json:"commId"`
	StnID                 string     `json:"stnId"`
	Active                bool       `json:"active"`
	LastActivationRequest *int64     `json:"lastActivationRequest"`
	LostPasswordRequest   *int64     `json:"lostPasswordRequest"`
	SignupStamp           *int64     `json:"signupStamp"`
	IsLogin               bool       `json:"islogin"`
	LastLogin             *time.Time `json:"lastLogin"`
}

//add Reset Password and Change Password form UserUpdate by Delta 26/08/2568

// ResetPasswordRequest สำหรับ reset password
type ResetPasswordRequest struct {
	Username    string `json:"username" binding:"required"`
	Email       string `json:"email" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

// ChangePasswordRequest สำหรับ change password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required"`
}

type UserContact struct {
	OrgID        string    `json:"orgId"`
	Username     string    `json:"username"`
	ContactName  string    `json:"contactName"`
	ContactPhone string    `json:"contactPhone"`
	ContactAddr  any       `json:"contactAddr"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	CreatedBy    string    `json:"createdBy"`
	UpdatedBy    string    `json:"updatedBy"`
}

type UserSkill struct {
	OrgID     string    `json:"orgId"`
	UserName  string    `json:"userName"`
	SkillID   string    `json:"skillId"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	CreatedBy string    `json:"createdBy"`
	UpdatedBy string    `json:"updatedBy"`
}

type UserSocial struct {
	OrgID      string    `json:"orgId"`
	Username   string    `json:"username"`
	SocialType string    `json:"socialType"`
	SocialID   string    `json:"socialId"`
	SocialName string    `json:"socialName"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	CreatedBy  string    `json:"createdBy"`
	UpdatedBy  string    `json:"updatedBy"`
}

type UserContactInsert struct {
	OrgID        string `json:"orgId"`
	Username     string `json:"username"`
	ContactName  string `json:"contactName"`
	ContactPhone string `json:"contactPhone"`
	ContactAddr  any    `json:"contactAddr"`
}

type UserContactInsertUpdate struct {
	ContactName  string `json:"contactName"`
	ContactPhone string `json:"contactPhone"`
	ContactAddr  any    `json:"contactAddr"`
}

type UserSkillInsert struct {
	OrgID    string `json:"orgId"`
	UserName string `json:"userName"`
	SkillID  string `json:"skillId"`
	Active   bool   `json:"active"`
}

type UserSkillUpdate struct {
	SkillID string `json:"skillId"`
	Active  bool   `json:"active"`
}

type UserSocialInsert struct {
	OrgID      string `json:"orgId"`
	Username   string `json:"username"`
	SocialType string `json:"socialType"`
	SocialID   string `json:"socialId"`
	SocialName string `json:"socialName"`
}

type UserSocialUpdate struct {
	OrgID      string `json:"orgId"`
	Username   string `json:"username"`
	SocialType string `json:"socialType"`
	SocialID   string `json:"socialId"`
	SocialName string `json:"socialName"`
	UpdatedBy  string `json:"updatedBy"`
}

type UmGroup struct {
	ID        int       `json:"id"`
	OrgID     string    `json:"orgId"`
	GrpID     string    `json:"grpId"`
	En        string    `json:"en"`
	Th        string    `json:"th"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	CreatedBy string    `json:"createdBy"`
	UpdatedBy string    `json:"updatedBy"`
}

// UserProfile สำหรับเก็บข้อมูล user profile ที่ใช้ในการจัดการ notifications
// แยกออกจาก Um_User เพื่อไม่ให้กระทบกับ model เดิม
type UserProfile struct {
	EmpID  string `json:"empId"`
	OrgID  string `json:"orgId"`
	RoleID string `json:"roleId"`
	DeptID string `json:"deptId"`
	StnID  string `json:"stnId"`
	CommID string `json:"commId"`
	GrpID  string `json:"grpId"`
}

type KafkaUnitStatus struct {
	Workspace string `json:"workspace"`  // maps to orgId or org name
	UserCode  string `json:"user_code"`  // maps to username in mdm_units
	CheckedIn bool   `json:"checked_in"` // maps to isLogin field
}
