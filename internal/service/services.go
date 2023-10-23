package service

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/taraslis453/shopify-customer-auth/config"
	"github.com/taraslis453/shopify-customer-auth/pkg/errs"
	"github.com/taraslis453/shopify-customer-auth/pkg/logging"

	"github.com/taraslis453/shopify-customer-auth/internal/entity"
)

type Services struct {
	Customer CustomerService
	Vendor   VendorService
}

// serviceContext provides a shared context for all services
type serviceContext struct {
	storages Storages
	cfg      *config.Config
	logger   logging.Logger
	apis     APIs
}

// Options is used to parameterize service
type Options struct {
	APIs     APIs
	Storages Storages
	Config   *config.Config
	Logger   logging.Logger
}

const (
	customerInvalidEmailOrPasswordErrCode = "invalid_email_or_password"
	customerNotFoundErrCode               = "customer_not_found"

	invalidTokenErrCode = "invalid_token"
	tokenExpiredErrCode = "token_expired"

	vendorIdNotFoundErrCode = "vendor_id_not_found"

	storeNotFoundErrCode = "store_not_found"
)

type CustomerService interface {
	// LoginCustomer is used to login customer by given email and password and return a new access token.
	LoginCustomer(ctx context.Context, opt LoginCustomerOptions) (string, error)
	// VerifyCustomerAccessToken is used to verify the customer by given token and return verified customer entity.
	VerifyCustomerAccessToken(ctx context.Context, token string) (*entity.Customer, error)
	// RefreshCustomerAccessToken is used to verify refresh token and then generate a new access token.
	RefreshCustomerAccessToken(ctx context.Context, tokenStr string) (string, error)
	// GenerateCustomerTokens is used to generate a pair of access and refresh tokens.
	GenerateCustomerTokens(ctx context.Context, userID string) (GenerateCustomerTokensOutput, error)
	// GetCustomer is used to get customer by given id
	GetCustomer(ctx context.Context, opts GetCustomerOptions) (*entity.Customer, error)
}

var (
	ErrLoginCustomerStoreNotFound          = errs.New("store not found", storeNotFoundErrCode)
	ErrLoginCustomerInvalidEmailOrPassword = errs.New("invalid email or password", customerInvalidEmailOrPasswordErrCode)

	ErrVerifyCustomerTokenInvalidToken     = errs.New("invalid authenticate token.", invalidTokenErrCode)
	ErrVerifyCustomerTokenTokenExpired     = errs.New("authenticate token expired", tokenExpiredErrCode)
	ErrVerifyCustomerTokenCustomerNotFound = errs.New("customer not found", customerNotFoundErrCode)

	ErrRefreshCustomerTokenInvalidToken     = errs.New("invalid refresh token", invalidTokenErrCode)
	ErrRefreshCustomerTokenTokenExpired     = errs.New("refresh token expired", tokenExpiredErrCode)
	ErrRefreshCustomerTokenCustomerNotFound = errs.New("customer not found", customerNotFoundErrCode)

	ErrGetCustomerCustomerNotFoundInStorage = errs.New("customer not found in storage", customerNotFoundErrCode)
	ErrGetCustomerCustomerNotFoundInVendor  = errs.New("customer not found in vendor", customerNotFoundErrCode)
)

type LoginCustomerOptions struct {
	Email         string
	Password      string
	StoreVendorID string
}

type VerifyTokenOptions struct {
	Token      string
	HMACSecret string
}

type GetCustomerOptions struct {
	ID           string
	EmailAddress string
}

type GenerateCustomerTokensOutput struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// VendorService provides business logic related to vendors.
type VendorService interface {
	// HandleInstall handles an oauth2 installation call for a vendor integration.
	HandleInstall(c *gin.Context) (redirectURL string, err error)
	// HandleRedirect handles an oauth2 redirect call for a vendor integration.
	HandleRedirect(c *gin.Context) (string, error)
}

var (
	// ErrHandleInstallVendorIDNotFound is returned when vendor ID is not found.
	ErrHandleInstallVendorIDNotFound = errs.New("vendor id not found", vendorIdNotFoundErrCode)
	// ErrHandleInstallStoreNotFound happens when store is not found while handling install
	ErrHandleInstallStoreNotFound = errs.New("store not found", storeNotFoundErrCode)
	// ErrHandleRedirectVendorIDNotFound is returned when vendor ID is not found.
	ErrHandleRedirectVendorIDNotFound = errs.New("vendor id not found", vendorIdNotFoundErrCode)
	// ErrHandleRedirectStoreNotFound happens when store is not found while handling redirect
	ErrHandleRedirectStoreNotFound = errs.New("store not found", storeNotFoundErrCode)
)
