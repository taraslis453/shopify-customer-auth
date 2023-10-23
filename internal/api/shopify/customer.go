package shopify

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/taraslis453/shopify-customer-auth/internal/entity"
	"github.com/taraslis453/shopify-customer-auth/internal/service"
)

func (s *shopifyAPI) GetLoggedInCustomerID(ctx context.Context, opt service.LoginCustomerOptions) (string, error) {
	logger := s.logger.Named("GetLoggedInCustomerID").With("email", opt.Email)

	query := `
	mutation SignInWithEmailAndPassword(
	    $email: String!, 
	    $password: String!,
	) {
	    customerAccessTokenCreate(input: { 
	        email: $email, 
	        password: $password,
	    }) {
	        customerAccessToken {
            accessToken
	        }
	        customerUserErrors {
	            code
	            message
	        }
	    }
	}`

	variables := map[string]interface{}{
		"email":    opt.Email,
		"password": opt.Password,
	}

	var customerLoginData struct {
		Data struct {
			CustomerAccessTokenCreate struct {
				CustomerAccessToken struct {
					AccessToken string `json:"accessToken"`
				} `json:"customerAccessToken"`
				CustomerUserErrors []struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				} `json:"customerUserErrors"`
			} `json:"customerAccessTokenCreate"`
		} `json:"data"`
	}

	resp, err := s.graphQL.R().
		SetBody(map[string]interface{}{
			"query":     query,
			"variables": variables,
		}).
		SetResult(&customerLoginData).
		Post("")
	if err != nil {
		logger.Error("failed to login customer", "err", err)
		return "", err
	}
	if resp.StatusCode() != http.StatusOK {
		logger.Error("failed to login customer", "response", resp.String())
		return "", fmt.Errorf("failed to login customer: http status %d, body %s", resp.StatusCode(), resp.String())
	}
	if len(customerLoginData.Data.CustomerAccessTokenCreate.CustomerUserErrors) > 0 && customerLoginData.Data.CustomerAccessTokenCreate.CustomerUserErrors[0].Code == "UNIDENTIFIED_CUSTOMER" {
		logger.Info("invalid email or password")
		return "", service.ErrLoginCustomerInvalidEmailOrPassword
	}

	query = `
  query GetCustomerByAccessToken($accessToken: String!) {
    customer(customerAccessToken: $accessToken) {
      id
    }
  }`

	variables = map[string]interface{}{
		"accessToken": customerLoginData.Data.CustomerAccessTokenCreate.CustomerAccessToken.AccessToken,
	}

	var customerData struct {
		Data struct {
			Customer struct {
				ID string `json:"id"`
			} `json:"customer"`
		} `json:"data"`
	}

	resp, err = s.graphQL.R().
		SetBody(map[string]interface{}{
			"query":     query,
			"variables": variables,
		}).
		SetResult(&customerData).
		Post("")
	if err != nil {
		logger.Error("failed to get customer by access token", "err", err)
		return "", err
	}
	if resp.StatusCode() != http.StatusOK {
		logger.Error("failed to get customer by access token", "response", resp.String())
		return "", fmt.Errorf("failed to get customer by access token: http status %d, body %s", resp.StatusCode(), resp.String())
	}

	customerID := strings.Replace(customerData.Data.Customer.ID, "gid://shopify/Customer/", "", 1)

	logger.Info("successfully logged in customer")
	return customerID, nil
}

// ShopifyCustomer contains information about customer.
type ShopifyCustomer struct {
	FirstName string `json:"first_name,omitempty"`
}

// GetCustomerByVendorID is used to get customer by given vendor id.
func (v *shopifyAPI) GetCustomerByVendorID(ctx context.Context, id string) (*entity.Customer, error) {
	logger := v.logger.
		Named("GetCustomerByID").
		WithContext(ctx).
		With("id", id)

	url := fmt.Sprintf("/admin/api/2023-01/customers/%s.json", id)

	var customer struct {
		Customer ShopifyCustomer `json:"customer"`
	}

	res, err := v.http.R().
		SetResult(&customer).
		Get(url)
	if err != nil {
		logger.Error("failed to send get shopify customer request", "err", err)
		return nil, fmt.Errorf("failed to send get shopify customer request: %w", err)
	}
	if res.StatusCode() != http.StatusOK {
		logger.Error("failed to get shopify customer", "resBody", res.String())
		return nil, fmt.Errorf("failed to get shopify customer: http status %d, body %s", res.StatusCode(), res.String())
	}
	logger = logger.With("resBody", res.String())

	logger.Info("successfully got shopify customer")
	return &entity.Customer{
		FirstName: customer.Customer.FirstName,
	}, nil
}
