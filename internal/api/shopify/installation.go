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
		"scope":           {"read_customers"},
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

	graphAPIAccessToken, err := v.GetGraphAPIAccessToken()
	if err != nil {
		logger.Error("failed to get GraphAPI access token", "err", err)
		return nil, fmt.Errorf("failed to get GraphAPI access token: %w", err)
	}
	v.store.GraphAPIAccessToken = graphAPIAccessToken
	logger = logger.With("updatedConfig", v.store)

	logger.Info("successfully got credentials")
	return v.store, nil
}

// GetGraphAPIAccessToken gets a GraphAPI access token for the store.
// https://shopify.dev/api/usage/authentication#access-tokens-for-the-storefront-api
// https://shopify.dev/apps/auth/oauth/delegate-access-tokens
func (v *shopifyAPI) GetGraphAPIAccessToken() (string, error) {
	logger := v.logger.Named("GetGraphAPIAccessToken")

	var credentials map[string]string

	res, err := v.http.R().
		SetHeader("X-Shopify-Access-Token", v.store.AccessToken).
		SetBody(map[string][]string{
			"delegate_access_scope": {
				"read_customers",
			},
		}).
		SetResult(&credentials).
		Post(fmt.Sprintf("https://%s/admin/access_tokens/delegate.json", v.store.VendorID))
	if err != nil {
		logger.Error("failed to send get GraphAPI access token request", "err", err)
		return "", fmt.Errorf("failed to send get GraphAPI access token request: %w", err)
	}
	if res.StatusCode() != http.StatusOK {
		logger.Error("failed to get GraphAPI access token", "resBody", res.String())
		return "", fmt.Errorf("failed to get GraphAPI access token: http status %d, body %s", res.StatusCode(), res.String())
	}
	logger = logger.With("resBody", res.String())

	token := credentials["access_token"]
	logger = logger.With("token", token)

	logger.Info("successfully got GraphAPI access token")
	return token, nil
}
