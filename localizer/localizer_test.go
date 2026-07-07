package chi_localizer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	chi_localizer "github.com/yca-software/yca-go-core/localizer"
)

type LocalizerSuite struct {
	suite.Suite
	localesDir string
	loc        chi_localizer.Localizer
}

func TestLocalizerSuite(t *testing.T) {
	suite.Run(t, new(LocalizerSuite))
}

func (s *LocalizerSuite) SetupSuite() {
	wd, err := os.Getwd()
	s.Require().NoError(err)
	s.localesDir = filepath.Join(wd, "testdata", "locales")
	s.loc = chi_localizer.New([]string{"en", "es", "tr"}, "en", s.localesDir)
	s.Require().NotNil(s.loc)
}

func (s *LocalizerSuite) TestNewLocalizer() {
	loc := chi_localizer.New([]string{"en", "es"}, "en", s.localesDir)
	s.NotNil(loc)
}

func (s *LocalizerSuite) TestTranslate_supportedLanguage() {
	msg := s.loc.Translate("en", "welcome.message", map[string]any{"Name": "John"})
	s.Equal("Hello John", msg)
}

func (s *LocalizerSuite) TestTranslate_unsupportedLanguageFallsBackToDefault() {
	msg := s.loc.Translate("fr", "welcome.message", map[string]any{"Name": "John"})
	s.Equal("Hello John", msg)
}

func (s *LocalizerSuite) TestTranslate_missingKeyReturnsKey() {
	msg := s.loc.Translate("en", "missing.key", nil)
	s.Equal("missing.key", msg)
}

func (s *LocalizerSuite) TestTranslate_missingKeyInPreferredLanguageDoesNotFallBack() {
	// welcome.message exists in en.json only; es bundle has no welcome namespace.
	msg := s.loc.Translate("es", "welcome.message", map[string]any{"Name": "John"})
	s.Equal("welcome.message", msg)
}

func (s *LocalizerSuite) TestTranslate_spanishErrorMessage() {
	msg := s.loc.Translate("es", "error.400.0001", nil)
	s.Equal("Solicitud inválida", msg)
}
