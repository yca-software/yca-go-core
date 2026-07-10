package yca_paddle_customer

import (
	"context"
	"errors"

	"github.com/PaddleHQ/paddle-go-sdk/v4"
)

type CustomerService interface {
	CreateCustomer(ctx context.Context, req CustomerInput) (*paddle.Customer, error)
	UpdateCustomer(ctx context.Context, customerID string, req CustomerInput) (*paddle.Customer, error)
	ArchiveCustomer(ctx context.Context, customerID string) error
	CreateCustomerPortalSession(ctx context.Context, customerID string) (*CustomerPortalSessionResult, error)
}

type customerService struct {
	sdk *paddle.SDK
}

func New(sdk *paddle.SDK) CustomerService {
	return &customerService{sdk: sdk}
}

func (c *customerService) CreateCustomer(ctx context.Context, req CustomerInput) (*paddle.Customer, error) {
	customer, err := c.sdk.CustomersClient.CreateCustomer(ctx, &paddle.CreateCustomerRequest{
		Email: req.BillingEmail,
		Name:  &req.Name,
		CustomData: paddle.CustomData{
			"organization_id": req.OrganizationID,
			"address":         req.Address,
			"city":            req.City,
			"zip":             req.Zip,
			"country":         req.Country,
			"timezone":        req.Timezone,
		},
	})
	if err == nil {
		return customer, nil
	}

	if !isCustomerAlreadyExistsError(err) {
		return nil, err
	}

	existing, resolveErr := c.resolveCustomerForDuplicateEmail(ctx, req.BillingEmail, err)
	if resolveErr != nil {
		return nil, resolveErr
	}
	if existing == nil {
		return nil, err
	}

	if existing.Status != paddle.StatusActive {
		return c.activateCustomer(ctx, existing.ID)
	}
	return existing, nil
}

// UpdateCustomer updates Paddle customer fields for a workspace.
func (c *customerService) UpdateCustomer(ctx context.Context, customerID string, req CustomerInput) (*paddle.Customer, error) {
	return c.sdk.CustomersClient.UpdateCustomer(ctx, &paddle.UpdateCustomerRequest{
		CustomerID: customerID,
		Email:      paddle.NewPatchField(req.BillingEmail),
		Name:       paddle.NewPatchField(&req.Name),
		CustomData: paddle.NewPatchField(paddle.CustomData{
			"organization_id": req.OrganizationID,
			"address":         req.Address,
			"city":            req.City,
			"zip":             req.Zip,
			"country":         req.Country,
			"timezone":        req.Timezone,
		}),
	})
}

// ArchiveCustomer archives a Paddle customer.
func (c *customerService) ArchiveCustomer(ctx context.Context, customerID string) error {
	if customerID == "" {
		return nil
	}

	_, err := c.sdk.CustomersClient.UpdateCustomer(ctx, &paddle.UpdateCustomerRequest{
		CustomerID: customerID,
		Status:     paddle.NewPatchField(paddle.StatusArchived),
	})

	return err
}

// CreateCustomerPortalSession opens a Paddle customer portal for subscription management.
func (c *customerService) CreateCustomerPortalSession(ctx context.Context, customerID string) (*CustomerPortalSessionResult, error) {
	session, err := c.sdk.CustomerPortalSessionsClient.CreateCustomerPortalSession(ctx, &paddle.CreateCustomerPortalSessionRequest{
		CustomerID: customerID,
	})
	if err != nil {
		return nil, err
	}
	url := session.URLs.General.Overview
	if url == "" {
		return nil, errors.New("paddle: portal url missing")
	}
	return &CustomerPortalSessionResult{URL: url}, nil
}

func (c *customerService) resolveCustomerForDuplicateEmail(ctx context.Context, billingEmail string, createErr error) (*paddle.Customer, error) {
	existing, listErr := c.getCustomerByEmail(ctx, billingEmail)
	if listErr != nil {
		return nil, listErr
	}
	if existing != nil {
		return existing, nil
	}
	if id := paddleCustomerIDFromAlreadyExistsError(createErr); id != "" {
		return c.sdk.CustomersClient.GetCustomer(ctx, &paddle.GetCustomerRequest{CustomerID: id})
	}
	return nil, nil
}

func (c *customerService) getCustomerByEmail(ctx context.Context, email string) (*paddle.Customer, error) {
	coll, err := c.sdk.CustomersClient.ListCustomers(ctx, &paddle.ListCustomersRequest{Email: []string{email}})
	if err != nil {
		return nil, err
	}
	var first *paddle.Customer
	_ = coll.Iter(ctx, func(c *paddle.Customer) (bool, error) {
		first = c
		return false, nil
	})
	return first, nil
}

func (c *customerService) activateCustomer(ctx context.Context, customerID string) (*paddle.Customer, error) {
	return c.sdk.CustomersClient.UpdateCustomer(ctx, &paddle.UpdateCustomerRequest{
		CustomerID: customerID,
		Status:     paddle.NewPatchField(paddle.StatusActive),
	})
}
