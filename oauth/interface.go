package oauth

import (
	"context"
)

type LoginOutput struct {
	Access   string `json:"access"`
	Identity string `json:"id"`
	Refresh  string `json:"refresh"`
}

type Client interface {
	Login(ctx context.Context, username, password string) (*LoginOutput, error)
	RequestResetPassword(ctx context.Context, email string) error
	ConfirmResetPassword(ctx context.Context, email, code, newPassword string) error
	ConfirmUserSignup(ctx context.Context, email, code string) error
	// CreateUser(ctx context.Context, username string) error
}
