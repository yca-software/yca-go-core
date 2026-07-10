# 2Chi Go Template

HTML template rendering for 2Chi projects (transactional email, etc.), built on Go's `html/template`.

```go
import yca_template "github.com/yca-software/yca-go-core/template"
```

## API

| Symbol | Description |
| --- | --- |
| `NewHTML(templatesPath string) *HTML` | Create a renderer for templates under `templatesPath` |
| `Render(name string, data any) (string, error)` | Render `{name}.html` with the given data |

Templates are loaded from `{templatesPath}/{name}.html`, parsed once, and cached in memory.

## Security

`Render` rejects path traversal — template names must not contain `..`, `/`, or `\`.

## Example

```go
renderer := yca_template.NewHTML("/app/templates/email")

body, err := renderer.Render("welcome", map[string]string{
    "Name": "Ada",
})
if err != nil {
    return err
}

// body: rendered HTML from welcome.html
```

### Sample template (`welcome.html`)

```html
<p>Hello, {{.Name}}!</p>
```
