package yca_google_maps

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/bradfitz/latlong"
	yca_error "github.com/yca-software/yca-go-core/error"
	yca_logger "github.com/yca-software/yca-go-core/logger"
)

type MapsConfig struct {
	APIKey string
}

func (c MapsConfig) Enabled() bool {
	return c.APIKey != ""
}

// Maps provides Google Places / geocoding helpers.
type Maps interface {
	AutocompleteLocation(ctx context.Context, input string) (*AutocompleteLocationResponse, error)
	GetPlaceDetails(ctx context.Context, placeID string) (*PlaceDetailsResponse, error)
	GetLocationData(ctx context.Context, placeID string) (*LocationData, error)
}

type mapsService struct {
	apiKey     string
	httpClient *http.Client
	logger     yca_logger.Logger
}

func NewMapsService(cfg MapsConfig, httpClient *http.Client, logger yca_logger.Logger) Maps {
	return &mapsService{
		apiKey:     cfg.APIKey,
		httpClient: httpClient,
		logger:     logger,
	}
}

// placeAutocompleteQuery builds Places autocomplete params. No `types` filter so
// establishments (restaurants, cafes, etc.) appear alongside addresses.
func placeAutocompleteQuery(input, apiKey string) url.Values {
	params := url.Values{}
	params.Add("input", input)
	params.Add("key", apiKey)
	return params
}

func (s *mapsService) AutocompleteLocation(ctx context.Context, input string) (*AutocompleteLocationResponse, error) {
	if input == "" {
		return &AutocompleteLocationResponse{Predictions: []PlacePrediction{}}, nil
	}

	apiURL := "https://maps.googleapis.com/maps/api/place/autocomplete/json"
	params := placeAutocompleteQuery(input, s.apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, yca_error.NewInternalServerError(err, "", nil)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logError(ctx, "AutocompleteLocation", "request failed", err)
		return nil, yca_error.NewInternalServerError(err, "", nil)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, yca_error.NewInternalServerError(err, "", nil)
	}

	var googleResponse GooglePlacesAutocompleteResponse

	if resp.StatusCode != http.StatusOK {
		s.logError(ctx, "AutocompleteLocation", fmt.Sprintf("status %d: %s", resp.StatusCode, string(body)), nil)
		if err := json.Unmarshal(body, &googleResponse); err == nil && googleResponse.ErrorMessage != "" {
			return nil, s.errorForPlacesStatus(googleResponse.Status)
		}
		return nil, yca_error.NewServiceUnavailableError(errors.New("location search unavailable"), "LocationSearchUnavailable", nil)
	}

	if err := json.Unmarshal(body, &googleResponse); err != nil {
		return nil, yca_error.NewInternalServerError(err, "LocationSearchProcessingErr", nil)
	}

	switch googleResponse.Status {
	case "OK":
	case "ZERO_RESULTS":
		return &AutocompleteLocationResponse{Predictions: []PlacePrediction{}}, nil
	default:
		return nil, s.errorForPlacesStatus(googleResponse.Status)
	}

	predictions := make([]PlacePrediction, len(googleResponse.Predictions))
	for i, p := range googleResponse.Predictions {
		predictions[i] = PlacePrediction{
			PlaceID:     p.PlaceID,
			Description: p.Description,
			StructuredFormatting: StructuredFormatting{
				MainText:      p.StructuredFormatting.MainText,
				SecondaryText: p.StructuredFormatting.SecondaryText,
			},
		}
	}

	return &AutocompleteLocationResponse{Predictions: predictions}, nil
}

func (s *mapsService) GetPlaceDetails(ctx context.Context, placeID string) (*PlaceDetailsResponse, error) {
	if placeID == "" {
		return nil, yca_error.NewBadRequestError(errors.New("place ID is required"), "", nil)
	}

	apiURL := "https://maps.googleapis.com/maps/api/place/details/json"
	params := url.Values{}
	params.Add("place_id", placeID)
	params.Add("key", s.apiKey)
	params.Add("fields", "place_id,formatted_address,address_components,geometry")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, yca_error.NewInternalServerError(err, "", nil)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logError(ctx, "GetPlaceDetails", "request failed", err)
		return nil, yca_error.NewInternalServerError(err, "", nil)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, yca_error.NewInternalServerError(err, "", nil)
	}

	if resp.StatusCode != http.StatusOK {
		s.logError(ctx, "GetPlaceDetails", fmt.Sprintf("status %d: %s", resp.StatusCode, string(body)), nil)
		return nil, yca_error.NewInternalServerError(errors.New("place details unavailable"), "", nil)
	}

	var googleResponse GooglePlacesDetailsResponse

	if err := json.Unmarshal(body, &googleResponse); err != nil {
		return nil, yca_error.NewInternalServerError(err, "", nil)
	}

	if googleResponse.Status != "OK" {
		s.logError(ctx, "GetPlaceDetails", fmt.Sprintf("places status %s: %s", googleResponse.Status, googleResponse.ErrorMessage), nil)
		return nil, yca_error.NewInternalServerError(errors.New(googleResponse.ErrorMessage), "", nil)
	}

	addressComponents := make([]AddressComponent, len(googleResponse.Result.AddressComponents))
	for i, ac := range googleResponse.Result.AddressComponents {
		addressComponents[i] = AddressComponent{
			LongName:  ac.LongName,
			ShortName: ac.ShortName,
			Types:     ac.Types,
		}
	}

	return &PlaceDetailsResponse{
		Result: PlaceDetails{
			PlaceID:           googleResponse.Result.PlaceID,
			FormattedAddress:  googleResponse.Result.FormattedAddress,
			AddressComponents: addressComponents,
			Geometry: PlaceGeometry{
				Location: Point{
					Lat: googleResponse.Result.Geometry.Location.Lat,
					Lng: googleResponse.Result.Geometry.Location.Lng,
				},
			},
		},
	}, nil
}

func (s *mapsService) GetLocationData(ctx context.Context, placeID string) (*LocationData, error) {
	placeDetails, err := s.GetPlaceDetails(ctx, placeID)
	if err != nil {
		return nil, err
	}

	place := placeDetails.Result

	var city, zip, country string
	for _, component := range place.AddressComponents {
		for _, t := range component.Types {
			if city == "" && t == "locality" {
				city = component.LongName
			} else if city == "" && t == "administrative_area_level_1" {
				city = component.LongName
			}
			if zip == "" && t == "postal_code" {
				zip = component.LongName
			}
			if country == "" && t == "country" {
				country = component.LongName
			}
		}
	}

	lat := place.Geometry.Location.Lat
	lng := place.Geometry.Location.Lng

	return &LocationData{
		Address:  place.FormattedAddress,
		City:     city,
		Zip:      zip,
		Country:  country,
		PlaceID:  place.PlaceID,
		Geo:      Point{Lat: lat, Lng: lng},
		Timezone: detectTimezone(lat, lng),
	}, nil
}

func detectTimezone(lat, lng float64) string {
	timezoneName := latlong.LookupZoneName(lat, lng)
	if timezoneName == "" {
		return "UTC"
	}
	if _, err := time.LoadLocation(timezoneName); err != nil {
		return "UTC"
	}
	return timezoneName
}

func (s *mapsService) errorForPlacesStatus(status string) *yca_error.Error {
	switch status {
	case "INVALID_REQUEST":
		return yca_error.NewBadRequestError(errors.New("invalid request"), "LocationSearchInvalidQuery", nil)
	case "OVER_QUERY_LIMIT":
		return yca_error.NewBadRequestError(errors.New("over query limit"), "LocationSearchTemporaryError", nil)
	case "REQUEST_DENIED":
		return yca_error.NewBadRequestError(errors.New("request denied"), "LocationSearchDenied", nil)
	case "UNKNOWN_ERROR":
		return yca_error.NewInternalServerError(errors.New("unknown error"), "", nil)
	default:
		return yca_error.NewServiceUnavailableError(errors.New("location search unavailable"), "LocationSearchUnavailable", nil)
	}
}

func (s *mapsService) logError(ctx context.Context, method, msg string, err error) {
	if s.logger == nil {
		return
	}
	args := []any{"method", method, "message", msg}
	if err != nil {
		args = append(args, "error", err)
	}
	s.logger.WithContext(ctx).Error("google maps API error", args...)
}
