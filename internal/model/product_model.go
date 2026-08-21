package model

import "time"

type ProductRes struct {
	ID          string             `json:"id"`
	CategoryID  *string            `json:"category_id,omitempty"`
	Category    *CategoryRes       `json:"category,omitempty"`
	Name        string             `json:"name"`
	Slug        string             `json:"slug"`
	Description string             `json:"description"`
	Price       int64              `json:"price"` // Rupiah utuh
	Stock       int                `json:"stock"`
	Weight      int                `json:"weight"`
	Images      []ProductImageRes  `json:"images"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   *time.Time         `json:"updated_at,omitempty"`
}

type ProductImageRes struct {
	ID        string `json:"id"`
	ImageURL  string `json:"image_url"`
	IsPrimary bool   `json:"is_primary"`
}

type ProductListRes = PaginatedRes[ProductRes]
