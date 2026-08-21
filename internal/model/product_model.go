package model

import (
	"time"

	"github.com/mrbayss/golang-simple-ecommerce/internal/entity"
)

type ProductRes struct {
	ID          string            `json:"id"`
	CategoryID  *string           `json:"category_id,omitempty"`
	Category    *CategoryRes      `json:"category,omitempty"`
	Name        string            `json:"name"`
	Slug        string            `json:"slug"`
	Description string            `json:"description"`
	Price       entity.Money      `json:"price"`
	Stock       int               `json:"stock"`
	Weight      int               `json:"weight"`
	Images      []ProductImageRes `json:"images"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   *time.Time        `json:"updated_at,omitempty"`
}

type ProductImageRes struct {
	ID        string `json:"id"`
	ImageURL  string `json:"image_url"`
	IsPrimary bool   `json:"is_primary"`
}

type ProductListRes = PaginatedRes[ProductRes]
