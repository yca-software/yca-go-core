package yca_google_maps

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockMaps is a testify mock for Maps.
type MockMaps struct {
	mock.Mock
}

func (m *MockMaps) AutocompleteLocation(ctx context.Context, input string) (*AutocompleteLocationResponse, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AutocompleteLocationResponse), args.Error(1)
}

func (m *MockMaps) GetPlaceDetails(ctx context.Context, placeID string) (*PlaceDetailsResponse, error) {
	args := m.Called(ctx, placeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*PlaceDetailsResponse), args.Error(1)
}

func (m *MockMaps) GetLocationData(ctx context.Context, placeID string) (*LocationData, error) {
	args := m.Called(ctx, placeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*LocationData), args.Error(1)
}
