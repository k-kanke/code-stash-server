package dto

type DeviceCodeRequest struct {
	ClientID string `json:"client_id"`
	Scope    string `json:"scope"`
}

type DeviceCodeResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete,omitempty"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

type VerifyDeviceCodeRequest struct {
	UserCode string `json:"user_code"`
}

type DeviceCodeStatusResponse struct {
	DeviceCode string   `json:"device_code"`
	UserCode   string   `json:"user_code"`
	ClientID   string   `json:"client_id"`
	ClientName string   `json:"client_name"`
	Scope      []string `json:"scope"`
	Status     string   `json:"status"`
	ExpiresAt  string   `json:"expires_at"`
	Interval   int      `json:"interval"`
}

type TokenRequest struct {
	GrantType    string `json:"grant_type"`
	DeviceCode   string `json:"device_code"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type TokenResponse struct {
	AccessToken  string  `json:"access_token"`
	TokenType    string  `json:"token_type"`
	ExpiresIn    int     `json:"expires_in"`
	RefreshToken *string `json:"refresh_token,omitempty"`
	Scope        string  `json:"scope,omitempty"`
}
