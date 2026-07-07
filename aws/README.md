# 2Chi Go AWS

Composable AWS integrations for 2Chi services.

```go
import chi_aws "github.com/yca-software/yca-go-core/aws"
```

Subpackages expose interfaces, config structs, and mocks:

```go
import (
    chi_aws_iot "github.com/yca-software/yca-go-core/aws/iot"
    chi_aws_s3 "github.com/yca-software/yca-go-core/aws/s3"
    chi_aws_ses "github.com/yca-software/yca-go-core/aws/ses"
    chi_aws_sqs "github.com/yca-software/yca-go-core/aws/sqs"
)
```

## Setup

```go
mod, err := chi_aws.New(ctx, chi_aws.Config{
    Region:   "eu-central-1",
    Endpoint: "http://localhost:4566", // optional (for LocalStack/dev)
    SES: &chi_aws_ses.Config{
        FromEmail: "noreply@example.com",
        FromName:  "2Chi",
    },
    S3: &chi_aws_s3.Config{},
    SQS: &chi_aws_sqs.Config{},
    IoT: &chi_aws_iot.Config{Region: "eu-central-1"},
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
  - Data endpoint resolved at runtime via `iot:DescribeEndpoint` (region required on `chi_aws_iot.Config`)

Each subpackage includes a testify mock (`MockSES`, `MockS3`, `MockSQS`, `MockIoT`) for unit tests.

## Testing

```bash
go test -race -count=1 ./...
```

The package tests are unit-level and do not make live AWS calls.
