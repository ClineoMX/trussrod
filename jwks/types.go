package jwks

import (
	"github.com/golang-jwt/jwt/v5"
)

// AccessClaims are the required fields from the access token.
type AccessClaims struct {
	jwt.RegisteredClaims
	CognitoGroups []string `json:"cognito:groups"`
	ClientId      string   `json:"client_id,omitempty"`
	Scope         string   `json:"scope,omitempty"`
	TokenUse      string   `json:"token_use"`
	Username      string   `json:"cognito:username"`
	UserID        string   `json:"sub"`
}

// IndentityClaims are the required fields from the Indentity token.
type IdentityClaims struct {
	jwt.RegisteredClaims
	TokenUse   string `json:"token_use"`
	ClientId   string `json:"client_id,omitempty"`
	CMKARN     string `json:"custom:CMK"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	IsVerified bool   `json:"email_verified"`
	Name       string `json:"name"`
	FamilyName string `json:"family_name"`
	Gender     string `json:"gender"`
}
