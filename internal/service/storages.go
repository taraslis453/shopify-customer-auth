package service

import (
	"context"

	"github.com/taraslis453/shopify-customer-auth/internal/entity"
)

type Storages struct {
	Customer CustomerStorage
}

type CustomerStorage interface {
	GetCustomer(ctx context.Context, filter GetCustomerFilter) (*entity.Customer, error)
	CreateCustomer(ctx context.Context, user *entity.Customer) (*entity.Customer, error)
	UpdateCustomer(ctx context.Context, id string, user *entity.Customer) (*entity.Customer, error)
}

type GetCustomerFilter struct {
	ID               *string
	VendorCustomerID *string
}
