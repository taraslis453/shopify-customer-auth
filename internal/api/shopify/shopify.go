package shopify

import (
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/taraslis453/shopify-customer-auth/config"
	"github.com/taraslis453/shopify-customer-auth/internal/entity"
	"github.com/taraslis453/shopify-customer-auth/internal/service"
	"github.com/taraslis453/shopify-customer-auth/pkg/logging"
)

// shopifyAPI implements the service.VendorAPI interface.
type shopifyAPI struct {
	http    *resty.Client
	graphQL *resty.Client
	store   *entity.Store
	logger  logging.Logger
	cfg     *config.Config
}

// Options is used to parameterize VendorShopify using New.
type Options struct {
	Logger logging.Logger
	Config *config.Config
}

var _ service.VendorAPI = (*shopifyAPI)(nil)

// New is used to create a new shopifyAPI instance.
func New(options *Options) *shopifyAPI {
	return &shopifyAPI{
		logger: options.Logger.Named("ShopifyAPI"),
		cfg:    options.Config,
	}
}

// WithStore implements service.VendorAPI.
func (v *shopifyAPI) WithStore(store *entity.Store) service.VendorAPI {
	var h *resty.Client
	var g *resty.Client

	shopName := strings.Split(store.VendorID, ".")[0]

	h = resty.New().
		SetBaseURL(fmt.Sprintf(`https://%s.myshopify.com`, shopName)).
		SetHeader("X-Shopify-Access-Token", store.AccessToken).
		SetHeader("Content-Type", "application/json")

	g = resty.New().
		SetBaseURL(fmt.Sprintf(`https://%s.myshopify.com/api/2023-01/graphql.json`, shopName)).
		SetHeader("Shopify-Storefront-Private-Token", store.GraphAPIAccessToken).
		SetHeader("Content-Type", "application/json")

	return &shopifyAPI{
		store:   store,
		http:    h,
		graphQL: g,
		logger:  v.logger,
		cfg:     v.cfg,
	}
}
