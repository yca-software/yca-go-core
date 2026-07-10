package yca_aws_s3

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"

	aws_sdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type Config struct {
	Region   string
	Endpoint string
}

func (c Config) Enabled() bool {
	return c.Region != "" || c.Endpoint != ""
}

type S3 interface {
	PutObject(ctx context.Context, opts PutObjectOptions) (publicURL string, err error)
	GetObject(ctx context.Context, opts GetObjectOptions) (io.ReadCloser, error)
	DeleteObject(ctx context.Context, bucket, key string) error
}

type s3Client struct {
	region string
	client *s3.Client
}

func NewS3Client(cfg aws_sdk.Config) S3 {
	client := s3.NewFromConfig(cfg)

	return &s3Client{
		region: cfg.Region,
		client: client,
	}
}

func (c *s3Client) PutObject(ctx context.Context, opts PutObjectOptions) (string, error) {
	if opts.Bucket == "" || opts.Key == "" {
		return "", fmt.Errorf("aws/s3: bucket and key are required")
	}

	if opts.Body == nil {
		return "", fmt.Errorf("aws/s3: body is required")
	}

	acl := opts.ACL
	if !opts.SkipObjectACL {
		if acl == "" {
			acl = types.ObjectCannedACLPublicRead
		}
	}

	cacheControl := opts.CacheControl
	if cacheControl == "" {
		cacheControl = "no-cache"
	}

	input := &s3.PutObjectInput{
		Bucket:       aws_sdk.String(opts.Bucket),
		Key:          aws_sdk.String(opts.Key),
		Body:         opts.Body,
		CacheControl: aws_sdk.String(cacheControl),
		ContentType:  aws_sdk.String(contentTypeForObject(opts.Key, opts.ContentType)),
	}

	if !opts.SkipObjectACL {
		input.ACL = acl
	}

	if cd := opts.ContentDisposition; cd != "" {
		input.ContentDisposition = aws_sdk.String(cd)
	}

	if _, err := c.client.PutObject(ctx, input); err != nil {
		return "", fmt.Errorf("aws/s3: put object: %w", err)
	}

	return ObjectPublicURL(c.region, opts.Bucket, opts.Key), nil
}

func (c *s3Client) GetObject(ctx context.Context, opts GetObjectOptions) (io.ReadCloser, error) {
	if opts.Bucket == "" || opts.Key == "" {
		return nil, fmt.Errorf("aws/s3: bucket and key are required")
	}
	out, err := c.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws_sdk.String(opts.Bucket),
		Key:    aws_sdk.String(opts.Key),
	})
	if err != nil {
		return nil, fmt.Errorf("aws/s3: get object: %w", err)
	}
	return out.Body, nil
}

func (c *s3Client) DeleteObject(ctx context.Context, bucket, key string) error {
	if bucket == "" || key == "" {
		return fmt.Errorf("aws/s3: bucket and key are required")
	}

	if _, err := c.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws_sdk.String(bucket),
		Key:    aws_sdk.String(key),
	}); err != nil {
		return fmt.Errorf("aws/s3: delete object: %w", err)
	}

	return nil
}

// ObjectPublicURL builds a browser-accessible URL for an object key.
func ObjectPublicURL(region, bucket, objectKey string) string {
	u := url.URL{Scheme: "https", Host: fmt.Sprintf("%s.s3.%s.amazonaws.com", bucket, region)}
	return u.JoinPath(strings.Split(objectKey, "/")...).String()
}

// contentTypeForObject picks a Content-Type for PutObject so browsers can display
// objects inline instead of treating them as generic binary downloads.
func contentTypeForObject(objectKey, explicit string) string {
	if ct := strings.TrimSpace(explicit); ct != "" && ct != "application/octet-stream" {
		return ct
	}

	switch strings.ToLower(filepath.Ext(objectKey)) {
	case ".pdf":
		return "application/pdf"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".txt":
		return "text/plain; charset=utf-8"
	case ".csv":
		return "text/csv; charset=utf-8"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".ppt":
		return "application/vnd.ms-powerpoint"
	case ".pptx":
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	default:
		return "application/octet-stream"
	}
}
