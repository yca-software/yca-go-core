package chi_aws_s3

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type S3HelpersSuite struct {
	suite.Suite
}

func TestS3HelpersSuite(t *testing.T) {
	suite.Run(t, new(S3HelpersSuite))
}

func (s *S3HelpersSuite) TestObjectKeyFromURL_PathStyle() {
	key, ok := ObjectKeyFromURL("https://s3.eu-central-1.amazonaws.com/my-bucket/path/to/file.pdf", "my-bucket")
	s.True(ok)
	s.Equal("path/to/file.pdf", key)
}

func (s *S3HelpersSuite) TestObjectKeyFromURL_VirtualHostedStyle() {
	key, ok := ObjectKeyFromURL("https://my-bucket.s3.eu-central-1.amazonaws.com/path/to/file.pdf", "my-bucket")
	s.True(ok)
	s.Equal("path/to/file.pdf", key)
}

func (s *S3HelpersSuite) TestObjectKeyFromURL_Invalid() {
	key, ok := ObjectKeyFromURL("not a url", "my-bucket")
	s.False(ok)
	s.Equal("", key)
}
