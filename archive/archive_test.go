package yca_archive_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	yca_archive "github.com/yca-software/yca-go-core/archive"
	yca_error "github.com/yca-software/yca-go-core/error"
)

type ArchiveTestSuite struct {
	suite.Suite
}

func TestArchiveTestSuite(t *testing.T) {
	suite.Run(t, new(ArchiveTestSuite))
}

func (s *ArchiveTestSuite) TestConstants() {
	s.Equal(yca_archive.ArchiveFilter("active"), yca_archive.ArchiveFilterActive)
	s.Equal(yca_archive.ArchiveFilter("archived"), yca_archive.ArchiveFilterArchived)
	s.Equal(yca_archive.ArchiveFilterActive, yca_archive.DefaultArchiveFilter)
	s.Equal(30*24*time.Hour, yca_archive.ArchivedDataRetentionPeriod)
}

func (s *ArchiveTestSuite) TestNormalizeArchiveFilter_EmptyDefaultsToActive() {
	filter, err := yca_archive.NormalizeArchiveFilter("")
	s.NoError(err)
	s.Equal(yca_archive.ArchiveFilterActive, filter)
}

func (s *ArchiveTestSuite) TestNormalizeArchiveFilter_Active() {
	filter, err := yca_archive.NormalizeArchiveFilter(yca_archive.ArchiveFilterActive)
	s.NoError(err)
	s.Equal(yca_archive.ArchiveFilterActive, filter)
}

func (s *ArchiveTestSuite) TestNormalizeArchiveFilter_Archived() {
	filter, err := yca_archive.NormalizeArchiveFilter(yca_archive.ArchiveFilterArchived)
	s.NoError(err)
	s.Equal(yca_archive.ArchiveFilterArchived, filter)
}

func (s *ArchiveTestSuite) TestNormalizeArchiveFilter_Invalid() {
	filter, err := yca_archive.NormalizeArchiveFilter(yca_archive.ArchiveFilter("nope"))
	s.Error(err)
	s.Empty(filter)
	s.Contains(err.Error(), `invalid archive filter "nope"`)
}

func (s *ArchiveTestSuite) TestParseArchiveFilterQuery_EmptyDefaultsToActive() {
	filter, err := yca_archive.ParseArchiveFilterQuery("")
	s.NoError(err)
	s.Equal(yca_archive.ArchiveFilterActive, filter)
}

func (s *ArchiveTestSuite) TestParseArchiveFilterQuery_Active() {
	filter, err := yca_archive.ParseArchiveFilterQuery("active")
	s.NoError(err)
	s.Equal(yca_archive.ArchiveFilterActive, filter)
}

func (s *ArchiveTestSuite) TestParseArchiveFilterQuery_Archived() {
	filter, err := yca_archive.ParseArchiveFilterQuery("archived")
	s.NoError(err)
	s.Equal(yca_archive.ArchiveFilterArchived, filter)
}

func (s *ArchiveTestSuite) TestParseArchiveFilterQuery_InvalidReturnsBadRequest() {
	filter, err := yca_archive.ParseArchiveFilterQuery("nope")
	s.Empty(filter)
	s.Error(err)

	apiErr, ok := yca_error.AsError(err)
	s.True(ok)
	s.Equal(http.StatusBadRequest, apiErr.StatusCode)
	s.Equal("InvalidArchiveFilter", apiErr.ErrorCode)
}
