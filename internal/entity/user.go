package entity

import (
	"time"

	"gorm.io/gorm"
)

// User represents the user model stored in the database.
type User struct {
	ID string `json:"id,omitempty" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" binding:"required"`

	Name         string `json:"name,omitempty"`
	Surname      string `json:"surname,omitempty"`
	Phone        string `json:"phone,omitempty"`
	EmailAddress string `json:"emailAddress,omitempty"`
	Password     string `json:"-"`

	CreatedAt time.Time      `json:"createdAt,omitempty" gorm:"index"`
	UpdatedAt time.Time      `json:"updatedAt,omitempty"`
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"index" swaggerignore:"true"`
}
