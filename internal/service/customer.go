package service

import (
	"context"
	"fmt"
	"time"

	"github.com/mitchellh/mapstructure"

	"github.com/taraslis453/shopify-customer-auth/pkg/errs"
	"github.com/taraslis453/shopify-customer-auth/pkg/token"

	"github.com/taraslis453/shopify-customer-auth/internal/entity"
)

var _ CustomerService = (*customerService)(nil)

type customerService struct {
	serviceContext
}

func NewCustomerService(options Options) *customerService {
	return &customerService{
		serviceContext: serviceContext{
			storages: options.Storages,
			cfg:      options.Config,
			logger:   options.Logger.Named("customerService"),
			apis:     options.APIs,
		},
	}
}

func (s *customerService) LoginCustomer(ctx context.Context, opts LoginCustomerOptions) (string, error) {
	logger := s.logger.
		Named("LoginCustomer").
		WithContext(ctx).
		With("opts", opts)

	store, err := s.storages.Store.GetStore(&opts.StoreVendorID)
	if err != nil {
		logger.Error("failed to get store", "err", err)
		return "", fmt.Errorf("failed to get store: %w", err)
	}
	if store == nil {
		logger.Info("store not found")
		return "", ErrLoginCustomerStoreNotFound
	}

	vendorCustomerID, err := s.apis.VendorAPI.WithStore(store).GetLoggedInCustomerID(ctx, LoginCustomerOptions{
		Email:    opts.Email,
		Password: opts.Password,
	})
	if err != nil {
		if errs.IsExpected(err) {
			logger.Info(err.Error())
			return "", err
		}

		logger.Error("failed to login customer", "err", err)
		return "", fmt.Errorf("failed to login customer: %w", err)
	}

	customer, err := s.storages.Customer.GetCustomer(GetCustomerFilter{
		VendorCustomerID: &vendorCustomerID,
	})
	if err != nil {
		logger.Error("failed to get customer through storage", "err", err)
		return "", fmt.Errorf("failed to get customer through storage: %w", err)
	}
	if customer == nil {
		customer, err = s.storages.Customer.CreateCustomer(&entity.Customer{
			VendorCustomerID: vendorCustomerID,
		})
		if err != nil {
			logger.Error("failed to create customer in storage", "err", err)
			return "", fmt.Errorf("failed to create customer in storage: %w", err)
		}
		logger.Debug("created customer", "createdCustomer", customer)
	} else {
		logger.Debug("got customer", "customer", customer)
	}

	// Generate a pair of tokens
	tokens, err := s.GenerateCustomerTokens(ctx, customer.ID)
	if err != nil {
		logger.Error("failed to generate tokens", "err", err)
		return "", fmt.Errorf("failed to generate tokens: %w", err)
	}

	customer, err = s.storages.Customer.UpdateCustomer(customer.ID, &entity.Customer{
		RefreshToken: tokens.RefreshToken,
	})
	if err != nil {
		logger.Error("failed to update customer", "err", err)
		return "", fmt.Errorf("failed to update customer: %w", err)
	}
	logger.Debug("updated customer", "customer", customer)

	logger.Info("customer logged in", "tokens", tokens)
	return tokens.AccessToken, nil
}

func (s *customerService) RefreshCustomerAccessToken(ctx context.Context, accessToken string) (string, error) {
	logger := s.logger.
		Named("RefreshCustomerAccessToken").
		WithContext(ctx).
		With("token", accessToken)

	accessTokenClaims, err := token.VerifyJWTToken(token.VerifyJWTTokenOptions{
		Token:                accessToken,
		Secret:               s.cfg.Auth.TokenSecretKey,
		NotToCheckExpiration: true,
	})
	if err != nil {
		logger.Info("invalid token", "err", err)
		return "", ErrRefreshCustomerTokenInvalidToken
	}
	var claimsData token.UserDataClaims
	if err := mapstructure.Decode(accessTokenClaims.GetPayload(), &claimsData); err != nil {
		return "", ErrRefreshCustomerTokenInvalidToken
	}

	customer, err := s.storages.Customer.GetCustomer(GetCustomerFilter{ID: &claimsData.UserID})
	if err != nil {
		logger.Error("failed to get customer", "err", err)
		return "", fmt.Errorf("failed to get customer: %w", err)
	}
	if customer == nil {
		logger.Info("customer not found")
		return "", ErrRefreshCustomerTokenCustomerNotFound
	}
	logger.Debug("got customer", "customer", customer)

	refreshAccessTokenClaims, err := token.VerifyJWTToken(token.VerifyJWTTokenOptions{
		Token:  customer.RefreshToken,
		Secret: s.cfg.Auth.TokenSecretKey,
	})
	if err != nil {
		logger.Info("invalid token", "err", err)
		return "", ErrRefreshCustomerTokenInvalidToken
	}
	if refreshAccessTokenClaims.ExpAt.Before(time.Now()) {
		logger.Info("token expired")
		return "", ErrRefreshCustomerTokenTokenExpired
	}

	t := time.Now()
	accessToken, err = token.SignJWTToken(
		&token.UniversalClaims{
			Iss:   s.cfg.Auth.TokenIssuer,
			ExpAt: t.Add(s.cfg.Auth.AccessTokenLifetime),
			NbfAt: t,
			IssAt: t,
			Payload: token.UserDataClaims{
				UserID: customer.ID,
			},
		},
		s.cfg.Auth.TokenSecretKey,
	)
	if err != nil {
		logger.Error("failed to generate access token", "err", err)
		return "", fmt.Errorf("failed to generate access token: %w", err)
	}
	logger.Debug("access token generated", "accessToken", accessToken)

	logger.Info("successfully refreshed token", "token", accessToken)
	return accessToken, nil
}

func (s *customerService) VerifyCustomerAccessToken(ctx context.Context, accessTokenStr string) (*entity.Customer, error) {
	logger := s.logger.
		Named("VerifyCustomerAccessToken").
		WithContext(ctx).
		With("accessTokenStr", accessTokenStr)

	accessTokenClaims, err := token.VerifyJWTToken(token.VerifyJWTTokenOptions{
		Token:  accessTokenStr,
		Secret: s.cfg.Auth.TokenSecretKey,
	})
	if err != nil {
		logger.Info("invalid token", "err", err)
		return nil, ErrVerifyCustomerTokenInvalidToken
	}
	var claimsData token.UserDataClaims
	if err := mapstructure.Decode(accessTokenClaims.GetPayload(), &claimsData); err != nil {
		return nil, ErrVerifyCustomerTokenInvalidToken
	}
	if accessTokenClaims.ExpAt.Before(time.Now()) {
		logger.Info("token expired")
		return nil, ErrVerifyCustomerTokenTokenExpired
	}

	customer, err := s.storages.Customer.GetCustomer(GetCustomerFilter{ID: &claimsData.UserID})
	if err != nil {
		logger.Error("failed to get customer", "err", err)
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}
	if customer == nil {
		logger.Info("customer not found")
		return nil, ErrVerifyCustomerTokenCustomerNotFound
	}

	logger.Info("verified token", "customer", customer)
	return customer, nil
}

func (ts *customerService) GenerateCustomerTokens(ctx context.Context, customerID string) (GenerateCustomerTokensOutput, error) {
	logger := ts.logger.
		Named("GenerateCustomerToken").
		WithContext(ctx).
		With("customerID", customerID)

	// Create new Access token
	t := time.Now()
	accessToken, err := token.SignJWTToken(
		&token.UniversalClaims{
			Iss:   ts.cfg.Auth.TokenIssuer,
			ExpAt: t.Add(ts.cfg.Auth.AccessTokenLifetime),
			NbfAt: t,
			IssAt: t,
			Payload: token.UserDataClaims{
				UserID: customerID,
			},
		},
		ts.cfg.Auth.TokenSecretKey,
	)
	if err != nil {
		logger.Error("failed to generate access token", "err", err)
		return GenerateCustomerTokensOutput{}, fmt.Errorf("failed to generate access token: %w", err)
	}
	logger.Debug("access token generated", "accessToken", accessToken)

	// Create new Refresh token
	t = time.Now()
	refreshToken, err := token.SignJWTToken(
		&token.UniversalClaims{
			Iss:   ts.cfg.Auth.TokenIssuer,
			ExpAt: t.Add(ts.cfg.Auth.RefreshTokenLifetime),
			NbfAt: t,
			IssAt: t,
			Payload: token.UserDataClaims{
				UserID: customerID,
			},
		},
		ts.cfg.Auth.TokenSecretKey,
	)
	if err != nil {
		logger.Error("failed to generate refresh token", "err", err)
		return GenerateCustomerTokensOutput{}, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	logger.Info("refresh token generated", "refreshToken", refreshToken)
	return GenerateCustomerTokensOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *customerService) GetCustomer(ctx context.Context, opts GetCustomerOptions) (*entity.Customer, error) {
	logger := s.logger.
		Named("GetCustomer").
		WithContext(ctx).
		With("opts", opts)

	customer, err := s.storages.Customer.GetCustomer(GetCustomerFilter{
		ID: &opts.ID,
	})
	if err != nil {
		logger.Error("failed to get customer", "err", err)
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}
	if customer == nil {
		logger.Info("customer not found")
		return nil, ErrGetCustomerCustomerNotFoundInStorage
	}
	logger.Debug("got customer", "customer", customer)

	store, err := s.storages.Store.GetStore(&opts.StoreVendorID)
	if err != nil {
		logger.Error("failed to get store", "err", err)
		return nil, fmt.Errorf("failed to get store: %w", err)
	}
	if store == nil {
		logger.Info("store not found")
		return nil, ErrGetCustomerStoreNotFound
	}

	vendorCustomer, err := s.apis.VendorAPI.WithStore(store).GetCustomerByVendorID(ctx, customer.VendorCustomerID)
	if err != nil {
		logger.Error("failed to get customer from vendor", "err", err)
		return nil, fmt.Errorf("failed to get customer from vendor: %w", err)
	}
	if vendorCustomer == nil {
		logger.Info("customer not found in vendor")
		return nil, ErrGetCustomerCustomerNotFoundInVendor
	}
	logger.Debug("got customer from vendor", "vendorCustomer", vendorCustomer)

	customer.FirstName = vendorCustomer.FirstName

	logger.Info("successfully got customer")
	return customer, nil
}
