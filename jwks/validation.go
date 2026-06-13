package jwks

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Validator interface {
	GrantAccess(tokenString string) (Claims, error)
}

type BaseValidator struct {
	jwks         *JWKS
	issuer       string
	audience     string
	leeway       time.Duration
	algWhitelist []string
}

func NewBaseValidator(issuer, audience string, leeway, cacheTTL time.Duration) *BaseValidator {
	url := fmt.Sprintf("%s/.well-known/jwks.json", issuer)
	jwks := NewJWKSCache(url, cacheTTL)

	return &BaseValidator{
		jwks:         jwks,
		issuer:       issuer,
		audience:     audience,
		leeway:       leeway,
		algWhitelist: []string{"RS256", "ES256"},
	}
}

type CognitoValidator struct {
	*BaseValidator
}

func NewCognitoValidator(issuer, audience string, leeway, cacheTTL time.Duration) *CognitoValidator {
	return &CognitoValidator{
		BaseValidator: NewBaseValidator(issuer, audience, leeway, cacheTTL),
	}
}

func (v *CognitoValidator) GrantAccess(tokenString string) (Claims, error) {
	claims := &CognitoAccessClaims{}
	parser := jwt.NewParser(
		jwt.WithValidMethods(v.algWhitelist),
		jwt.WithIssuer(v.issuer),
		jwt.WithLeeway(v.leeway),
	)
	token, err := parser.ParseWithClaims(tokenString, claims, v.jwks.Keyfunc)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	if claims.TokenUse() != "access" {
		fmt.Printf("invalid token use: %s", claims.TokenUse())
		return nil, errors.New("invalid token use")
	}

	if v.audience != "" {
		if claims.ClientId() != v.audience {
			return nil, errors.New("audience/client_id mismatch")
		}
	}

	return claims, nil
}
