package chi_types

import (
	"bytes"
	"database/sql/driver"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
)

// Point is a WGS84 (SRID 4326) geographic point for PostGIS geography/geometry columns.
type Point struct {
	Lng float64 `json:"lng"`
	Lat float64 `json:"lat"`
}

type ewkbPoint struct {
	ByteOrder byte
	WkbType   uint32
	SRID      uint32
	Point     Point
}

func (p *Point) String() string {
	return fmt.Sprintf("SRID=4326;POINT(%v %v)", p.Lng, p.Lat)
}

// Scan implements sql.Scanner for PostGIS EWKB hex or byte payloads.
func (p *Point) Scan(val any) error {
	switch v := val.(type) {
	case nil:
		return nil
	case string:
		b, err := hex.DecodeString(v)
		if err != nil {
			return err
		}
		return p.scanBytes(b)
	case []byte:
		return p.scanBytes(v)
	default:
		return fmt.Errorf("Point.Scan: unsupported type %T", val)
	}
}

func (p *Point) scanBytes(b []byte) error {
	r := bytes.NewReader(b)
	var ewkbP ewkbPoint
	if err := binary.Read(r, binary.LittleEndian, &ewkbP); err != nil {
		return err
	}
	if ewkbP.ByteOrder != 1 || ewkbP.WkbType != 0x20000001 || ewkbP.SRID != 4326 {
		return fmt.Errorf("Point.Scan: unexpected ewkb %#v", ewkbP)
	}
	*p = ewkbP.Point
	return nil
}

// Value implements driver.Valuer using PostGIS EWKT (SRID=4326;POINT(lng lat)).
func (p Point) Value() (driver.Value, error) {
	return p.String(), nil
}

type Polygon []Point

type ewkbPolygon struct {
	ByteOrder byte   // 1 (LittleEndian)
	WkbType   uint32 // 0x20000003 (PolygonS)
	SRID      uint32 // 4326
	Rings     uint32
	Count     uint32
}

func (p *Polygon) String() string {
	points := []string{}
	for _, point := range *p {
		points = append(points, fmt.Sprintf("%v %v", point.Lng, point.Lat))
	}
	points = append(points, fmt.Sprintf("%v %v", (*p)[0].Lng, (*p)[0].Lat))
	return fmt.Sprintf("SRID=4326;POLYGON((%s))", strings.Join(points, ","))
}

// Scan implements sql.Scanner for PostGIS EWKB hex or byte payloads.
func (p *Polygon) Scan(val any) error {
	switch v := val.(type) {
	case nil:
		return nil
	case string:
		b, err := hex.DecodeString(v)
		if err != nil {
			return err
		}
		return p.scanBytes(b)
	case []byte:
		return p.scanBytes(v)
	default:
		return fmt.Errorf("Polygon.Scan: unsupported type %T", val)
	}
}

func (p *Polygon) scanBytes(b []byte) error {
	r := bytes.NewReader(b)

	var ewkbP ewkbPolygon
	if err := binary.Read(r, binary.LittleEndian, &ewkbP); err != nil {
		return err
	}

	if ewkbP.ByteOrder != 1 || ewkbP.WkbType != 0x20000003 || ewkbP.SRID != 4326 || ewkbP.Rings != 1 {
		return fmt.Errorf("Polygon.Scan: unexpected ewkb %#v", ewkbP)
	}

	points := make([]Point, ewkbP.Count)
	if err := binary.Read(r, binary.LittleEndian, &points); err != nil {
		return err
	}
	*p = points

	return nil
}

func (p Polygon) Value() (driver.Value, error) {
	return p.String(), nil
}
