package client

import (
	"errors"
	"fmt"
	"strings"

	"github.com/flagship-io/flagship-go-sdk/v2/pkg/cache"
	"github.com/flagship-io/flagship-go-sdk/v2/pkg/model"

	"github.com/flagship-io/flagship-go-sdk/v2/pkg/bucketing"
	"github.com/flagship-io/flagship-go-sdk/v2/pkg/utils"

	"github.com/flagship-io/flagship-go-sdk/v2/pkg/decision"
	"github.com/flagship-io/flagship-go-sdk/v2/pkg/logging"
	"github.com/flagship-io/flagship-go-sdk/v2/pkg/tracking"
)

// DecisionMode represents the decision mode of the Client engine
type DecisionMode string

// The different decision modes
const (
	API       DecisionMode = "API"
	Bucketing DecisionMode = "Bucketing"
)

const (
	STATUS_INITIALIZING = "INITIALIZING"
	STATUS_READY        = "READY"
)

// Client represent the Flagship SDK client object
type Client struct {
	envID             string
	apiKey            string
	decisionMode      DecisionMode
	decisionClient    decision.ClientInterface
	trackingAPIClient tracking.APIClientInterface
	cacheManager      cache.Manager
	status            string
}

var clientLogger = logging.CreateLogger("FS Client")

// Create creates a Client from options
func Create(f *Options) (*Client, error) {
	clientLogger.Info(fmt.Sprintf("Creating FS Client with Decision Mode : %s", f.decisionMode))

	var err error

	if f.EnvID == "" {
		err = errors.New("EnvID is required")
		return nil, err
	}

	if f.APIKey == "" {
		err = errors.New("APIKey is required")
		return nil, err
	}

	client := &Client{
		envID:             f.EnvID,
		apiKey:            f.APIKey,
		status:            STATUS_INITIALIZING,
		trackingAPIClient: f.trackingAPIClient,
	}

	if len(f.cacheManagerOptions) > 0 {
		cacheManager, err := cache.InitManager(f.cacheManagerOptions...)
		if err != nil {
			clientLogger.Error("Got error when initializing cache", err)
		}
		client.cacheManager = cacheManager
	}

	if client.trackingAPIClient == nil {
		client.trackingAPIClient, err = tracking.NewAPIClient(client.envID, f.APIKey, f.decisionAPIOptions...)
	}

	if client.decisionClient == nil {
		client.decisionMode = f.decisionMode
		if f.decisionMode == Bucketing {
			client.decisionClient, err = bucketing.NewEngine(client.envID, client.cacheManager, f.bucketingOptions...)
			if err != nil {
				clientLogger.Error("Got error when creating bucketing engine", err)
			}
		} else {
			client.decisionClient, err = decision.NewAPIClient(client.envID, f.APIKey, f.decisionAPIOptions...)
			if err != nil {
				clientLogger.Error("Got error when creating Decision API engine", err)
			}
		}
	}

	client.status = STATUS_READY
	return client, err
}

// GetStatus returns the current client status
func (c *Client) GetStatus() string {
	return c.status
}

// NewVisitor returns a new Visitor from ID and context
func (c *Client) NewVisitor(visitorID string, context model.Context, options ...VisitorOptionBuilder) (visitor *Visitor, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = utils.HandleRecovered(r, clientLogger)
		}
	}()

	clientLogger.Info(fmt.Sprintf("Creating new visitor with id : %s", visitorID))

	errs := context.Validate()
	if len(errs) > 0 {
		errorStrings := []string{}
		for _, e := range errs {
			clientLogger.Error("Context error", e)
			errorStrings = append(errorStrings, e.Error())
		}
		return nil, fmt.Errorf("Invalid context : %s", strings.Join(errorStrings, ", "))
	}

	id := visitorID
	var anonymousID *string
	if id == "" {
		id = generateAnonymousID()
	}

	// Build visitor options
	visitorOptions := &VisitorOptions{}
	visitorOptions.BuildVisitorOptions(options...)

	// Set anonymous ID is visitor is created already authenticated
	if visitorOptions.IsAuthenticated {
		newAnonID := generateAnonymousID()
		anonymousID = &newAnonID
	}

	return &Visitor{
		ID:                id,
		AnonymousID:       anonymousID,
		Context:           context,
		decisionClient:    c.decisionClient,
		decisionMode:      c.decisionMode,
		trackingAPIClient: c.trackingAPIClient,
		cacheManager:      c.cacheManager,
	}, nil
}

// SendHit sends a tracking hit to the Data Collect API
func (c *Client) SendHit(visitorID string, anonymousID *string, hit model.HitInterface) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = utils.HandleRecovered(r, clientLogger)
		}
	}()

	clientLogger.Info(fmt.Sprintf("Sending hit for visitor with id : %s", visitorID))
	err = c.trackingAPIClient.SendHit(visitorID, anonymousID, hit)

	if err != nil {
		err = fmt.Errorf("Error when sending hit: %s", err.Error())
	}
	return err
}

// Dispose disposes the Client and close all connections
func (c *Client) Dispose() (err error) {
	return err
}

// GetEnvID returns the current set env id
func (c *Client) GetEnvID() string {
	return c.envID
}

// GetDecisionMode returns the current decision mode
func (c *Client) GetDecisionMode() DecisionMode {
	return c.decisionMode
}

// GetCacheManager returns the current cache manager
func (c *Client) GetCacheManager() cache.Manager {
	return c.cacheManager
}
