package jwks

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Validator is a struct to hold jwt validation logic.
type Validator struct {
	jwks         *JWKS
	issuer       string
	audience     string
	leeway       time.Duration
	algWhitelist []string
}

// NewValidator creates a new Validator instance with jwks ttl defaulted to 10 minutes.
func NewValidator(url, issuer, audience string) (*Validator, error) {
	jwks := NewJWKSCache(url, 10*time.Minute)

	return &Validator{
		jwks:         jwks,
		issuer:       issuer,
		audience:     audience,
		leeway:       2 * time.Minute,
		algWhitelist: []string{"RS256", "ES256"},
	}, nil
}

// Close destroys instance
func (v *Validator) Close() error {
	return nil
}

// GrantAccess validates access token and return AccessClaims if valid.
func (v *Validator) GrantAccess(tokenString string) (*AccessClaims, error) {
	claims := &AccessClaims{}
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

	if claims.TokenUse != "access" {
		return nil, errors.New("invalid token use")
	}

	if v.audience != "" {
		if claims.ClientId != v.audience {
			return nil, errors.New("audience/client_id mismatch")
		}
	}

	return claims, nil
}

// GrantIdentity validates identity token and return AccessClaims if valid.
func (v *Validator) GrantIdentity(tokenString string) (*IdentityClaims, error) {
	claims := &IdentityClaims{}
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

	if claims.TokenUse != "id" {
		return nil, errors.New("invalid token use")
	}

	return claims, nil
}
