package model

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
