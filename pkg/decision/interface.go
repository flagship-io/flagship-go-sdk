package decision

import "github.com/abtasty/flagship-go-sdk/pkg/model"

// ClientInterface is the modification engine interface
type ClientInterface interface {
	GetModifications(visitorID string, context map[string]interface{}) (*model.APIClientResponse, error)
}
