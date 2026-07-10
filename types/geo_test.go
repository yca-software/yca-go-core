package yca_types_test

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/suite"
	yca_types "github.com/yca-software/yca-go-core/types"
)

type GeoSuite struct {
	suite.Suite
}

func TestGeoSuite(t *testing.T) {
	suite.Run(t, new(GeoSuite))
}

func (s *GeoSuite) TestPointValue() {
	p := yca_types.Point{Lng: 2.3522, Lat: 48.8566}
	val, err := p.Value()
	s.Require().NoError(err)
	s.Equal("SRID=4326;POINT(2.3522 48.8566)", val)
}

func (s *GeoSuite) TestPointScanEWKB() {
	want := yca_types.Point{Lng: -73.9857, Lat: 40.7484}
	hexPayload := hex.EncodeToString(buildEWKB(want.Lng, want.Lat))

	var got yca_types.Point
	s.Require().NoError(got.Scan(hexPayload))
	s.Equal(want, got)
}

func buildEWKB(lng, lat float64) []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, byte(1))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x20000001))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(4326))
	_ = binary.Write(&buf, binary.LittleEndian, lng)
	_ = binary.Write(&buf, binary.LittleEndian, lat)
	return buf.Bytes()
}

func (s *GeoSuite) TestPolygonScanEWKBBytes() {
	ring := []yca_types.Point{
		{Lng: 10.7, Lat: 59.9},
		{Lng: 10.71, Lat: 59.9},
		{Lng: 10.71, Lat: 59.91},
		{Lng: 10.7, Lat: 59.91},
	}
	payload := buildPolygonEWKB(ring)

	var got yca_types.Polygon
	s.Require().NoError(got.Scan(payload))
	s.Len(got, 4)
	s.InDelta(10.7, got[0].Lng, 0.001)
	s.InDelta(59.9, got[0].Lat, 0.001)
}

func buildPolygonEWKB(ring []yca_types.Point) []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, byte(1))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x20000003))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(4326))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(1))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(len(ring)))
	for _, p := range ring {
		_ = binary.Write(&buf, binary.LittleEndian, p.Lng)
		_ = binary.Write(&buf, binary.LittleEndian, p.Lat)
	}
	return buf.Bytes()
}
