# 2Chi Go Google

Google integrations for 2Chi projects: OAuth sign-in and Places / location helpers.

```go
import yca_google "github.com/yca-software/yca-go-core/google"
```

Subpackages hold service interfaces, types, and mocks:

```go
import (
    yca_google_maps "github.com/yca-software/yca-go-core/google/maps"
    yca_google_oauth "github.com/yca-software/yca-go-core/google/oauth"
)
```

## Setup

```go
mod := yca_google.New(yca_google.Config{
    OAuth: &yca_google_oauth.OAuthConfig{
        ClientID:     cfg.GoogleClientID,
        ClientSecret: cfg.GoogleClientSecret,
        RedirectURL:  cfg.GoogleRedirectURL,
    },
    Maps: &yca_google_maps.MapsConfig{
        APIKey: cfg.GoogleMapsAPIKey,
    },
    HTTPClient: httpClient,
    Logger:     logger,
})
```

Only enabled capabilities are wired; disabled config leaves the corresponding `Module` field nil.

## OAuth (`/oauth`)


| Symbol        | Description                     |
| ------------- | ------------------------------- |
| `OAuthConfig` | Client ID, secret, redirect URL |
| `OAuth`       | `GetUserInfo(ctx, code)`        |
| `MockOAuth`   | Testify mock                    |


## Maps (`/maps`)


| Symbol       | Description                                            |
| ------------ | ------------------------------------------------------ |
| `MapsConfig` | Places API key                                         |
| `Maps`       | Autocomplete, place details, normalized `LocationData` |
| `MockMaps`   | Testify mock                                           |


`LocationData` includes address fields, coordinates, and timezone (from lat/lng).

## Example

```go
user, err := mod.OAuth.GetUserInfo(ctx, authCode)

predictions, err := mod.Maps.AutocompleteLocation(ctx, "Paris")
location, err := mod.Maps.GetLocationData(ctx, placeID)
```

## Tests

```bash
go test -race -count=1 ./...
```

HTTP clients are mocked in tests; no live Google API calls.