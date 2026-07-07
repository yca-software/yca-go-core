package chi_aws_s3

import (
	"io"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// PutObjectOptions configures an S3 object upload.
type PutObjectOptions struct {
	Bucket      string
	Key         string
	Body        io.Reader
	ContentType string
	ACL         types.ObjectCannedACL
	// SkipObjectACL omits a canned ACL on the object (use bucket policy for public reads).
	SkipObjectACL      bool
	CacheControl       string
	ContentDisposition string
}

// GetObjectOptions configures an S3 object download.
type GetObjectOptions struct {
	Bucket string
	Key    string
}
