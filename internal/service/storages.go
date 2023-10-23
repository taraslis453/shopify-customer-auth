package service

import (
	"github.com/taraslis453/shopify-customer-auth/internal/entity"
)

type Storages struct {
	Customer CustomerStorage
	Store    StoreStorage
}

type CustomerStorage interface {
	GetCustomer(filter GetCustomerFilter) (*entity.Customer, error)
	CreateCustomer(user *entity.Customer) (*entity.Customer, error)
	UpdateCustomer(id string, user *entity.Customer) (*entity.Customer, error)
}

type GetCustomerFilter struct {
	ID               *string
	VendorCustomerID *string
}

type StoreStorage interface {
	GetStore(vendorID *string) (*entity.Store, error)
	UpdateStore(id string, store *entity.Store) (*entity.Store, error)
}
