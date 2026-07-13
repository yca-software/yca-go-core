# 2Chi Go AWS

Composable AWS integrations for 2Chi services.

```go
import yca_aws "github.com/yca-software/yca-go-core/aws"
```

Subpackages expose interfaces, config structs, and mocks:

```go
import (
    yca_aws_iot "github.com/yca-software/yca-go-core/aws/iot"
    yca_aws_s3 "github.com/yca-software/yca-go-core/aws/s3"
    yca_aws_ses "github.com/yca-software/yca-go-core/aws/ses"
    yca_aws_sqs "github.com/yca-software/yca-go-core/aws/sqs"
)
```

## Setup

```go
mod, err := yca_aws.New(ctx, yca_aws.Config{
    Region:   "eu-central-1",
    Endpoint: "http://localhost:4566", // optional (for LocalStack/dev)
    SES: &yca_aws_ses.Config{
        FromEmail: "noreply@example.com",
        FromName:  "2Chi",
    },
    S3: &yca_aws_s3.Config{},
    SQS: &yca_aws_sqs.Config{},
    IoT: &yca_aws_iot.Config{Region: "eu-central-1"},
})
if err != nil {
    return err
}
```

Only enabled sub-configs are wired. Disabled services stay `nil` on `Module`.

## Services

- `ses`:
  - `SES.Send(ctx, SESEmailDataPayload)`
- `s3`:
  - `S3.PutObject`, `S3.GetObject`, `S3.DeleteObject`
  - Helpers: `ObjectPublicURL`, `ObjectKeyFromURL`
- `sqs`:
  - `SQS.SendMessage`, `SQS.ReceiveMessages`, `SQS.DeleteMessage`, `SQS.ChangeMessageVisibility`
- `iot`:
  - `IoT.GetThingShadow`, `IoT.UpdateThingShadow`
  - `IoT.GetThingConnectivityData` (requires **fleet indexing** enabled on the AWS IoT account)
  - Data endpoint resolved at runtime via `iot:DescribeEndpoint` (region required on `yca_aws_iot.Config`)

Each subpackage includes a testify mock (`MockSES`, `MockS3`, `MockSQS`, `MockIoT`) for unit tests.

## Testing

```bash
go test -race -count=1 ./...
```

The package tests are unit-level and do not make live AWS calls.
