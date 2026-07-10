package yca_google_maps

type Point struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type LocationData struct {
	Address  string `json:"address"`
	City     string `json:"city"`
	Zip      string `json:"zip"`
	Country  string `json:"country"`
	PlaceID  string `json:"placeId"`
	Geo      Point  `json:"geo"`
	Timezone string `json:"timezone"`
}

type AddressComponent struct {
	LongName  string   `json:"longName"`
	ShortName string   `json:"shortName"`
	Types     []string `json:"types"`
}

type PlaceGeometry struct {
	Location Point `json:"location"`
}

type PlaceDetails struct {
	PlaceID           string             `json:"placeId"`
	FormattedAddress  string             `json:"formattedAddress"`
	AddressComponents []AddressComponent `json:"addressComponents"`
	Geometry          PlaceGeometry      `json:"geometry"`
}

type StructuredFormatting struct {
	MainText      string `json:"mainText"`
	SecondaryText string `json:"secondaryText"`
}

type PlacePrediction struct {
	PlaceID              string               `json:"placeId"`
	Description          string               `json:"description"`
	StructuredFormatting StructuredFormatting `json:"structuredFormatting"`
}

type AutocompleteLocationResponse struct {
	Predictions []PlacePrediction `json:"predictions"`
}

type PlaceDetailsResponse struct {
	Result PlaceDetails `json:"result"`
}

// GooglePlacesAutocompleteResponse is the raw Places API autocomplete payload.
type GooglePlacesAutocompleteResponse struct {
	Status       string                               `json:"status"`
	Predictions  []GooglePlacesAutocompletePrediction `json:"predictions"`
	ErrorMessage string                               `json:"error_message,omitempty"`
}

type GooglePlacesAutocompletePrediction struct {
	PlaceID              string                           `json:"place_id"`
	Description          string                           `json:"description"`
	StructuredFormatting GooglePlacesStructuredFormatting `json:"structured_formatting"`
}

type GooglePlacesStructuredFormatting struct {
	MainText      string `json:"main_text"`
	SecondaryText string `json:"secondary_text"`
}

// GooglePlacesDetailsResponse is the raw Places API details payload.
type GooglePlacesDetailsResponse struct {
	Status       string                    `json:"status"`
	Result       GooglePlacesDetailsResult `json:"result"`
	ErrorMessage string                    `json:"error_message,omitempty"`
}

type GooglePlacesDetailsResult struct {
	PlaceID           string                         `json:"place_id"`
	FormattedAddress  string                         `json:"formatted_address"`
	AddressComponents []GooglePlacesAddressComponent `json:"address_components"`
	Geometry          GooglePlacesGeometry           `json:"geometry"`
}

type GooglePlacesAddressComponent struct {
	LongName  string   `json:"long_name"`
	ShortName string   `json:"short_name"`
	Types     []string `json:"types"`
}

type GooglePlacesGeometry struct {
	Location GooglePlacesLatLng `json:"location"`
}

type GooglePlacesLatLng struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}
