package chi_paddle_subscription

import (
	"context"

	"github.com/PaddleHQ/paddle-go-sdk/v4"
)

type SubscriptionService interface {
	CancelSubscription(ctx context.Context, subscriptionID string) (*paddle.Subscription, error)
	UpdateSubscriptionItems(ctx context.Context, subscriptionID, priceID string) (*paddle.Subscription, error)
	GetSubscription(ctx context.Context, subscriptionID string) (*paddle.Subscription, error)
}

type subscriptionService struct {
	sdk *paddle.SDK
}

func New(sdk *paddle.SDK) SubscriptionService {
	return &subscriptionService{sdk: sdk}
}

// CancelSubscription cancels a Paddle subscription by ID.
func (s *subscriptionService) CancelSubscription(ctx context.Context, subscriptionID string) (*paddle.Subscription, error) {
	return s.sdk.SubscriptionsClient.CancelSubscription(ctx, &paddle.CancelSubscriptionRequest{
		SubscriptionID: subscriptionID,
	})
}

// UpdateSubscriptionItems changes subscription items to a new price (immediate plan change).
func (s *subscriptionService) UpdateSubscriptionItems(ctx context.Context, subscriptionID, priceID string) (*paddle.Subscription, error) {
	return s.sdk.SubscriptionsClient.UpdateSubscription(ctx, &paddle.UpdateSubscriptionRequest{
		SubscriptionID: subscriptionID,
		Items: paddle.NewPatchField([]paddle.UpdateSubscriptionItems{
			*paddle.NewUpdateSubscriptionItemsSubscriptionUpdateItemFromCatalog(&paddle.SubscriptionUpdateItemFromCatalog{
				PriceID:  priceID,
				Quantity: 1,
			}),
		}),
		ProrationBillingMode: paddle.NewPatchField(paddle.ProrationBillingModeProratedImmediately),
	})
}

// GetSubscription loads a subscription by ID.
func (s *subscriptionService) GetSubscription(ctx context.Context, subscriptionID string) (*paddle.Subscription, error) {
	return s.sdk.SubscriptionsClient.GetSubscription(ctx, &paddle.GetSubscriptionRequest{SubscriptionID: subscriptionID})
}
