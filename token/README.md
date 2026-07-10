# 2Chi Go Token

Opaque token generation and HMAC-SHA256 hashing for 2Chi projects. Use for refresh tokens, invites, and API keys — not for user passwords (see `2chi-go-password` for Argon2id).

```go
import yca_token "github.com/yca-software/yca-go-core/token"
```

## Generate


| Symbol                                  | Description                           |
| --------------------------------------- | ------------------------------------- |
| `TokenLength`                           | Random bytes before encoding (32)     |
| `GenerateOpaqueToken() (string, error)` | URL-safe base64 token without padding |


Each call reads 32 cryptographically random bytes and returns a unique token.

## Hash


| Symbol                                  | Description                                  |
| --------------------------------------- | -------------------------------------------- |
| `NewHasher(pepper string) *Hasher`      | Create a hasher; panics if `pepper` is empty |
| `Hash(token string) string`             | HMAC-SHA256 hex digest for DB storage        |
| `Verify(storedHash, token string) bool` | Constant-time verification                   |


Pass `pepper` from app config (e.g. `TOKEN_HASH_PEPPER`). Store the hash, never the raw token.

## Example

```go
// Issue a token to the client once
token, err := yca_token.GenerateOpaqueToken()
if err != nil {
    return err
}

// Store hash in the database
hasher := yca_token.NewHasher(cfg.TokenHashPepper)
stored := hasher.Hash(token)

// Later, verify the presented token
if hasher.Verify(stored, presentedToken) {
    // valid
}
```

