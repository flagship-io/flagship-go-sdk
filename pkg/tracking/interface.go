package tracking

import "github.com/flagship-io/flagship-go-sdk/v3/pkg/model"

// APIClientInterface sends a hit to the data collect
type APIClientInterface interface {
	SendHit(visitorID string, anonymousID *string, hit model.HitInterface) error
	ActivateCampaign(request model.ActivationHit) error
	SendEvent(request model.Event) error
}
