package yca_aws_sqs

import (
	"context"
	"fmt"

	aws_sdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type Config struct {
	Region   string
	Endpoint string
}

func (c Config) Enabled() bool {
	return c.Region != "" || c.Endpoint != ""
}

type SQS interface {
	SendMessage(ctx context.Context, queueURL string, body []byte, attributes map[string]string) error
	ReceiveMessages(ctx context.Context, opts ReceiveOptions) ([]QueueMessage, error)
	DeleteMessage(ctx context.Context, queueURL, receiptHandle string) error
	ChangeMessageVisibility(ctx context.Context, queueURL, receiptHandle string, visibilityTimeout int32) error
}

type sqsClient struct {
	client *sqs.Client
}

func NewSQSClient(cfg aws_sdk.Config) SQS {
	client := sqs.NewFromConfig(cfg)

	return &sqsClient{
		client: client,
	}
}

func (c *sqsClient) SendMessage(ctx context.Context, queueURL string, body []byte, attributes map[string]string) error {
	if queueURL == "" {
		return fmt.Errorf("aws/sqs: queue URL is required")
	}

	input := &sqs.SendMessageInput{
		QueueUrl:    &queueURL,
		MessageBody: aws_sdk.String(string(body)),
	}

	if len(attributes) > 0 {
		input.MessageAttributes = make(map[string]types.MessageAttributeValue, len(attributes))
		for k, v := range attributes {
			input.MessageAttributes[k] = types.MessageAttributeValue{
				DataType:    aws_sdk.String("String"),
				StringValue: aws_sdk.String(v),
			}
		}
	}

	if _, err := c.client.SendMessage(ctx, input); err != nil {
		return fmt.Errorf("aws/sqs: send message: %w", err)
	}

	return nil
}

func (c *sqsClient) ReceiveMessages(ctx context.Context, opts ReceiveOptions) ([]QueueMessage, error) {
	if opts.QueueURL == "" {
		return nil, fmt.Errorf("aws/sqs: queue URL is required")
	}

	max := opts.MaxMessages
	if max < 1 {
		max = 1
	}

	input := &sqs.ReceiveMessageInput{
		QueueUrl:            &opts.QueueURL,
		MaxNumberOfMessages: max,
		WaitTimeSeconds:     opts.WaitTimeSeconds,
	}

	if opts.VisibilityTimeout > 0 {
		input.VisibilityTimeout = opts.VisibilityTimeout
	}

	if len(opts.MessageAttributeNames) > 0 {
		input.MessageAttributeNames = opts.MessageAttributeNames
	}

	out, err := c.client.ReceiveMessage(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("aws/sqs: receive messages: %w", err)
	}

	msgs := make([]QueueMessage, 0, len(out.Messages))
	for _, m := range out.Messages {
		qm := QueueMessage{
			Attributes: make(map[string]string),
		}

		if m.MessageId != nil {
			qm.ID = *m.MessageId
		}

		if m.Body != nil {
			qm.Body = *m.Body
		}

		if m.ReceiptHandle != nil {
			qm.ReceiptHandle = *m.ReceiptHandle
		}

		for k, v := range m.MessageAttributes {
			if v.StringValue != nil {
				qm.Attributes[k] = *v.StringValue
			}
		}
		msgs = append(msgs, qm)
	}

	return msgs, nil
}

func (c *sqsClient) DeleteMessage(ctx context.Context, queueURL, receiptHandle string) error {
	if queueURL == "" || receiptHandle == "" {
		return fmt.Errorf("aws/sqs: queue URL and receipt handle are required")
	}

	if _, err := c.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      &queueURL,
		ReceiptHandle: &receiptHandle,
	}); err != nil {
		return fmt.Errorf("aws/sqs: delete message: %w", err)
	}

	return nil
}

func (c *sqsClient) ChangeMessageVisibility(ctx context.Context, queueURL, receiptHandle string, visibilityTimeout int32) error {
	if queueURL == "" || receiptHandle == "" {
		return fmt.Errorf("aws/sqs: queue URL and receipt handle are required")
	}

	if _, err := c.client.ChangeMessageVisibility(ctx, &sqs.ChangeMessageVisibilityInput{
		QueueUrl:          &queueURL,
		ReceiptHandle:     &receiptHandle,
		VisibilityTimeout: visibilityTimeout,
	}); err != nil {
		return fmt.Errorf("aws/sqs: change visibility: %w", err)
	}

	return nil
}
