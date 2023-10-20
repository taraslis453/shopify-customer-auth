package entity

// Customer model represents a customer.
type Customer struct {
	ID               string `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	VendorCustomerID string `json:"vendorCustomerId" gorm:"index"`
	// NOTE: we are getting it from Vendor API and not storing it in our DB
	FirstName string `json:"firstName" gorm:"-"`

	RefreshToken string `json:"refreshToken"`
}
