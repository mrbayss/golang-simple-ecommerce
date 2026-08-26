package model

type PageMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type PaginatedRes[T any] struct {
	Data []T      `json:"data"`
	Meta PageMeta `json:"meta"`
}
