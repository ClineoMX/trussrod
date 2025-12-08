package oauth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/Domedik/trussrod/jwks"
	"github.com/Domedik/trussrod/settings"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	provider "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

type Cognito struct {
	client *provider.Client
	hash   func(username string) string
	jwks   *jwks.JWKS
	config *settings.OAuthConfig
}

func (c *Cognito) Login(ctx context.Context, username, password string) (*LoginOutput, error) {
	params := &provider.AdminInitiateAuthInput{
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

	return &LoginOutput{
		Access:   *result.AuthenticationResult.AccessToken,
		Identity: *result.AuthenticationResult.IdToken,
		Refresh:  *result.AuthenticationResult.RefreshToken,
	}, nil
}

func (c *Cognito) RequestResetPassword(ctx context.Context, email string) error {
	input := &provider.AdminResetUserPasswordInput{
		UserPoolId: aws.String(c.config.UserPool),
		Username:   aws.String(email),
	}

	_, err := c.client.AdminResetUserPassword(ctx, input)
	return err
}

func (c *Cognito) ConfirmResetPassword(ctx context.Context, email, code, password string) error {
	hash := c.hash(email)

	input := &provider.ConfirmForgotPasswordInput{
		ClientId:         aws.String(c.config.App),
		Username:         aws.String(email),
		ConfirmationCode: aws.String(code),
		Password:         aws.String(password),
		SecretHash:       aws.String(hash),
	}

	_, err := c.client.ConfirmForgotPassword(ctx, input)
	return err
}

func (c *Cognito) ConfirmUserSignup(ctx context.Context, email, code string) error {
	hash := c.hash(email)

	input := &provider.ConfirmSignUpInput{
		ClientId:         aws.String(c.config.App),
		Username:         aws.String(email),
		ConfirmationCode: aws.String(code),
		SecretHash:       aws.String(hash),
	}

	_, err := c.client.ConfirmSignUp(ctx, input)
	return err
}

func NewCognitoClient(c *settings.DomedikConfig) (*Cognito, error) {
	ctx := context.Background()
	conf, err := config.LoadDefaultConfig(ctx, config.WithRegion(c.Cloud.Region))
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/.well-known/jwks.json", c.OAuth.Issuer)

	return &Cognito{
		client: provider.NewFromConfig(conf),
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

func (c *Cognito) CreatePatientUser(ctx context.Context, username string, phone string) error {
	input := provider.AdminCreateUserInput{
		UserPoolId: aws.String(c.config.PatientsUserPool),
		Username:   aws.String(username),
		UserAttributes: []types.AttributeType{
			{
				Name:  aws.String("phone_number"),
				Value: aws.String(phone),
			},
			{
				Name:  aws.String("phone_number_verified"),
				Value: aws.String("true"),
			},
		},
		MessageAction: types.MessageActionTypeSuppress,
	}

	_, err := c.client.AdminCreateUser(ctx, &input)

	return err
}

func (c *Cognito) DeletePatientUser(ctx context.Context, username string) error {
	input := provider.AdminDeleteUserInput{
		UserPoolId: aws.String(c.config.PatientsUserPool),
		Username:   aws.String(username),
	}

	_, err := c.client.AdminDeleteUser(ctx, &input)

	return err
}
