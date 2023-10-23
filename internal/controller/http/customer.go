package httpcontroller

import (
	"github.com/gin-gonic/gin"

	"github.com/taraslis453/shopify-customer-auth/internal/service"
	"github.com/taraslis453/shopify-customer-auth/pkg/errs"
)

type customerRoutes struct {
	routerContext
}

func newCustomerRoutes(options RouterOptions) {
	r := &customerRoutes{
		routerContext{
			services: options.Services,
			logger:   options.Logger.Named("customerRoutes"),
			cfg:      options.Config,
		},
	}

	p := options.Handler.Group("/customers")
	{
		p.POST("/login", errorHandler(options, r.loginCustomer))
		p.POST("/refresh-token", errorHandler(options, r.refreshToken))
		p.GET("/me", newAuthMiddleware(options), errorHandler(options, r.getCustomer))
	}
}

type loginCustomerRequestBody struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type loginCustomerResponse struct {
	AccessToken string `json:"accessToken"`
}

func (r *customerRoutes) loginCustomer(c *gin.Context) (interface{}, *httpErr) {
	logger := r.logger.Named("loginCustomer").WithContext(c)

	var body loginCustomerRequestBody
	if err := c.ShouldBindJSON(&body); err != nil {
		logger.Info("failed to parse request body", "err", err)
		return nil, &httpErr{Type: httpErrTypeClient, Message: "invalid request body", Details: err}
	}
	logger = logger.With("body", body)
	logger.Info("successfully parsed request body")

	accessToken, err := r.services.Customer.LoginCustomer(c, service.LoginCustomerOptions{
		Email:    body.Email,
		Password: body.Password,
		// https://taras-store88.myshopify.com => taras-store88.myshopify.com
		StoreVendorID: c.GetHeader("Origin")[8:],
	})
	if err != nil {
		if errs.IsExpected(err) {
			logger.Info(err.Error())
			return nil, &httpErr{Type: httpErrTypeClient, Message: err.Error(), Code: errs.GetCode(err)}
		}

		logger.Error("failed to login user", "err", err)
		return nil, &httpErr{Type: httpErrTypeServer, Message: "failed to login user", Details: err}
	}

	logger.Info("successfully logged in user")
	return loginCustomerResponse{
		AccessToken: accessToken,
	}, nil
}

type refreshTokenResponseBody struct {
	AccessToken string `json:"accessToken"`
}

func (r *customerRoutes) refreshToken(c *gin.Context) (interface{}, *httpErr) {
	logger := r.logger.Named("refreshToken").WithContext(c)

	token, err := getAuthToken(c.GetHeader("Authorization"))
	if err != nil {
		logger.Info(err.Error())
		return nil, &httpErr{Type: httpErrTypeClient, Message: err.Error()}
	}

	refreshedToken, err := r.services.Customer.RefreshCustomerAccessToken(c, token)
	if err != nil {
		if errs.IsExpected(err) {
			logger.Info(err.Error())
			return nil, &httpErr{Type: httpErrTypeClient, Message: err.Error(), Code: errs.GetCode(err)}
		}
		logger.Error("failed to refresh token", "err", err)
		return nil, &httpErr{Type: httpErrTypeServer, Message: "failed to refresh token", Details: err}
	}

	logger.Info("token successfully refreshed")
	return refreshTokenResponseBody{
		AccessToken: refreshedToken,
	}, nil
}

type getCustomerResponseBody struct {
	ID        string `json:"id"`
	FirstName string `json:"firstName"`
}

func (r *customerRoutes) getCustomer(c *gin.Context) (interface{}, *httpErr) {
	logger := r.logger.Named("getCustomer").WithContext(c)

	customerID := c.GetString("userID")
	logger = logger.With("customerID", customerID)

	customer, err := r.services.Customer.GetCustomer(c, service.GetCustomerOptions{
		ID: customerID,
	})
	if err != nil {
		if errs.IsExpected(err) {
			logger.Info(err.Error())
			return nil, &httpErr{Type: httpErrTypeClient, Message: err.Error(), Code: errs.GetCode(err)}
		}
		logger.Error("failed to get customer", "err", err)
		return nil, &httpErr{Type: httpErrTypeServer, Message: "failed to get customer", Details: err}
	}

	logger.Info("successfully got customer")
	return getCustomerResponseBody{
		ID:        customer.ID,
		FirstName: customer.FirstName,
	}, nil
}
