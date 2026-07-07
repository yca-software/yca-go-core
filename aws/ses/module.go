package chi_aws_ses

import (
	"context"
	"fmt"

	aws_sdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

type Config struct {
	Region   string
	Endpoint string

	FromEmail string
	FromName  string
}

func (c Config) Enabled() bool {
	return c.Region != "" || c.Endpoint != "" || c.FromEmail != ""
}

// SES sends email via Amazon SES.
type SES interface {
	Send(ctx context.Context, msg SESEmailDataPayload) error
}

type sesClient struct {
	client    *sesv2.Client
	fromEmail string
	fromName  string
}

func NewSESClient(awsCfg aws_sdk.Config, cfg Config) SES {
	client := sesv2.NewFromConfig(awsCfg)

	return &sesClient{
		client:    client,
		fromEmail: cfg.FromEmail,
		fromName:  cfg.FromName,
	}
}

func formatFromAddress(fromEmail, fromName string) string {
	if fromName != "" {
		return fmt.Sprintf("%s <%s>", fromName, fromEmail)
	}
	return fromEmail
}

func (c *sesClient) Send(ctx context.Context, msg SESEmailDataPayload) error {
	from := formatFromAddress(c.fromEmail, c.fromName)

	_, err := c.client.SendEmail(ctx, &sesv2.SendEmailInput{
		FromEmailAddress: aws_sdk.String(from),
		Destination: &types.Destination{
			ToAddresses: []string{msg.To},
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{Data: aws_sdk.String(msg.Subject)},
				Body: &types.Body{
					Html: &types.Content{Data: aws_sdk.String(msg.HTML)},
				},
			},
		},
	})

	return err
}
