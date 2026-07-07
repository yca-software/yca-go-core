package chi_aws_sqs

// QueueMessage is a received SQS message.
type QueueMessage struct {
	ID            string
	Body          string
	ReceiptHandle string
	Attributes    map[string]string
}

// ReceiveOptions configures ReceiveMessages.
type ReceiveOptions struct {
	QueueURL              string
	MaxMessages           int32
	WaitTimeSeconds       int32
	VisibilityTimeout     int32
	MessageAttributeNames []string
}
