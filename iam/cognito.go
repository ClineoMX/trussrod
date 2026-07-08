package iam

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/clineomx/trussrod/apperr"
)

type CognitoClient struct {
	client   client
	UserPool string
}

func (p *CognitoClient) CreateUser(ctx context.Context, email, password, name, familyName string) (*User, error) {
	username, err := NewUUIDV4()
	if err != nil {
		return nil, err
	}

	out, err := p.client.AdminCreateUser(ctx, &cognitoidentityprovider.AdminCreateUserInput{
		UserPoolId:        aws.String(p.UserPool),
		Username:          aws.String(username),
		TemporaryPassword: aws.String(password),
		MessageAction:     types.MessageActionTypeSuppress,
		UserAttributes: []types.AttributeType{
			{Name: aws.String("email"), Value: aws.String(email)},
			{Name: aws.String("email_verified"), Value: aws.String("true")},
			{Name: aws.String("given_name"), Value: aws.String(name)},
			{Name: aws.String("family_name"), Value: aws.String(familyName)},
		},
	})
	if err != nil {
		return nil, translate(err)
	}

	_, err = p.client.AdminSetUserPassword(ctx, &cognitoidentityprovider.AdminSetUserPasswordInput{
		UserPoolId: aws.String(p.UserPool),
		Username:   aws.String(username),
		Password:   aws.String(password),
		Permanent:  true,
	})
	if err != nil {
		_ = p.DeleteUser(ctx, username)
		return nil, translate(err)
	}

	for _, attr := range out.User.Attributes {
		if aws.ToString(attr.Name) == "sub" {
			return &User{Sub: aws.ToString(attr.Value), Username: username}, nil
		}
	}
	_ = p.DeleteUser(ctx, username)
	return nil, fmt.Errorf("COGNITO:NO_SUB %s", username)
}

func (p *CognitoClient) AddToGroup(ctx context.Context, username, role string) error {
	_, err := p.client.AdminAddUserToGroup(ctx, &cognitoidentityprovider.AdminAddUserToGroupInput{
		UserPoolId: aws.String(p.UserPool),
		Username:   aws.String(username),
		GroupName:  aws.String(strings.ToUpper(role)),
	})
	return translate(err)
}

func (p *CognitoClient) DeleteUser(ctx context.Context, username string) error {
	_, err := p.client.AdminDeleteUser(ctx, &cognitoidentityprovider.AdminDeleteUserInput{
		UserPoolId: aws.String(p.UserPool),
		Username:   aws.String(username),
	})
	return translate(err)
}

type client interface {
	AdminCreateUser(context.Context, *cognitoidentityprovider.AdminCreateUserInput, ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.AdminCreateUserOutput, error)
	AdminSetUserPassword(context.Context, *cognitoidentityprovider.AdminSetUserPasswordInput, ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.AdminSetUserPasswordOutput, error)
	AdminAddUserToGroup(context.Context, *cognitoidentityprovider.AdminAddUserToGroupInput, ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.AdminAddUserToGroupOutput, error)
	AdminDeleteUser(context.Context, *cognitoidentityprovider.AdminDeleteUserInput, ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.AdminDeleteUserOutput, error)
}

func translate(err error) error {
	var invalidPassword *types.InvalidPasswordException
	if errors.As(err, &invalidPassword) {
		return apperr.BadRequest("COGNITO:INVALID_PASSWORD")
	}
	var usernameExists *types.UsernameExistsException
	var aliasExists *types.AliasExistsException
	if errors.As(err, &usernameExists) || errors.As(err, &aliasExists) {
		return apperr.Conflict("COGNITO:EMAIL_TAKEN")
	}
	return err
}
