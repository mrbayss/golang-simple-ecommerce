package model

import (
	"time"

	"github.com/mrbayss/golang-simple-ecommerce/internal/entity"
)

type CreateOrderItemReq struct {
	ProductID string `json:"product_id" validate:"required,uuid"`
	Quantity  int    `json:"quantity" validate:"required,min=1,max=99"`
}

type CreateOrderReq struct {
	CustomerName    string               `json:"customer_name" validate:"required,min=2,max=100"`
	CustomerPhone   string               `json:"customer_phone" validate:"required,min=9,max=20"`
	CustomerAddress *string              `json:"customer_address" validate:"omitempty,max=500"`
	Notes           *string              `json:"notes" validate:"omitempty,max=255"`
	PaymentMethod   string               `json:"payment_method" validate:"required,oneof=COD TRANSFER QRIS"`
	Items           []CreateOrderItemReq `json:"items" validate:"required,min=1,max=50,dive"`
}

type UpdateOrderStatusReq struct {
	Status entity.OrderStatus `json:"status" validate:"required,oneof=PENDING CONFIRMED READY COMPLETED CANCELLED"`
}

type OrderItemRes struct {
	ID           string       `json:"id"`
	ProductID    string       `json:"product_id"`
	ProductName  string       `json:"product_name"`
	ProductPrice entity.Money `json:"product_price"`
	Quantity     int          `json:"quantity"`
	Subtotal     entity.Money `json:"subtotal"`
}

type OrderRes struct {
	ID              string         `json:"id"`
	OrderCode       string         `json:"order_code"`
	CustomerName    string         `json:"customer_name"`
	CustomerPhone   string         `json:"customer_phone"`
	CustomerAddress *string        `json:"customer_address,omitempty"`
	Notes           *string        `json:"notes,omitempty"`
	PaymentMethod   string         `json:"payment_method"`
	Status          string         `json:"status"`
	TotalPrice      entity.Money   `json:"total_price"`
	Items           []OrderItemRes `json:"items"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       *time.Time     `json:"updated_at,omitempty"`
}

type OrderListRes = PaginatedRes[OrderRes]
