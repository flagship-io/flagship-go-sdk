package tracking

import "github.com/abtasty/flagship-go-sdk/pkg/model"

// APIClientInterface sends a hit to the data collect
type APIClientInterface interface {
	SendHit(visitorID string, hit model.HitInterface) error
	ActivateCampaign(request model.ActivationHit) error
	SendEvent(request model.Event) error
}
