package service

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/taraslis453/shopify-customer-auth/pkg/errs"
)

type vendorService struct {
	serviceContext
}

var _ VendorService = (*vendorService)(nil)

func NewVendorService(options Options) *vendorService {
	return &vendorService{
		serviceContext: serviceContext{
			apis:     options.APIs,
			cfg:      options.Config,
			logger:   options.Logger.Named("Vendor"),
			storages: options.Storages,
		},
	}
}

func (s *vendorService) HandleInstall(c *gin.Context) (string, error) {
	logger := s.logger.
		Named("HandleInstall").
		WithContext(c)

	vendorID := getVendorIDFromQuery(c)
	if vendorID == "" {
		logger.Info("vendorID not found")
		return "", ErrHandleInstallVendorIDNotFound
	}

	store, err := s.storages.Store.GetStore(&vendorID)
	if err != nil {
		logger.Error("failed to get store", "err", err)
		return "", fmt.Errorf("failed to get store: %w", err)
	}
	if store == nil {
		logger.Info("store not found")
		return "", ErrHandleInstallStoreNotFound
	}
	logger = logger.With("store", store)
	logger.Debug("got store")

	redirectURL := s.cfg.App.BaseURL + "/vendors/redirect"

	newStore, redirectURL, err := s.apis.VendorAPI.WithStore(store).HandleInstall(c, redirectURL)
	if err != nil {
		logger.Error("failed to handle install", "err", err)
		return "", fmt.Errorf("failed to handle install: %w", err)
	}
	logger = logger.With("newStore", newStore).With("redirectURL", redirectURL)

	updatedStore, err := s.storages.Store.UpdateStore(store.ID, newStore)
	if err != nil {
		logger.Error("failed to update store", "err", err)
		return "", fmt.Errorf("failed to update store: %w", err)
	}
	logger = logger.With("updatedStore", updatedStore)

	logger.Info("successfully handled install")
	return redirectURL, nil
}

func (s *vendorService) HandleRedirect(c *gin.Context) (string, error) {
	logger := s.logger.
		Named("HandleRedirect").
		WithContext(c)

	vendorID := getVendorIDFromQuery(c)
	if vendorID == "" {
		logger.Info("vendorID not found")
		return "", ErrHandleRedirectVendorIDNotFound
	}
	logger = logger.With("vendorID", vendorID)
	logger.Debug("got vendor id from query")

	store, err := s.storages.Store.GetStore(&vendorID)
	if err != nil {
		logger.Error("failed to get store", "err", err)
		return "", fmt.Errorf("failed to get store: %w", err)
	}
	if store == nil {
		logger.Info("store config not found")
		return "", ErrHandleRedirectStoreNotFound
	}
	logger = logger.With("store", store)
	logger.Debug("got store")

	newStore, err := s.apis.VendorAPI.WithStore(store).HandleRedirect(c)
	if err != nil {
		if errs.IsExpected(err) {
			logger.Info(err.Error())
			return "", err
		}
		logger.Error("failed to handle redirect", "err", err)
		return "", fmt.Errorf("failed to handle redirect: %w", err)
	}
	logger = logger.With("newVendorConfig", newStore)
	logger.Debug("handled redirect")

	updatedStore, err := s.storages.Store.UpdateStore(newStore.ID, newStore)
	if err != nil {
		logger.Error("failed to update store", "err", err)
		return "", fmt.Errorf("failed to update store: %w", err)
	}
	logger = logger.With("updatedStore", updatedStore)
	logger.Debug("updated store")

	logger.Info("successfully handled redirect")
	return "", nil
}

func getVendorIDFromQuery(c *gin.Context) string {
	// shopify
	if c.Query("shop") != "" {
		return c.Query("shop")
	}

	return ""
}
