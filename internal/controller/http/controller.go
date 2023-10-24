package httpcontroller

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	// third party
	"github.com/DataDog/gostackparse"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/taraslis453/shopify-customer-auth/config"
	"github.com/taraslis453/shopify-customer-auth/internal/service"
	"github.com/taraslis453/shopify-customer-auth/pkg/errs"
	"github.com/taraslis453/shopify-customer-auth/pkg/logging"
)

// routerContext provides a shared context for all routers.
type routerContext struct {
	services service.Services
	logger   logging.Logger
	cfg      *config.Config
}

// RouterOptions provides shared options for all routers.
type RouterOptions struct {
	Handler  *gin.RouterGroup
	Services service.Services
	Logger   logging.Logger
	Config   *config.Config
}

// Options is used to parameterize http controller via New.
type Options struct {
	Handler  *gin.Engine
	Services service.Services
	Logger   logging.Logger
	Config   *config.Config
}

// New is used to create new http controller.
func New(options Options) {
	// options
	options.Handler.Use(gin.Logger(), gin.Recovery(), requestIDMiddleware, corsMiddleware)

	routerOptions := RouterOptions{
		Handler:  options.Handler.Group(""),
		Services: options.Services,
		Logger:   options.Logger.Named("HTTPController"),
		Config:   options.Config,
	}

	options.Handler.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })
	{
		newCustomerRoutes(routerOptions)
		newVendorRoutes(routerOptions)
	}
}

// httpErr provides a base error type for all http controller errors
type httpErr struct {
	Type    httpErrType `json:"-"`
	Code    string      `json:"code,omitempty"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// httpErrType is used to define httpErr type
type httpErrType string

const (
	// httpErrTypeServer is an "unexpected" internal server error
	httpErrTypeServer httpErrType = "server"
	// httpErrTypeClient is an "expected" business error
	httpErrTypeClient httpErrType = "client"
)

// Error is used to convert an httpErr to a string
func (err httpErr) Error() string {
	return fmt.Sprintf("%s: %s", err.Type, err.Message)
}

// errorHandler provides unified error handling for all handlers.
func errorHandler(options RouterOptions, handler func(c *gin.Context) (interface{}, *httpErr)) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := options.Logger.Named("errorHandler")

		// handle panics
		defer func() {
			if err := recover(); err != nil {
				// get stacktrace
				stacktrace, errors := gostackparse.Parse(bytes.NewReader(debug.Stack()))
				if len(errors) > 0 || len(stacktrace) == 0 {
					logger.Error("get stacktrace errors", "stacktraceErrors", errors, "stacktrace", "unknown", "err", err)
				} else {
					logger.Error("unhandled error", "err", err, "stacktrace", stacktrace)
				}

				err := c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("%v", err))
				if err != nil {
					logger.Error("failed to abort with error", "err", err)
				}
			}
		}()

		// execute handler
		body, err := handler(c)
		logger = logger.With("body", body, "err", err)

		// check if middleware
		if body == nil && err == nil {
			logger.Info("body and err are nil")
			return
		}

		if err != nil {
			if err.Type == httpErrTypeServer {
				logger.Error("internal server error")
				c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			} else {
				logger.Info("client error")
				c.AbortWithStatusJSON(http.StatusUnprocessableEntity, err)
			}
			return
		}

		logger.Info("request handled")
		c.JSON(http.StatusOK, body)
	}
}

// corsMiddleware - used to allow incoming cross-origin requests.
func corsMiddleware(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "*")
	c.Header("Access-Control-Allow-Headers", "*")
	c.Header("Content-Type", "application/json")
	if c.Request.Method != "OPTIONS" {
		c.Next()
	} else {
		c.AbortWithStatus(http.StatusOK)
	}
}

// newAuthMiddleware is used to get auth token from request headers and validate it.
func newAuthMiddleware(options RouterOptions) gin.HandlerFunc {
	logger := options.Logger.Named("authMiddleware")

	return errorHandler(options, func(c *gin.Context) (interface{}, *httpErr) {
		rawToken := c.GetHeader("Authorization")

		token, err := getAuthToken(rawToken)
		if err != nil {
			logger.Info(err.Error())
			return nil, &httpErr{Type: httpErrTypeClient, Message: err.Error()}
		}

		customer, err := options.Services.Customer.VerifyCustomerAccessToken(c, token)
		if err != nil {
			if errs.IsExpected(err) {
				logger.Info(err.Error())
				return nil, &httpErr{Type: httpErrTypeClient, Code: fmt.Sprint(http.StatusUnauthorized), Message: err.Error()}
			}

			logger.Error("failed to verify access token")
			return nil, &httpErr{
				Type:    httpErrTypeServer,
				Message: "failed to verify access token",
				Details: err,
			}
		}
		logger = logger.With("customer", customer)
		logger.Debug("verified token")

		c.Set("userID", customer.ID)

		logger.Info("successfully validated auth token")
		return nil, nil
	})
}

func getAuthToken(rawToken string) (string, error) {
	if rawToken == "" {
		return "", fmt.Errorf("empty auth token")
	}

	// Split Bearer and token
	splitRawToken := strings.Split(rawToken, " ")
	if len(splitRawToken) != 2 {
		return "", fmt.Errorf("malformed auth token")
	}

	token := splitRawToken[1]

	return token, nil
}

// requestIDMiddleware is used to add request id to gin context.
func requestIDMiddleware(c *gin.Context) {
	c.Set("RequestID", uuid.NewString())
}
