package converter

import (
	"github.com/mrbayss/golang-simple-ecommerce/internal/entity"
	"github.com/mrbayss/golang-simple-ecommerce/internal/model"
)

func ToOrderRes(order *entity.Order) *model.OrderRes {
	res := &model.OrderRes{
		ID:              order.ID.String(),
		OrderCode:       order.OrderCode,
		CustomerName:    order.CustomerName,
		CustomerPhone:   order.CustomerPhone,
		CustomerAddress: order.CustomerAddress,
		Notes:           order.Notes,
		PaymentMethod:   string(order.PaymentMethod),
		Status:          string(order.Status),
		TotalPrice:      order.TotalPrice,
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
		Items:           make([]model.OrderItemRes, 0, len(order.Items)),
	}

	for _, item := range order.Items {
		res.Items = append(res.Items, model.OrderItemRes{
			ID:           item.ID.String(),
			ProductID:    item.ProductID.String(),
			ProductName:  item.ProductName,
			ProductPrice: item.ProductPrice,
			Quantity:     item.Quantity,
			Subtotal:     item.Subtotal,
		})
	}

	return res
}
