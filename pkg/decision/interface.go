package decision

import (
	"github.com/flagship-io/flagship-go-sdk/v2/pkg/model"
	"google.golang.org/protobuf/types/known/structpb"
)

// ClientInterface is the modification engine interface
type ClientInterface interface {
	GetModifications(visitorID string, context map[string]*structpb.Value) (*model.APIClientResponse, error)
}
