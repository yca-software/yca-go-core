package yca_google_maps_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	yca_error "github.com/yca-software/yca-go-core/error"
	yca_google_maps "github.com/yca-software/yca-go-core/google/maps"
)

type MapsSuite struct {
	suite.Suite
}

func TestMapsSuite(t *testing.T) {
	suite.Run(t, new(MapsSuite))
}

func (s *MapsSuite) TestMapsConfig_Enabled() {
	s.False((yca_google_maps.MapsConfig{}).Enabled())
	s.True((yca_google_maps.MapsConfig{APIKey: "key"}).Enabled())
}

func (s *MapsSuite) newMaps(handler http.HandlerFunc) yca_google_maps.Maps {
	return yca_google_maps.NewMapsService(
		yca_google_maps.MapsConfig{APIKey: "test-key"},
		&http.Client{Transport: handlerTransport{handler: handler}},
		nil,
	)
}

func (s *MapsSuite) TestAutocompleteLocation_emptyInput() {
	maps := s.newMaps(func(w http.ResponseWriter, r *http.Request) {
		s.Fail("should not call API for empty input")
	})

	resp, err := maps.AutocompleteLocation(context.Background(), "")
	s.Require().NoError(err)
	s.Empty(resp.Predictions)
}

func (s *MapsSuite) TestAutocompleteLocation_success() {
	maps := s.newMaps(func(w http.ResponseWriter, r *http.Request) {
		s.Contains(r.URL.Path, "autocomplete")
		s.Equal("paris", r.URL.Query().Get("input"))
		_ = json.NewEncoder(w).Encode(yca_google_maps.GooglePlacesAutocompleteResponse{
			Status: "OK",
			Predictions: []yca_google_maps.GooglePlacesAutocompletePrediction{
				{
					PlaceID:     "place-1",
					Description: "Paris, France",
					StructuredFormatting: yca_google_maps.GooglePlacesStructuredFormatting{
						MainText:      "Paris",
						SecondaryText: "France",
					},
				},
			},
		})
	})

	resp, err := maps.AutocompleteLocation(context.Background(), "paris")
	s.Require().NoError(err)
	s.Require().Len(resp.Predictions, 1)
	s.Equal("place-1", resp.Predictions[0].PlaceID)
	s.Equal("Paris", resp.Predictions[0].StructuredFormatting.MainText)
}

func (s *MapsSuite) TestAutocompleteLocation_zeroResults() {
	maps := s.newMaps(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(yca_google_maps.GooglePlacesAutocompleteResponse{Status: "ZERO_RESULTS"})
	})

	resp, err := maps.AutocompleteLocation(context.Background(), "zzzznone")
	s.Require().NoError(err)
	s.Empty(resp.Predictions)
}

func (s *MapsSuite) TestGetPlaceDetails_emptyPlaceID() {
	maps := s.newMaps(func(w http.ResponseWriter, r *http.Request) {
		s.Fail("should not call API for empty place ID")
	})

	_, err := maps.GetPlaceDetails(context.Background(), "")
	s.Require().Error(err)

	var apiErr *yca_error.Error
	s.Require().ErrorAs(err, &apiErr)
	s.Equal(400, apiErr.StatusCode)
}

func (s *MapsSuite) TestGetPlaceDetails_andGetLocationData() {
	maps := s.newMaps(func(w http.ResponseWriter, r *http.Request) {
		s.Contains(r.URL.Path, "details")
		_ = json.NewEncoder(w).Encode(yca_google_maps.GooglePlacesDetailsResponse{
			Status: "OK",
			Result: yca_google_maps.GooglePlacesDetailsResult{
				PlaceID:          "place-paris",
				FormattedAddress: "Paris, France",
				AddressComponents: []yca_google_maps.GooglePlacesAddressComponent{
					{LongName: "Paris", Types: []string{"locality"}},
					{LongName: "75001", Types: []string{"postal_code"}},
					{LongName: "France", Types: []string{"country"}},
				},
				Geometry: yca_google_maps.GooglePlacesGeometry{
					Location: yca_google_maps.GooglePlacesLatLng{Lat: 48.8566, Lng: 2.3522},
				},
			},
		})
	})

	details, err := maps.GetPlaceDetails(context.Background(), "place-paris")
	s.Require().NoError(err)
	s.Equal("Paris, France", details.Result.FormattedAddress)

	location, err := maps.GetLocationData(context.Background(), "place-paris")
	s.Require().NoError(err)
	s.Equal("Paris", location.City)
	s.Equal("75001", location.Zip)
	s.Equal("France", location.Country)
	s.Equal("Europe/Paris", location.Timezone)
	s.InDelta(48.8566, location.Geo.Lat, 0.001)
}

type handlerTransport struct {
	handler http.HandlerFunc
}

func (t handlerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	rec := &responseRecorder{header: make(http.Header)}
	t.handler(rec, req)
	return rec.Result(), nil
}

type responseRecorder struct {
	code   int
	header http.Header
	body   strings.Builder
}

func (r *responseRecorder) Header() http.Header { return r.header }

func (r *responseRecorder) Write(b []byte) (int, error) {
	if r.code == 0 {
		r.code = http.StatusOK
	}
	return r.body.Write(b)
}

func (r *responseRecorder) WriteHeader(statusCode int) { r.code = statusCode }

func (r *responseRecorder) Result() *http.Response {
	return &http.Response{
		StatusCode: r.code,
		Header:     r.header,
		Body:       io.NopCloser(strings.NewReader(r.body.String())),
	}
}
