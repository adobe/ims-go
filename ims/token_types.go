package ims

type TokenType string

const (
	AccessToken       TokenType = "access_token"
	RefreshToken      TokenType = "refresh_token"
	ServiceToken      TokenType = "service_token"
	DeviceToken       TokenType = "device_token"
	AuthorizationCode TokenType = "authorization_code"
)
