package events

import (
	"context"
	"time"
)

type Metadata struct {
	At       time.Time `json:"at"`
	UserId   string    `json:"user_id"`
	ObjectId string    `json:"object_id"`
}

type Message struct {
	Topic    Topic     `json:"topic"`
	Metadata *Metadata `json:"metadata"`
}

type MessageParams struct {
	UserId   string
	ObjectId string
	Topic    Topic
}

type EventQueue interface {
	Publish(context.Context, map[string]any) error
	Close() error
}
