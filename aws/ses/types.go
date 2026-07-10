package yca_aws_ses

// SESEmailDataPayload is a single outbound email sent via SES.
type SESEmailDataPayload struct {
	To      string
	Subject string
	HTML    string
}
