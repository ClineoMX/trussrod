package iam

import (
	"context"
)

type User struct {
	Sub      string
	Username string
}

type Client interface {
	CreateUser(ctx context.Context, username, password string) (*User, error)
	AddUserToGroup(ctx context.Context, username, group string) error
	DeleteUser(ctx context.Context, username string) error
}
