package entity

import (
	"time"

	"gorm.io/gorm"
)

// Store model represents a store.
type Store struct {
	ID string `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`

	VendorID              string `json:"vendorId" binding:"required"`
	Scope                 string `json:"scope"`
	ClientSecret          string `json:"clientSecret" binding:"required"`
	ClientID              string `json:"clientId" binding:"required"`
	AccessToken           string `json:"accessToken"`
	StoreFrontAccessToken string `json:"storeFrontAccessToken"`

	CreatedAt time.Time      `json:"createdAt,omitempty" gorm:"index"`
	UpdatedAt time.Time      `json:"updatedAt,omitempty"`
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"index"`
}
