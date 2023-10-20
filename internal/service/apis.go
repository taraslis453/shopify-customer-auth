package service

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/taraslis453/shopify-customer-auth/internal/entity"
)

// APIs provides a collection of API interfaces.
type APIs struct {
	VendorAPI
}

type VendorAPI interface {
	// WithStore returns a new Vendor based on a store config.
	WithStore(config *entity.Store) VendorAPI
	// HandleInstall handles an oauth2 installation call.
	HandleInstall(c *gin.Context, redirectURL string) (*entity.Store, string, error)
	// HandleRedirect handles an oauth2 redirect call.
	HandleRedirect(c *gin.Context) (newConfig *entity.Store, err error)
	// GetLoggedInCustomerID returns the id of the logged in customer.
	GetLoggedInCustomerID(ctx context.Context, opt LoginCustomerOptions) (string, error)
	// GetCustomerByVendorID returns the customer by vendor id.
	GetCustomerByVendorID(ctx context.Context, vendorCustomerID string) (*entity.Customer, error)
}

// VendorListCustomersFilterOperator represents operator variants for VendorListCustomersFilter.
type VendorListCustomersFilterOperator string

const (
	VendorListCustomersFilterOperatorAND VendorListCustomersFilterOperator = "AND"
	VendorListCustomersFilterOperatorOR  VendorListCustomersFilterOperator = "OR"
)

// VendorListCustomersFilter provides filters for querying of vendor customers.
type VendorListCustomersFilter struct {
	// Email is the email of the customer to filter by.
	Email *string
	// Phone is the phone of the customer to filter by.
	Phone *string
	// Operator is the operator for request querying.
	// default - AND
	Operator VendorListCustomersFilterOperator
}
