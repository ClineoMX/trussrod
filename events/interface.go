package events

import (
	"context"
)

type EventQueue interface {
	Publish(ctx context.Context, message map[string]any) error
	Close() error
}
