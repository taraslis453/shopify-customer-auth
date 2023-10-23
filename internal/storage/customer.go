package storage

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/taraslis453/shopify-customer-auth/pkg/postgresql"

	"github.com/taraslis453/shopify-customer-auth/internal/entity"
	"github.com/taraslis453/shopify-customer-auth/internal/service"
)

var _ service.CustomerStorage = (*customerStorage)(nil)

type customerStorage struct {
	*postgresql.PostgreSQLGorm
}

func NewCustomerStorage(postgresql *postgresql.PostgreSQLGorm) *customerStorage {
	return &customerStorage{postgresql}
}

func (r *customerStorage) GetCustomer(filter service.GetCustomerFilter) (*entity.Customer, error) {
	stmt := r.DB
	if filter.ID != nil {
		stmt = stmt.Where(entity.Customer{ID: *filter.ID})
	}
	if filter.VendorCustomerID != nil {
		stmt = stmt.Where(entity.Customer{VendorCustomerID: *filter.VendorCustomerID})
	}

	var customer entity.Customer
	err := stmt.First(&customer).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return &customer, nil
}

func (r *customerStorage) CreateCustomer(customer *entity.Customer) (*entity.Customer, error) {
	err := r.DB.Create(customer).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	return customer, nil
}

func (r *customerStorage) UpdateCustomer(id string, customer *entity.Customer) (*entity.Customer, error) {
	err := r.DB.Model(&entity.Customer{}).Where("id = ?", id).Updates(customer).Error
	if err != nil {
		return nil, fmt.Errorf("failed to update customer: %w", err)
	}

	return customer, nil
}
