package yca_aws_s3

import (
	"net/url"
	"strings"
)

// ObjectKeyFromURL returns the S3 object key from a public URL produced by [S3.PutObject] / [ObjectPublicURL]
// (path-style with bucket in the path, or virtual-hosted *.amazonaws.com).
func ObjectKeyFromURL(raw string, bucket string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false
	}
	u, err := url.Parse(raw)
	if err != nil || u.Path == "" || u.Path == "/" {
		return "", false
	}
	path := strings.TrimPrefix(u.Path, "/")
	prefix := bucket + "/"
	if strings.HasPrefix(path, prefix) {
		key := strings.TrimPrefix(path, prefix)
		return key, key != ""
	}
	host := u.Hostname()
	if strings.HasPrefix(host, bucket+".s3.") && strings.HasSuffix(host, ".amazonaws.com") {
		return path, path != ""
	}
	return "", false
}
