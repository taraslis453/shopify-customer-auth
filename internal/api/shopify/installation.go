package shopify

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"

	"github.com/taraslis453/shopify-customer-auth/internal/entity"
)

// HandleInstall handles an oauth2 installation call.
func (v *shopifyAPI) HandleInstall(c *gin.Context, redirectURL string) (*entity.Store, string, error) {
	logger := v.logger.
		Named("HandleInstall").
		With("redirectURL", redirectURL)

	if v.store == nil {
		logger.Error("missing store vendor config")
		return nil, "", fmt.Errorf("missing store vendor config")
	}

	values := url.Values{
		"client_id":       {v.store.ClientID},
		"scope":           {"read_products,write_products,unauthenticated_read_content,unauthenticated_read_customer_tags,unauthenticated_read_product_tags,unauthenticated_read_product_listings,unauthenticated_write_checkouts,unauthenticated_read_checkouts,unauthenticated_write_customers,unauthenticated_read_customers,read_customers"},
		"redirect_uri":    {redirectURL},
		"state":           {""},        // nonce
		"grant_options[]": {"offline"}, // https://shopify.dev/concepts/about-apis/authentication#api-access-modes
	}
	logger = logger.With("values", values.Encode())

	logger.Info("installation link successfully created")
	return v.store, fmt.Sprintf(
		"https://%s/admin/oauth/authorize?%s",
		v.store.VendorID,
		values.Encode(),
	), nil
}

type shopifyRedirectQuery struct {
	Code      string `form:"code" json:"code" binding:"required"`
	HMAC      string `form:"hmac" json:"hmac" binding:"required"`
	Host      string `form:"host" json:"host" binding:"required"`
	Shop      string `form:"shop" json:"shop" binding:"required"`
	State     string `form:"state" json:"state"`
	Timestamp string `form:"timestamp" json:"timestamp" binding:"required"`
}

// HandleRedirect handles an oauth2 redirect call.
func (v *shopifyAPI) HandleRedirect(c *gin.Context) (*entity.Store, error) {
	logger := v.logger.Named("HandleRedirect").WithContext(c)

	if v.store == nil {
		logger.Error("missing store vendor config")
		return nil, fmt.Errorf("missing store vendor config")
	}

	var query shopifyRedirectQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		logger.Error("invalid request query", "err", err)
		return nil, fmt.Errorf("invalid request query: %w", err)
	}
	logger = logger.With("query", query)
	logger.Debug("request query parsed")

	// get access token
	var credentials map[string]string
	res, err := v.http.R().
		SetQueryParams(map[string]string{
			"client_id":     v.store.ClientID,
			"client_secret": v.store.ClientSecret,
			"code":          query.Code,
		}).
		SetResult(&credentials).
		Post(fmt.Sprintf("https://%s/admin/oauth/access_token", v.store.VendorID))
	if err != nil {
		logger.Error("failed to get shopify access token", "err", err)
		return nil, fmt.Errorf("failed to get shopify access token: %w", err)
	}
	if res.StatusCode() != http.StatusOK {
		logger.Error("failed to get shopify access token", "resBody", res.String())
		return nil, fmt.Errorf("failed to get shopify access token: http status %d, body %s", res.StatusCode(), res.String())
	}
	logger = logger.With("resBody", res.String())
	logger.Debug("got credentials")

	// update config
	v.store.AccessToken = credentials["access_token"]
	v.store.Scope = credentials["scope"]

	storeFrontAccessToken, err := v.getStoreFrontAccessToken()
	if err != nil {
		logger.Error("failed to get StoreFront access token", "err", err)
		return nil, fmt.Errorf("failed to get StoreFront access token: %w", err)
	}
	v.store.StoreFrontAccessToken = storeFrontAccessToken

	logger.Info("successfully got credentials")
	return v.store, nil
}

// getStoreFrontAccessToken gets a StoreFront access token for the store.
func (v *shopifyAPI) getStoreFrontAccessToken() (string, error) {
	logger := v.logger.Named("GetStoreFrontAccessToken")

	var credentials struct {
		StoreFrontAccessToken struct {
			AccessToken string `json:"access_token"`
		} `json:"storefront_access_token"`
	}

	res, err := v.http.R().
		SetHeader("X-Shopify-Access-Token", v.store.AccessToken).
		SetBody(map[string]interface{}{
			"storefront_access_token": map[string]interface{}{
				"title": "StoreFront Access Token",
			},
		}).
		SetResult(&credentials).
		Post(fmt.Sprintf("https://%s/admin/api/2023-04/storefront_access_tokens.json", v.store.VendorID))
	if err != nil {
		logger.Error("failed to send get StoreFront access token request", "err", err)
		return "", fmt.Errorf("failed to send get StoreFront access token request: %w", err)
	}
	if res.StatusCode() != http.StatusOK {
		logger.Error("failed to get StoreFront access token", "resBody", res.String())
		return "", fmt.Errorf("failed to get StoreFront access token: http status %d, body %s", res.StatusCode(), res.String())

	}
	logger = logger.With("resBody", res.String())

	token := credentials.StoreFrontAccessToken.AccessToken
	logger = logger.With("token", token)

	logger.Info("successfully got StoreFront access token")
	return token, nil
}
