package yca_aws_s3

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/suite"
)

type S3ModuleSuite struct {
	suite.Suite
}

func TestS3ModuleSuite(t *testing.T) {
	suite.Run(t, new(S3ModuleSuite))
}

func (s *S3ModuleSuite) TestConfigEnabled() {
	s.False((Config{}).Enabled())
	s.True((Config{Region: "eu-central-1"}).Enabled())
	s.True((Config{Endpoint: "http://localhost:4566"}).Enabled())
}

func (s *S3ModuleSuite) TestObjectPublicURL() {
	url := ObjectPublicURL("eu-central-1", "my-bucket", "avatars/user-1.png")
	s.Equal("https://my-bucket.s3.eu-central-1.amazonaws.com/avatars/user-1.png", url)
}

func (s *S3ModuleSuite) TestContentTypeForObject_PrefersExplicitType() {
	ct := contentTypeForObject("file.pdf", "image/png")
	s.Equal("image/png", ct)
}

func (s *S3ModuleSuite) TestContentTypeForObject_DefaultsByExtension() {
	s.Equal("application/pdf", contentTypeForObject("doc.pdf", ""))
	s.Equal("image/jpeg", contentTypeForObject("photo.jpg", "application/octet-stream"))
	s.Equal("application/octet-stream", contentTypeForObject("blob.unknown", ""))
}

func (s *S3ModuleSuite) TestPutObjectOptions_DefaultACLZeroValueIsValid() {
	var opts PutObjectOptions
	s.Equal(types.ObjectCannedACL(""), opts.ACL)
}

func (s *S3ModuleSuite) TestPresignGetObject_RequiresBucketAndKey() {
	client := &s3Client{}
	_, err := client.PresignGetObject(context.Background(), PresignGetObjectOptions{})
	s.Error(err)
	s.Contains(err.Error(), "bucket and key are required")

	_, err = client.PresignGetObject(context.Background(), PresignGetObjectOptions{Bucket: "b"})
	s.Error(err)
	s.Contains(err.Error(), "bucket and key are required")

	_, err = client.PresignGetObject(context.Background(), PresignGetObjectOptions{Key: "k"})
	s.Error(err)
	s.Contains(err.Error(), "bucket and key are required")
}
