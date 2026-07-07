package chi_jsonb_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	chi_jsonb "github.com/yca-software/yca-go-core/jsonb"
)

type JSONBSuite struct {
	suite.Suite
}

func TestJSONBSuite(t *testing.T) {
	suite.Run(t, new(JSONBSuite))
}

func (s *JSONBSuite) TestJSONBValueAndScan() {
	src := []string{"workspace:read", "workspace:write"}
	val, err := chi_jsonb.JSONBValue(src)
	s.Require().NoError(err)

	var dest []string
	s.Require().NoError(chi_jsonb.JSONBScan(val, &dest))
	s.Equal(src, dest)
}

func (s *JSONBSuite) TestJSONBScanNil() {
	var dest []string
	s.NoError(chi_jsonb.JSONBScan(nil, &dest))
	s.Nil(dest)
}
