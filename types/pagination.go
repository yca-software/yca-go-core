package chi_types

type PaginatedListResponse[T any] struct {
	Items   []T  `json:"items"`
	HasNext bool `json:"hasNext"`
}
