package chi_template_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	chi_template "github.com/yca-software/yca-go-core/template"
)

type HTMLSuite struct {
	suite.Suite
	dir string
}

func TestHTMLSuite(t *testing.T) {
	suite.Run(t, new(HTMLSuite))
}

func (s *HTMLSuite) SetupTest() {
	dir, err := os.MkdirTemp("", "go-core-template-*")
	s.Require().NoError(err)
	s.dir = dir
}

func (s *HTMLSuite) TearDownTest() {
	if s.dir != "" {
		_ = os.RemoveAll(s.dir)
	}
}

func (s *HTMLSuite) writeTemplate(name, content string) {
	path := filepath.Join(s.dir, name+".html")
	s.Require().NoError(os.WriteFile(path, []byte(content), 0o644))
}

func (s *HTMLSuite) TestRender_success() {
	s.writeTemplate("welcome", "<p>Hello, {{.Name}}!</p>")

	renderer := chi_template.NewHTML(s.dir)
	out, err := renderer.Render("welcome", map[string]string{"Name": "World"})
	s.Require().NoError(err)
	s.Equal("<p>Hello, World!</p>", out)
}

func (s *HTMLSuite) TestRender_rejectsPathTraversal() {
	renderer := chi_template.NewHTML(s.dir)

	_, err := renderer.Render("../secret", nil)
	s.Require().Error(err)
	s.Contains(err.Error(), "invalid template name")

	_, err = renderer.Render(`foo/bar`, nil)
	s.Require().Error(err)
	s.Contains(err.Error(), "invalid template name")
}

func (s *HTMLSuite) TestRender_cachesParsedTemplate() {
	s.writeTemplate("cached", "<p>{{.Value}}</p>")

	renderer := chi_template.NewHTML(s.dir)
	first, err := renderer.Render("cached", map[string]string{"Value": "one"})
	s.Require().NoError(err)
	s.Equal("<p>one</p>", first)

	s.Require().NoError(os.WriteFile(filepath.Join(s.dir, "cached.html"), []byte("<p>{{.Value}}-updated</p>"), 0o644))

	second, err := renderer.Render("cached", map[string]string{"Value": "two"})
	s.Require().NoError(err)
	s.Equal("<p>two</p>", second)
}
