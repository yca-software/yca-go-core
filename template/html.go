package yca_template

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

	pagePath := filepath.Join(r.templatesPath, name+".html")
	rel, err := filepath.Rel(r.templatesPath, pagePath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("invalid template name: %q", name)
	}

	templateName := fmt.Sprintf("%s.html", name)
	tmpl, err := r.loadTemplate(name, templateName, pagePath)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, templateName, data); err != nil {
		return "", fmt.Errorf("execute template %q: %w", name, err)
	}

	return buf.String(), nil
}

func (r *HTML) loadTemplate(name, templateName, pagePath string) (*template.Template, error) {
	if cached, ok := r.cache.Load(name); ok {
		return cached.(*template.Template), nil
	}

	layoutPath := filepath.Join(r.templatesPath, "_layout.html")
	layoutRel, layoutErr := filepath.Rel(r.templatesPath, layoutPath)
	hasLayout := layoutErr == nil && !strings.HasPrefix(layoutRel, "..")
	if hasLayout {
		if _, err := os.Stat(layoutPath); err != nil {
			hasLayout = false
		}
	}

	var tmpl *template.Template
	var err error
	if hasLayout {
		if _, err = os.Stat(pagePath); err != nil {
			return nil, fmt.Errorf("read template: %w", err)
		}
		tmpl, err = template.New(templateName).ParseFiles(layoutPath, pagePath)
		if err != nil {
			return nil, fmt.Errorf("parse template %q: %w", name, err)
		}
	} else {
		var htmlData []byte
		htmlData, err = os.ReadFile(pagePath)
		if err != nil {
			return nil, fmt.Errorf("read template: %w", err)
		}
		tmpl, err = template.New(templateName).Parse(string(htmlData))
		if err != nil {
			return nil, fmt.Errorf("parse template %q: %w", name, err)
		}
	}

	actual, _ := r.cache.LoadOrStore(name, tmpl)
	return actual.(*template.Template), nil
}
