package oauth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/Domedik/trussrod/jwks"
	"github.com/Domedik/trussrod/settings"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/golang-jwt/jwt/v5"
)

type Cognito struct {
	client *cognitoidentityprovider.Client
	hash   func(username string) string
	jwks   *jwks.JWKS
	config *settings.OAuthConfig
}

func (c *Cognito) getUserByEmail(ctx context.Context, email string) (string, error) {
	input := &cognitoidentityprovider.ListUsersInput{
		UserPoolId: aws.String(c.config.UserPool),
		Filter:     aws.String("email = \"" + email + "\""),
	}

	result, err := c.client.ListUsers(ctx, input)
	if err != nil {
		return "", err
	}

	if len(result.Users) == 0 {
		return "", errors.New("no user found")
	}

	for _, attr := range result.Users[0].Attributes {
		if *attr.Name == "sub" {
			return *attr.Value, nil
		}
	}

	return "", errors.New("no user found")
}

func (c *Cognito) Login(ctx context.Context, username, password string) (*types.AuthenticationResultType, error) {
	params := &cognitoidentityprovider.AdminInitiateAuthInput{
		AuthFlow:   types.AuthFlowType(c.config.Flow),
		ClientId:   aws.String(c.config.App),
		UserPoolId: aws.String(c.config.UserPool),
		AuthParameters: map[string]string{
			"USERNAME":    username,
			"PASSWORD":    password,
			"SECRET_HASH": c.hash(username),
		},
	}
	result, err := c.client.AdminInitiateAuth(ctx, params)
	if err != nil {
		return nil, err
	}
	return result.AuthenticationResult, nil
}

func (c *Cognito) RequestResetPassword(ctx context.Context, email string) (*cognitoidentityprovider.AdminResetUserPasswordOutput, error) {
	subs, err := c.getUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	input := &cognitoidentityprovider.AdminResetUserPasswordInput{
		UserPoolId: aws.String(c.config.UserPool),
		Username:   aws.String(subs),
	}

	return c.client.AdminResetUserPassword(ctx, input)
}

func (c *Cognito) ConfirmResetPassword(ctx context.Context, email, code, password string) error {
	username, err := c.getUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	hash := c.hash(username)

	input := &cognitoidentityprovider.ConfirmForgotPasswordInput{
		ClientId:         aws.String(c.config.App),
		Username:         aws.String(username),
		ConfirmationCode: aws.String(code),
		Password:         aws.String(password),
		SecretHash:       aws.String(hash),
	}

	_, err = c.client.ConfirmForgotPassword(ctx, input)
	return err
}

func (c *Cognito) ConfirmUserSignup(ctx context.Context, email, code string) error {
	username, err := c.getUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	hash := c.hash(username)

	input := &cognitoidentityprovider.ConfirmSignUpInput{
		ClientId:         aws.String(c.config.App),
		Username:         aws.String(username),
		ConfirmationCode: aws.String(code),
		SecretHash:       aws.String(hash),
	}

	_, err = c.client.ConfirmSignUp(ctx, input)
	return err
}

func (c *Cognito) ValidateToken(ctx context.Context, jwtString string) (jwt.Claims, error) {
	token, err := jwt.Parse(jwtString, c.jwks.Keyfunc)
	if token.Valid {
		return token.Claims, nil
	}
	return nil, err
}

func NewCognitoClient(c *settings.DomedikConfig) (*Cognito, error) {
	ctx := context.Background()
	conf, err := config.LoadDefaultConfig(ctx, config.WithRegion(c.Cloud.Region))

	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/.well-known/jwks.json", c.OAuth.Issuer)

	return &Cognito{
		client: cognitoidentityprovider.NewFromConfig(conf),
		hash: func(username string) string {
			data := username + c.OAuth.App
			h := hmac.New(sha256.New, []byte(c.OAuth.Secret))
			h.Write([]byte(data))
			hash := h.Sum(nil)
			secretHash := base64.StdEncoding.EncodeToString(hash)
			return secretHash
		},
		jwks:   jwks.NewJWKSCache(url, time.Minute*10),
		config: c.OAuth,
	}, nil
}

func (*Cognito) Close() error {
	return nil
}

func (c *Cognito) Ping(ctx context.Context) error {
	return nil
}
