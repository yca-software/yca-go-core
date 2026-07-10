# 2Chi Go Localizer

JSON-based i18n for 2Chi projects, built on [go-i18n](https://github.com/nicksnyder/go-i18n).

```go
import yca_localizer "github.com/yca-software/yca-go-core/localizer"
```

## API

| Symbol | Description |
| --- | --- |
| `New(supportedLanguages, defaultLanguage, localesPath string) Localizer` | Load locale files and return a localizer |
| `Translate(lang, key string, data map[string]any) string` | Resolve a message for the given language |

`Translate` looks up `key` using dot notation (for example `welcome.message` maps to `welcome.message` in go-i18n). `data` supplies template variables for messages such as `Hello {{.Name}}`.

## Locale files

For each entry in `supportedLanguages`, `New` loads `{localesPath}/{lang}.json` when the file exists. Files use nested JSON objects; keys are joined with dots to form message IDs.

```json
{
  "welcome": {
    "message": "Hello {{.Name}}"
  },
  "error": {
    "400": {
      "0001": "Invalid request"
    }
  }
}
```

## Fallback behavior

| Case | Result |
| --- | --- |
| Unsupported `lang` | Uses `defaultLanguage` |
| Missing translation in a supported language | Returns the `key` unchanged |
| Missing locale file | Language is still registered; only keys present in loaded bundles resolve |

## Example

```go
loc := yca_localizer.New(
    []string{"en", "es", "tr"},
    "en",
    "/app/locales",
)

msg := loc.Translate("en", "welcome.message", map[string]any{"Name": "John"})
// "Hello John"

errMsg := loc.Translate("es", "error.400.0001", nil)
// "Solicitud inválida"
```
