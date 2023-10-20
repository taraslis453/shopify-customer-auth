package shopify

import (
	"context"
	"fmt"
	"net/http"

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
              customer {
                id
              }
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

	var data struct {
		CustomerAccessTokenCreate struct {
			CustomerAccessToken struct {
				Customer struct {
					ID string `json:"id"`
				} `json:"customer"`
			} `json:"customerAccessToken"`
			CustomerUserErrors []struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"customerUserErrors"`
		} `json:"customerAccessTokenCreate"`
	}

	resp, err := s.graphQL.R().
		SetBody(map[string]interface{}{
			"query":     query,
			"variables": variables,
		}).
		SetResult(&data).
		Post("")
	if err != nil {
		logger.Error("failed to login customer", "err", err)
		return "", err
	}
	if resp.StatusCode() != http.StatusOK {
		logger.Error("failed to login customer", "response", resp.String())
		return "", err
	}
	if len(data.CustomerAccessTokenCreate.CustomerUserErrors) > 0 && data.CustomerAccessTokenCreate.CustomerUserErrors[0].Code == "UNIDENTIFIED_CUSTOMER" {
		logger.Info("invalid email or password")
		return "", service.ErrLoginCustomerInvalidEmailOrPassword
	}

	logger.Debug("got customer", "customer", data.CustomerAccessTokenCreate.CustomerAccessToken.Customer)
	return data.CustomerAccessTokenCreate.CustomerAccessToken.Customer.ID, nil
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
