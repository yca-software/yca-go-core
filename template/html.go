package chi_template

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type HTML struct {
	templatesPath string
	cache         sync.Map // template name -> *template.Template
}

func NewHTML(templatesPath string) *HTML {
	return &HTML{templatesPath: templatesPath}
}

func (r *HTML) Render(name string, data any) (string, error) {
	if strings.Contains(name, "..") || strings.ContainsAny(name, `/\`) {
		return "", fmt.Errorf("invalid template name: %q", name)
	}

	path := filepath.Join(r.templatesPath, name+".html")
	rel, err := filepath.Rel(r.templatesPath, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("invalid template name: %q", name)
	}

	templateName := fmt.Sprintf("%s.html", name)
	tmpl, err := r.loadTemplate(name, templateName, path)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, templateName, data); err != nil {
		return "", fmt.Errorf("execute template %q: %w", name, err)
	}

	return buf.String(), nil
}

func (r *HTML) loadTemplate(name, templateName, path string) (*template.Template, error) {
	if cached, ok := r.cache.Load(name); ok {
		return cached.(*template.Template), nil
	}

	htmlData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read template: %w", err)
	}

	tmpl, err := template.New(templateName).Parse(string(htmlData))
	if err != nil {
		return nil, fmt.Errorf("parse template %q: %w", name, err)
	}

	actual, _ := r.cache.LoadOrStore(name, tmpl)
	return actual.(*template.Template), nil
}
