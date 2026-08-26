package converter

import (
	"github.com/mrbayss/golang-simple-ecommerce/internal/entity"
	"github.com/mrbayss/golang-simple-ecommerce/internal/model"
)

func ToCategoryRes(category *entity.Category) *model.CategoryRes {
	return &model.CategoryRes{
		ID:        category.ID.String(),
		Name:      category.Name,
		Slug:      category.Slug,
		CreatedAt: category.CreatedAt,
		UpdatedAt: category.UpdatedAt,
	}
}
