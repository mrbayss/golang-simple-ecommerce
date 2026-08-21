package converter

import (
	"github.com/mrbayss/golang-simple-ecommerce/internal/entity"
	"github.com/mrbayss/golang-simple-ecommerce/internal/model"
)

func ToProductRes(product *entity.Product) *model.ProductRes {
	res := &model.ProductRes{
		ID:          product.ID.String(),
		Name:        product.Name,
		Slug:        product.Slug,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		Weight:      product.Weight,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
		Images:      make([]model.ProductImageRes, 0, len(product.Images)),
	}

	if product.CategoryID != nil && product.Category != nil {
		categoryID := product.CategoryID.String()
		res.CategoryID = &categoryID
		res.Category = ToCategoryRes(product.Category)
	}

	for _, img := range product.Images {
		res.Images = append(res.Images, model.ProductImageRes{
			ID:        img.ID.String(),
			ImageURL:  img.ImageURL,
			IsPrimary: img.IsPrimary,
		})
	}

	return res
}
