# 2Chi Go Password

Argon2id password hashing and verification for 2Chi projects.

```go
import yca_password "github.com/yca-software/yca-go-core/password"
```

## API

| Function | Description |
| --- | --- |
| `Hash(password string) (string, error)` | Hash a plaintext password and return a PHC-style Argon2id string |
| `Compare(password, encodedHash string) bool` | Verify a password against an encoded hash |

`Compare` returns `false` for invalid or malformed hashes (no error is returned). Hash comparison is constant-time to reduce timing side channels.

## Hash format

Hashes use the PHC string format:

```text
$argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>
```

Salt and hash segments are base64-encoded (raw, no padding).

## Parameters

| Parameter | Value |
| --- | --- |
| Algorithm | Argon2id |
| Memory | 64 MB (`65536` KiB) |
| Iterations | 3 |
| Parallelism | 4 |
| Key length | 32 bytes |
| Salt length | 16 bytes (128 bits) |

Each call to `Hash` generates a fresh random salt, so the same password produces different encoded strings on every hash.

## Example

```go
hash, err := yca_password.Hash("my-secure-password")
if err != nil {
    return err
}

// Store hash in your database, then verify on login:
if yca_password.Compare("my-secure-password", hash) {
    // password matches
}
```
