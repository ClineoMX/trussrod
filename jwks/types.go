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
	groups   []string `json:"cognito:groups"`
	clientId string   `json:"client_id,omitempty"`
	scope    string   `json:"scope,omitempty"`
	Use      string   `json:"token_use"`
	username string   `json:"cognito:username"`
	userId   string   `json:"sub"`
}

func (c *CognitoAccessClaims) TokenUse() string {
	return c.Use
}

func (c *CognitoAccessClaims) ClientId() string {
	return c.clientId
}

func (c *CognitoAccessClaims) Username() string {
	return c.username
}

func (c *CognitoAccessClaims) UserID() string {
	return c.userId
}

func (c *CognitoAccessClaims) Role() string {
	if len(c.groups) == 0 {
		return "NONE"
	}

	return strings.ToUpper(strings.TrimSpace(c.groups[0]))
}

func (c *CognitoAccessClaims) Scope() string {
	return c.scope
}

func (c *CognitoAccessClaims) JTI() string {
	return c.RegisteredClaims.ID
}
