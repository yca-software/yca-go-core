package chi_archive_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	chi_archive "github.com/yca-software/yca-go-core/archive"
	chi_error "github.com/yca-software/yca-go-core/error"
)

type ArchiveTestSuite struct {
	suite.Suite
}

func TestArchiveTestSuite(t *testing.T) {
	suite.Run(t, new(ArchiveTestSuite))
}

func (s *ArchiveTestSuite) TestConstants() {
	s.Equal(chi_archive.ArchiveFilter("active"), chi_archive.ArchiveFilterActive)
	s.Equal(chi_archive.ArchiveFilter("archived"), chi_archive.ArchiveFilterArchived)
	s.Equal(chi_archive.ArchiveFilterActive, chi_archive.DefaultArchiveFilter)
	s.Equal(30*24*time.Hour, chi_archive.ArchivedDataRetentionPeriod)
}

func (s *ArchiveTestSuite) TestNormalizeArchiveFilter_EmptyDefaultsToActive() {
	filter, err := chi_archive.NormalizeArchiveFilter("")
	s.NoError(err)
	s.Equal(chi_archive.ArchiveFilterActive, filter)
}

func (s *ArchiveTestSuite) TestNormalizeArchiveFilter_Active() {
	filter, err := chi_archive.NormalizeArchiveFilter(chi_archive.ArchiveFilterActive)
	s.NoError(err)
	s.Equal(chi_archive.ArchiveFilterActive, filter)
}

func (s *ArchiveTestSuite) TestNormalizeArchiveFilter_Archived() {
	filter, err := chi_archive.NormalizeArchiveFilter(chi_archive.ArchiveFilterArchived)
	s.NoError(err)
	s.Equal(chi_archive.ArchiveFilterArchived, filter)
}

func (s *ArchiveTestSuite) TestNormalizeArchiveFilter_Invalid() {
	filter, err := chi_archive.NormalizeArchiveFilter(chi_archive.ArchiveFilter("nope"))
	s.Error(err)
	s.Empty(filter)
	s.Contains(err.Error(), `invalid archive filter "nope"`)
}

func (s *ArchiveTestSuite) TestParseArchiveFilterQuery_EmptyDefaultsToActive() {
	filter, err := chi_archive.ParseArchiveFilterQuery("")
	s.NoError(err)
	s.Equal(chi_archive.ArchiveFilterActive, filter)
}

func (s *ArchiveTestSuite) TestParseArchiveFilterQuery_Active() {
	filter, err := chi_archive.ParseArchiveFilterQuery("active")
	s.NoError(err)
	s.Equal(chi_archive.ArchiveFilterActive, filter)
}

func (s *ArchiveTestSuite) TestParseArchiveFilterQuery_Archived() {
	filter, err := chi_archive.ParseArchiveFilterQuery("archived")
	s.NoError(err)
	s.Equal(chi_archive.ArchiveFilterArchived, filter)
}

func (s *ArchiveTestSuite) TestParseArchiveFilterQuery_InvalidReturnsBadRequest() {
	filter, err := chi_archive.ParseArchiveFilterQuery("nope")
	s.Empty(filter)
	s.Error(err)

	apiErr, ok := chi_error.AsError(err)
	s.True(ok)
	s.Equal(http.StatusBadRequest, apiErr.StatusCode)
	s.Equal("InvalidArchiveFilter", apiErr.ErrorCode)
}
