package decision

import "github.com/flagship-io/flagship-go-sdk/v3/pkg/model"

// ClientInterface is the modification engine interface
type ClientInterface interface {
	GetModifications(visitorID string, anonymousID *string, context model.Context) (*model.APIClientResponse, error)
}
