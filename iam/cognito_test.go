package iam

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/clineomx/trussrod/apperr"
)

type mockCognitoAPI struct {
	createUserFunc  func(in *cognitoidentityprovider.AdminCreateUserInput) (*cognitoidentityprovider.AdminCreateUserOutput, error)
	setPasswordFunc func(in *cognitoidentityprovider.AdminSetUserPasswordInput) (*cognitoidentityprovider.AdminSetUserPasswordOutput, error)
	addToGroupFunc  func(in *cognitoidentityprovider.AdminAddUserToGroupInput) (*cognitoidentityprovider.AdminAddUserToGroupOutput, error)
	deleteUserFunc  func(in *cognitoidentityprovider.AdminDeleteUserInput) (*cognitoidentityprovider.AdminDeleteUserOutput, error)

	setPasswordCalls []*cognitoidentityprovider.AdminSetUserPasswordInput
	deleteUserCalls  []string
}

func (m *mockCognitoAPI) AdminCreateUser(_ context.Context, in *cognitoidentityprovider.AdminCreateUserInput, _ ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.AdminCreateUserOutput, error) {
	return m.createUserFunc(in)
}

func (m *mockCognitoAPI) AdminSetUserPassword(_ context.Context, in *cognitoidentityprovider.AdminSetUserPasswordInput, _ ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.AdminSetUserPasswordOutput, error) {
	m.setPasswordCalls = append(m.setPasswordCalls, in)
	return m.setPasswordFunc(in)
}

func (m *mockCognitoAPI) AdminAddUserToGroup(_ context.Context, in *cognitoidentityprovider.AdminAddUserToGroupInput, _ ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.AdminAddUserToGroupOutput, error) {
	return m.addToGroupFunc(in)
}

func (m *mockCognitoAPI) AdminDeleteUser(_ context.Context, in *cognitoidentityprovider.AdminDeleteUserInput, _ ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.AdminDeleteUserOutput, error) {
	m.deleteUserCalls = append(m.deleteUserCalls, aws.ToString(in.Username))
	if m.deleteUserFunc != nil {
		return m.deleteUserFunc(in)
	}
	return &cognitoidentityprovider.AdminDeleteUserOutput{}, nil
}

func attrMap(attrs []types.AttributeType) map[string]string {
	out := make(map[string]string, len(attrs))
	for _, a := range attrs {
		out[aws.ToString(a.Name)] = aws.ToString(a.Value)
	}
	return out
}

func TestCognitoClient_CreateUser_Success(t *testing.T) {
	m := &mockCognitoAPI{
		createUserFunc: func(in *cognitoidentityprovider.AdminCreateUserInput) (*cognitoidentityprovider.AdminCreateUserOutput, error) {
			if aws.ToString(in.UserPoolId) != "pool-1" {
				t.Fatalf("expected UserPoolId pool-1, got %q", aws.ToString(in.UserPoolId))
			}
			if in.MessageAction != types.MessageActionTypeSuppress {
				t.Fatalf("expected MessageAction suppress, got %v", in.MessageAction)
			}
			attrs := attrMap(in.UserAttributes)
			if attrs["email"] != "alice@example.com" {
				t.Fatalf("expected email attribute alice@example.com, got %q", attrs["email"])
			}
			if attrs["email_verified"] != "true" {
				t.Fatalf("expected email_verified true, got %q", attrs["email_verified"])
			}
			if attrs["given_name"] != "Alice" || attrs["family_name"] != "Smith" {
				t.Fatalf("unexpected name attributes: %+v", attrs)
			}
			return &cognitoidentityprovider.AdminCreateUserOutput{
				User: &types.UserType{
					Attributes: []types.AttributeType{
						{Name: aws.String("sub"), Value: aws.String("sub-123")},
					},
				},
			}, nil
		},
		setPasswordFunc: func(in *cognitoidentityprovider.AdminSetUserPasswordInput) (*cognitoidentityprovider.AdminSetUserPasswordOutput, error) {
			if !in.Permanent {
				t.Fatal("expected permanent password to be set")
			}
			return &cognitoidentityprovider.AdminSetUserPasswordOutput{}, nil
		},
	}

	c := &CognitoClient{client: m, UserPool: "pool-1"}
	user, err := c.CreateUser(context.Background(), "alice@example.com", "Passw0rd!", "Alice", "Smith")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Sub != "sub-123" {
		t.Fatalf("expected sub sub-123, got %q", user.Sub)
	}
	if user.Username == "" {
		t.Fatal("expected a generated username")
	}
	if len(m.setPasswordCalls) != 1 || aws.ToString(m.setPasswordCalls[0].Username) != user.Username {
		t.Fatalf("expected AdminSetUserPassword called with generated username %q", user.Username)
	}
	if len(m.deleteUserCalls) != 0 {
		t.Fatalf("expected no cleanup deletes on success, got %v", m.deleteUserCalls)
	}
}

func TestCognitoClient_CreateUser_InvalidPassword(t *testing.T) {
	m := &mockCognitoAPI{
		createUserFunc: func(*cognitoidentityprovider.AdminCreateUserInput) (*cognitoidentityprovider.AdminCreateUserOutput, error) {
			return nil, &types.InvalidPasswordException{Message: aws.String("too weak")}
		},
	}

	c := &CognitoClient{client: m, UserPool: "pool-1"}
	user, err := c.CreateUser(context.Background(), "alice@example.com", "weak", "Alice", "Smith")
	if user != nil {
		t.Fatalf("expected nil user, got %+v", user)
	}
	var appErr *apperr.AppError
	if !errors.As(err, &appErr) || appErr.Code != "BAD_REQUEST" {
		t.Fatalf("expected BAD_REQUEST app error, got %v", err)
	}
	if len(m.deleteUserCalls) != 0 {
		t.Fatalf("expected no cleanup delete when create itself fails, got %v", m.deleteUserCalls)
	}
}

func TestCognitoClient_CreateUser_UsernameOrAliasExists(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"username exists", &types.UsernameExistsException{Message: aws.String("taken")}},
		{"alias exists", &types.AliasExistsException{Message: aws.String("taken")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockCognitoAPI{
				createUserFunc: func(*cognitoidentityprovider.AdminCreateUserInput) (*cognitoidentityprovider.AdminCreateUserOutput, error) {
					return nil, tt.err
				},
			}

			c := &CognitoClient{client: m, UserPool: "pool-1"}
			_, err := c.CreateUser(context.Background(), "alice@example.com", "Passw0rd!", "Alice", "Smith")
			var appErr *apperr.AppError
			if !errors.As(err, &appErr) || appErr.Code != "RESOURCE_CONFLICT" {
				t.Fatalf("expected RESOURCE_CONFLICT app error, got %v", err)
			}
		})
	}
}

func TestCognitoClient_CreateUser_SetPasswordFails_CleansUpUser(t *testing.T) {
	setPasswordErr := errors.New("set password boom")
	m := &mockCognitoAPI{
		createUserFunc: func(*cognitoidentityprovider.AdminCreateUserInput) (*cognitoidentityprovider.AdminCreateUserOutput, error) {
			return &cognitoidentityprovider.AdminCreateUserOutput{
				User: &types.UserType{
					Attributes: []types.AttributeType{
						{Name: aws.String("sub"), Value: aws.String("sub-123")},
					},
				},
			}, nil
		},
		setPasswordFunc: func(*cognitoidentityprovider.AdminSetUserPasswordInput) (*cognitoidentityprovider.AdminSetUserPasswordOutput, error) {
			return nil, setPasswordErr
		},
	}

	c := &CognitoClient{client: m, UserPool: "pool-1"}
	user, err := c.CreateUser(context.Background(), "alice@example.com", "Passw0rd!", "Alice", "Smith")
	if user != nil {
		t.Fatalf("expected nil user, got %+v", user)
	}
	if !errors.Is(err, setPasswordErr) {
		t.Fatalf("expected underlying set password error, got %v", err)
	}
	if len(m.deleteUserCalls) != 1 {
		t.Fatalf("expected cleanup delete for created user, got %v", m.deleteUserCalls)
	}
	if len(m.setPasswordCalls) != 1 {
		t.Fatal("expected AdminSetUserPassword to have been attempted")
	}
	if m.deleteUserCalls[0] != aws.ToString(m.setPasswordCalls[0].Username) {
		t.Fatalf("expected cleanup delete for the same username that failed, got %q vs %q",
			m.deleteUserCalls[0], aws.ToString(m.setPasswordCalls[0].Username))
	}
}

func TestCognitoClient_CreateUser_NoSubAttribute_CleansUpUser(t *testing.T) {
	m := &mockCognitoAPI{
		createUserFunc: func(*cognitoidentityprovider.AdminCreateUserInput) (*cognitoidentityprovider.AdminCreateUserOutput, error) {
			return &cognitoidentityprovider.AdminCreateUserOutput{
				User: &types.UserType{
					Attributes: []types.AttributeType{
						{Name: aws.String("email"), Value: aws.String("alice@example.com")},
					},
				},
			}, nil
		},
		setPasswordFunc: func(*cognitoidentityprovider.AdminSetUserPasswordInput) (*cognitoidentityprovider.AdminSetUserPasswordOutput, error) {
			return &cognitoidentityprovider.AdminSetUserPasswordOutput{}, nil
		},
	}

	c := &CognitoClient{client: m, UserPool: "pool-1"}
	user, err := c.CreateUser(context.Background(), "alice@example.com", "Passw0rd!", "Alice", "Smith")
	if user != nil {
		t.Fatalf("expected nil user, got %+v", user)
	}
	if err == nil || !strings.Contains(err.Error(), "COGNITO:NO_SUB") {
		t.Fatalf("expected COGNITO:NO_SUB error, got %v", err)
	}
	if len(m.deleteUserCalls) != 1 {
		t.Fatalf("expected cleanup delete when sub attribute is missing, got %v", m.deleteUserCalls)
	}
}

func TestCognitoClient_AddToGroup_Success(t *testing.T) {
	m := &mockCognitoAPI{
		addToGroupFunc: func(in *cognitoidentityprovider.AdminAddUserToGroupInput) (*cognitoidentityprovider.AdminAddUserToGroupOutput, error) {
			if aws.ToString(in.UserPoolId) != "pool-1" {
				t.Fatalf("expected UserPoolId pool-1, got %q", aws.ToString(in.UserPoolId))
			}
			if aws.ToString(in.Username) != "alice" {
				t.Fatalf("expected username alice, got %q", aws.ToString(in.Username))
			}
			if aws.ToString(in.GroupName) != "DOCTOR" {
				t.Fatalf("expected group name upper-cased to DOCTOR, got %q", aws.ToString(in.GroupName))
			}
			return &cognitoidentityprovider.AdminAddUserToGroupOutput{}, nil
		},
	}

	c := &CognitoClient{client: m, UserPool: "pool-1"}
	if err := c.AddToGroup(context.Background(), "alice", "doctor"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCognitoClient_AddToGroup_Error(t *testing.T) {
	boom := errors.New("boom")
	m := &mockCognitoAPI{
		addToGroupFunc: func(*cognitoidentityprovider.AdminAddUserToGroupInput) (*cognitoidentityprovider.AdminAddUserToGroupOutput, error) {
			return nil, boom
		},
	}

	c := &CognitoClient{client: m, UserPool: "pool-1"}
	err := c.AddToGroup(context.Background(), "alice", "doctor")
	if !errors.Is(err, boom) {
		t.Fatalf("expected passthrough error, got %v", err)
	}
}

func TestCognitoClient_DeleteUser(t *testing.T) {
	m := &mockCognitoAPI{}
	c := &CognitoClient{client: m, UserPool: "pool-1"}

	if err := c.DeleteUser(context.Background(), "alice"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.deleteUserCalls) != 1 || m.deleteUserCalls[0] != "alice" {
		t.Fatalf("expected delete call for alice, got %v", m.deleteUserCalls)
	}
}

func TestCognitoClient_DeleteUser_Error(t *testing.T) {
	boom := errors.New("boom")
	m := &mockCognitoAPI{
		deleteUserFunc: func(*cognitoidentityprovider.AdminDeleteUserInput) (*cognitoidentityprovider.AdminDeleteUserOutput, error) {
			return nil, boom
		},
	}

	c := &CognitoClient{client: m, UserPool: "pool-1"}
	err := c.DeleteUser(context.Background(), "alice")
	if !errors.Is(err, boom) {
		t.Fatalf("expected passthrough error, got %v", err)
	}
}

func TestTranslate(t *testing.T) {
	genericErr := errors.New("generic failure")

	tests := []struct {
		name     string
		in       error
		wantNil  bool
		wantCode string
		wantSame error
	}{
		{name: "nil error", in: nil, wantNil: true},
		{name: "generic error passes through", in: genericErr, wantSame: genericErr},
		{name: "invalid password", in: &types.InvalidPasswordException{Message: aws.String("bad")}, wantCode: "BAD_REQUEST"},
		{name: "username exists", in: &types.UsernameExistsException{Message: aws.String("taken")}, wantCode: "RESOURCE_CONFLICT"},
		{name: "alias exists", in: &types.AliasExistsException{Message: aws.String("taken")}, wantCode: "RESOURCE_CONFLICT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := translate(tt.in)
			if tt.wantNil {
				if got != nil {
					t.Fatalf("expected nil, got %v", got)
				}
				return
			}
			if tt.wantSame != nil {
				if !errors.Is(got, tt.wantSame) {
					t.Fatalf("expected passthrough of %v, got %v", tt.wantSame, got)
				}
				return
			}
			var appErr *apperr.AppError
			if !errors.As(got, &appErr) || appErr.Code != tt.wantCode {
				t.Fatalf("expected app error with code %q, got %v", tt.wantCode, got)
			}
		})
	}
}
