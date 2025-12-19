package events

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type SQS struct {
	URN  string
	conn *sqs.Client
}

func NewSQSClient(urn, region string) (*SQS, error) {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}
	client := sqs.NewFromConfig(cfg)
	return &SQS{
		conn: client,
		URN:  urn,
	}, nil
}

func (s *SQS) Close() error {
	return nil
}

func (s *SQS) Publish(ctx context.Context, message map[string]any) error {
	marshalled, err := json.Marshal(message)
	if err != nil {
		return err
	}
	_, err = s.conn.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    &s.URN,
		MessageBody: aws.String(string(marshalled)),
	})
	return err
}
