package storage

import (
	"github.com/taraslis453/shopify-customer-auth/pkg/postgresql"

	"github.com/taraslis453/shopify-customer-auth/internal/entity"
	"github.com/taraslis453/shopify-customer-auth/internal/service"
)

var _ service.StoreStorage = (*storeStorage)(nil)

type storeStorage struct {
	*postgresql.PostgreSQLGorm
}

func NewStoreStorage(postgresql *postgresql.PostgreSQLGorm) *storeStorage {
	return &storeStorage{postgresql}
}

func (s *storeStorage) GetStore(vendorID *string) (*entity.Store, error) {
	store := &entity.Store{}
	err := s.DB.Where("vendor_id = ?", vendorID).First(store).Error
	if err != nil {
		return nil, err
	}

	return store, nil
}

func (s *storeStorage) UpdateStore(id string, store *entity.Store) (*entity.Store, error) {
	err := s.DB.Model(&entity.Store{}).Where("id = ?", id).Updates(store).Error
	if err != nil {
		return nil, err
	}

	return store, nil
}
