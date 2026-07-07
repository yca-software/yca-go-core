package chi_paddle_customer

// CustomerInput is the data required to create or update a Paddle customer.
type CustomerInput struct {
	OrganizationID string
	Name           string
	BillingEmail   string
	Address        string
	City           string
	Zip            string
	Country        string
	Timezone       string
}

// CustomerPortalSessionResult is returned after creating a customer portal session.
type CustomerPortalSessionResult struct {
	URL string `json:"url"`
}
