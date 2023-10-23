package httpcontroller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/taraslis453/shopify-customer-auth/pkg/errs"
)

// vendorRouter represents vendor service router.
type vendorRouter struct {
	routerContext
}

// newVendorRoutes is used to setup vendor routes.
func newVendorRoutes(options RouterOptions) {
	r := &vendorRouter{
		routerContext{
			services: options.Services,
			logger:   options.Logger.Named("vendorRoutes"),
			cfg:      options.Config,
		},
	}

	p := options.Handler.Group("/vendors")
	{
		p.GET("/install", errorHandler(options, r.installHandler))
		p.GET("/redirect", errorHandler(options, r.redirectHandler))
	}
}

func (r *vendorRouter) installHandler(c *gin.Context) (interface{}, *httpErr) {
	logger := r.logger.Named("installHandler").WithContext(c)

	redirectURL, err := r.services.Vendor.HandleInstall(c)
	if err != nil {
		if errs.IsExpected(err) {
			logger.Info(err.Error())
			return nil, &httpErr{Message: err.Error(), Code: errs.GetCode(err)}
		}
		logger.Error("failed to handle oauth2 installation call", "err", err)
		return nil, &httpErr{Message: "failed to handle oauth2 installation call", Details: err}
	}
	logger = logger.With("redirectURL", redirectURL)

	logger.Info("handling oauth2 redirect call")
	c.Redirect(http.StatusFound, redirectURL)
	return nil, nil
}

func (r *vendorRouter) redirectHandler(c *gin.Context) (interface{}, *httpErr) {
	logger := r.logger.Named("redirectHandler").WithContext(c)

	redirectURL, err := r.services.Vendor.HandleRedirect(c)
	if err != nil {
		if errs.IsExpected(err) {
			logger.Info(fmt.Sprintf("didn't handle vendor redirect: %s", err.Error()))
			return nil, &httpErr{Message: err.Error(), Code: errs.GetCode(err)}
		}
		logger.Error("failed to handle vendor redirect", "err", err)
		return nil, &httpErr{Message: "failed to handle vendor redirect", Details: err}
	}
	if redirectURL != "" {
		logger.Info("handling oauth2 redirect call")
		c.Redirect(http.StatusFound, redirectURL)
		return nil, nil
	}
	logger = logger.With("redirectURL", redirectURL)

	logger.Info("successfully installed the app")
	return map[string]string{
		"message": "successfully installed the app",
	}, nil
}
