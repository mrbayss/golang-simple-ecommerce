package model

import (
	"time"

	"github.com/mrbayss/golang-simple-ecommerce/internal/entity"
)

// CreateProductReq is the form part of a multipart product creation request.
type CreateProductReq struct {
	Name        string  `form:"name" validate:"required,min=3"`
	Slug        string  `form:"slug" validate:"omitempty"`
	Description string  `form:"description" validate:"omitempty"`
	Price       float64 `form:"price" validate:"required,gt=0"`
	Stock       int     `form:"stock" validate:"required,gte=0"`
	Weight      int     `form:"weight" validate:"omitempty,gte=0"`
	CategoryID  string  `form:"category_id" validate:"omitempty,uuid"`
}

// UpdateProductReq is the form part of a multipart product update request.
// All fields are optional; only sent fields are applied.
type UpdateProductReq struct {
	Name        *string  `form:"name" validate:"omitempty,min=3"`
	Slug        *string  `form:"slug" validate:"omitempty"`
	Description *string  `form:"description" validate:"omitempty"`
	Price       *float64 `form:"price" validate:"omitempty,gt=0"`
	Stock       *int     `form:"stock" validate:"omitempty,gte=0"`
	Weight      *int     `form:"weight" validate:"omitempty,gte=0"`
	CategoryID  *string  `form:"category_id" validate:"omitempty,uuid"`
}

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
