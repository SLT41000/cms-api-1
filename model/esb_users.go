package model

type ESBUserStaffPayload struct {
	UserCode  string       `json:"user_code"`
	FirstName string       `json:"first_name"`
	LastName  string       `json:"last_name"`
	Phone     string       `json:"phone"`
	Email     string       `json:"email"`
	Username  string       `json:"username"`
	Password  string       `json:"password"`
	Workspace string       `json:"workspace"`
	Namespace *interface{} `json:"namespace"`
	Source    string       `json:"source"`
}
