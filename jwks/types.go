package jwks

import (
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type Claims interface {
	TokenUse() string
	ClientId() string
	Username() string
	UserID() string
	Role() string
	Scope() string
	JTI() string
}

type CognitoAccessClaims struct {
	jwt.RegisteredClaims
	Groups []string `json:"cognito:groups"`
	Client string   `json:"client_id,omitempty"`
	Bound  string   `json:"scope,omitempty"`
	Use    string   `json:"token_use"`
	User   string   `json:"cognito:username"`
	Sub    string   `json:"sub"`
}

func (c *CognitoAccessClaims) TokenUse() string {
	return c.Use
}

func (c *CognitoAccessClaims) ClientId() string {
	return c.Client
}

func (c *CognitoAccessClaims) Username() string {
	return c.User
}

func (c *CognitoAccessClaims) UserID() string {
	return c.Sub
}

func (c *CognitoAccessClaims) Role() string {
	if len(c.Groups) == 0 {
		return "NONE"
	}

	return strings.ToUpper(strings.TrimSpace(c.Groups[0]))
}

func (c *CognitoAccessClaims) Scope() string {
	return c.Bound
}

func (c *CognitoAccessClaims) JTI() string {
	return c.RegisteredClaims.ID
}
