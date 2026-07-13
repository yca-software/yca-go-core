package yca_aws_iot

import "time"

// ThingConnectivityData is the connectivity snapshot from AWS IoT Core fleet indexing.
type ThingConnectivityData struct {
	Connected        bool
	Timestamp        time.Time
	DisconnectReason string
	ClientID         string
}
