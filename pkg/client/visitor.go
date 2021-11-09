package client

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/flagship-io/flagship-go-sdk/v2/pkg/cache"
	"github.com/flagship-io/flagship-go-sdk/v2/pkg/decision"
	"github.com/flagship-io/flagship-go-sdk/v2/pkg/logging"
	"github.com/flagship-io/flagship-go-sdk/v2/pkg/model"
	"github.com/flagship-io/flagship-go-sdk/v2/pkg/tracking"
	"github.com/flagship-io/flagship-go-sdk/v2/pkg/utils"
)

var visitorLogger = logging.CreateLogger("FS Visitor")

// Visitor represents a visitor instance of the Flagship SDK
type Visitor struct {
	ID                string
	AnonymousID       *string
	Context           model.Context
	decisionClient    decision.ClientInterface
	decisionMode      DecisionMode
	decisionResponse  *model.APIClientResponse
	flagInfos         map[string]model.FlagInfos
	trackingAPIClient tracking.APIClientInterface
	cacheManager      cache.Manager
}

// ModificationInfo represents additional info linked to the modification key, for third party services
type ModificationInfo struct {
	CampaignID       string
	VariationGroupID string
	VariationID      string
	IsReference      bool
	Value            interface{}
}

func generateAnonymousID() string {
	newID := time.Now().Format("20060102030405.000000")
	return newID[:len(newID)-1]
}

// UpdateContext updates the Visitor context with new value
func (v *Visitor) UpdateContext(newContext model.Context) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = utils.HandleRecovered(r, visitorLogger)
		}
	}()

	errs := newContext.Validate()
	if len(errs) > 0 {
		errorStrings := []string{}
		for _, e := range errs {
			visitorLogger.Error("Context error", e)
			errorStrings = append(errorStrings, e.Error())
		}
		return fmt.Errorf("Invalid context : %s", strings.Join(errorStrings, ", "))
	}

	v.Context = newContext
	return nil
}

// UpdateContextKey updates a single Visitor context key with new value
func (v *Visitor) UpdateContextKey(key string, value interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = utils.HandleRecovered(r, visitorLogger)
		}
	}()

	newContext := model.Context{}
	for k, v := range v.Context {
		newContext[k] = v
	}

	newContext[key] = value

	errs := newContext.Validate()
	if len(errs) > 0 {
		errorStrings := []string{}
		for _, e := range errs {
			visitorLogger.Error("Context error", e)
			errorStrings = append(errorStrings, e.Error())
		}
		return fmt.Errorf("Invalid context : %s", strings.Join(errorStrings, ", "))
	}

	v.Context = newContext
	return nil
}

// Authenticate set the authenticated ID for the visitor, along with optional new context and re-synchronize flag
func (v *Visitor) Authenticate(newID string, newContext map[string]interface{}, sync bool) (err error) {
	if v.decisionMode != API {
		err = errors.New("authenticate() is ignored in BUCKETING mode")
		return err
	}
	if v.AnonymousID == nil {
		anonID := v.ID
		v.AnonymousID = &anonID
	}
	v.ID = newID
	if newContext != nil {
		err = v.UpdateContext(newContext)
		if err != nil {
			return err
		}
	}
	if sync {
		err = v.SynchronizeModifications()
	}
	return err
}

// Unauthenticate unset the authenticated ID for the visitor
func (v *Visitor) Unauthenticate(newContext map[string]interface{}, sync bool) (err error) {
	if v.decisionMode != API {
		err = errors.New("unauthenticate() is ignored in BUCKETING mode")
		return err
	}
	if v.AnonymousID != nil {
		v.ID = *v.AnonymousID
		v.AnonymousID = nil
	}

	if newContext != nil {
		err = v.UpdateContext(newContext)
		if err != nil {
			return err
		}
	}
	if sync {
		err = v.SynchronizeModifications()
	}
	return err
}

// SynchronizeModifications updates the latest campaigns and modifications for the visitor
func (v *Visitor) SynchronizeModifications() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = utils.HandleRecovered(r, visitorLogger)
		}
	}()

	if v.ID == "" {
		err := errors.New("Visitor ID should not be empty")
		visitorLogger.Error("Visitor ID is not set", err)
		return err
	}

	visitorLogger.Info(fmt.Sprintf("Getting modifications for visitor with id : %s", v.ID))
	resp, err := v.decisionClient.GetModifications(v.ID, v.AnonymousID, v.Context)

	if err != nil {
		visitorLogger.Error("Error when calling Decision engine: ", err)
		return err
	}

	if v.trackingAPIClient != nil && v.decisionMode != API {
		go func() {
			visitorLogger.Info("Sending context info to event collect in the background")
			err := v.trackingAPIClient.SendEvent(model.Event{
				VisitorID: v.ID,
				Type:      model.CONTEXT,
				Data:      v.Context,
			})
			if err != nil {
				visitorLogger.Warn("Error when sending context: ", err)
			} else {
				visitorLogger.Info("Context sent successfully")
			}
		}()
	}

	v.decisionResponse = resp

	v.flagInfos = map[string]model.FlagInfos{}

	visitorLogger.Info(fmt.Sprintf("Got %d campaign(s) for visitor with id : %s", len(resp.Campaigns), v.ID))
	for _, c := range resp.Campaigns {
		for k, val := range c.Variation.Modifications.Value {
			v.flagInfos[k] = model.FlagInfos{
				Value:    val,
				Campaign: c,
			}
		}
	}

	return nil
}

// getModification gets a flag value as interface{}
func (v *Visitor) getModification(key string, activate bool) (flagValue interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = utils.HandleRecovered(r, visitorLogger)
		}
	}()

	if v.flagInfos == nil {
		err := errors.New("Visitor modifications have not been synchronized")
		visitorLogger.Error("Visitor modifications are not set", err)

		return false, err
	}

	flagInfos, ok := v.flagInfos[key]

	if !ok {
		return nil, fmt.Errorf("key %s not set in decision infos", key)
	}

	if activate {
		err := v.activateModification(key)
		if err != nil {
			visitorLogger.Debug(fmt.Sprintf("Error occurred when activating campaign : %v.", err))
		}
	}
	flagValue = flagInfos.Value
	return flagValue, nil
}

// GetAllModifications return all the modifications
func (v *Visitor) GetAllModifications() (flagInfos map[string]model.FlagInfos) {
	return v.flagInfos
}

// GetDecisionResponse return the decision response
func (v *Visitor) GetDecisionResponse() *model.APIClientResponse {
	return v.decisionResponse
}

// GetModificationBool get a modification bool by its key
func (v *Visitor) GetModificationBool(key string, defaultValue bool, activate bool) (castVal bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = utils.HandleRecovered(r, visitorLogger)
		}
	}()

	val, err := v.getModification(key, activate)

	if err != nil {
		visitorLogger.Debug(fmt.Sprintf("Error occurred when getting flag value : %v. Fallback to default value", err))
		return defaultValue, err
	}

	if val == nil {
		visitorLogger.Info("Flag value is null in Flagship. Fallback to default value")
		return defaultValue, nil
	}

	castVal, ok := val.(bool)

	if !ok {
		visitorLogger.Debug(fmt.Sprintf("Key %s value %v is not of type bool. Fallback to default value", key, val))
		return defaultValue, fmt.Errorf("Key value cast error : expected bool, got %v", val)
	}

	return castVal, nil
}

// GetModificationString get a modification string by its key
func (v *Visitor) GetModificationString(key string, defaultValue string, activate bool) (castVal string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = utils.HandleRecovered(r, visitorLogger)
		}
	}()

	val, err := v.getModification(key, activate)

	if err != nil {
		visitorLogger.Debug(fmt.Sprintf("Error occurred when getting flag value : %v. Fallback to default value", err))
		return defaultValue, err
	}

	if val == nil {
		visitorLogger.Info("Flag value is null in Flagship. Fallback to default value")
		return defaultValue, nil
	}

	castVal, ok := val.(string)

	if !ok {
		visitorLogger.Debug(fmt.Sprintf("Key %s value %v is not of type string. Fallback to default value", key, val))
		return defaultValue, fmt.Errorf("Key value cast error : expected string, got %v", val)
	}

	return castVal, nil
}

// GetModificationNumber get a modification number as float64 by its key
func (v *Visitor) GetModificationNumber(key string, defaultValue float64, activate bool) (castVal float64, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = utils.HandleRecovered(r, visitorLogger)
		}
	}()

	val, err := v.getModification(key, activate)

	if err != nil {
		visitorLogger.Debug(fmt.Sprintf("Error occurred when getting flag value : %v. Fallback to default value", err))
		return defaultValue, err
	}

	if val == nil {
		visitorLogger.Info("Flag value is null in Flagship. Fallback to default value")
		return defaultValue, nil
	}

	castVal, ok := val.(float64)

	if !ok {
		visitorLogger.Debug(fmt.Sprintf("Key %s value %v is not of type float. Fallback to default value", key, val))
		return defaultValue, fmt.Errorf("Key value cast error : expected float64, got %v", val)
	}

	return castVal, nil
}

// GetModificationObject get a modification object as map[string]interface{} by its key
func (v *Visitor) GetModificationObject(key string, defaultValue map[string]interface{}, activate bool) (castVal map[string]interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = utils.HandleRecovered(r, visitorLogger)
		}
	}()

	val, err := v.getModification(key, activate)

	if err != nil {
		visitorLogger.Debug(fmt.Sprintf("Error occurred when getting flag value : %v. Fallback to default value", err))
		return defaultValue, err
	}

	if val == nil {
		visitorLogger.Info("Flag value is null in Flagship. Fallback to default value")
		return defaultValue, nil
	}

	castVal, ok := val.(map[string]interface{})

	if !ok {
		visitorLogger.Debug(fmt.Sprintf("Key %s value %v is not of type map[string]interface{}. Fallback to default value", key, val))
		return defaultValue, fmt.Errorf("Key value cast error : expected map[string]interface{}, got %v", val)
	}

	return castVal, nil
}

// GetModificationArray get a modification array as []interface{} by its key
func (v *Visitor) GetModificationArray(key string, defaultValue []interface{}, activate bool) (castVal []interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = utils.HandleRecovered(r, visitorLogger)
		}
	}()

	val, err := v.getModification(key, activate)

	if err != nil {
		visitorLogger.Debug(fmt.Sprintf("Error occurred when getting flag value : %v. Fallback to default value", err))
		return defaultValue, err
	}

	if val == nil {
		visitorLogger.Info("Flag value is null in Flagship. Fallback to default value")
		return defaultValue, nil
	}

	castVal, ok := val.([]interface{})

	if !ok {
		visitorLogger.Debug(fmt.Sprintf("Key %s value %v is not of type []interface{}. Fallback to default value", key, val))
		return defaultValue, fmt.Errorf("Key value cast error : expected []interface{}, got %v", val)
	}

	return castVal, nil
}

// GetModificationInfo returns a modification info by its key
func (v *Visitor) GetModificationInfo(key string) (modifInfo *ModificationInfo, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = utils.HandleRecovered(r, visitorLogger)
		}
	}()

	if v.flagInfos == nil {
		err := errors.New("Visitor modifications have not been synchronized")
		visitorLogger.Error("Visitor modifications are not set", err)

		return nil, err
	}

	flagInfos, ok := v.flagInfos[key]

	if !ok {
		err = fmt.Errorf("key %v is not in any campaign", key)
		visitorLogger.Debug(err.Error())
		return nil, err
	}

	return &ModificationInfo{
		CampaignID:       flagInfos.Campaign.ID,
		VariationGroupID: flagInfos.Campaign.VariationGroupID,
		VariationID:      flagInfos.Campaign.Variation.ID,
		IsReference:      flagInfos.Campaign.Variation.Reference,
		Value:            flagInfos.Value,
	}, nil
}

func (v *Visitor) activateModification(key string) error {
	if v.flagInfos == nil {
		err := errors.New("Visitor modifications have not been synchronized")
		return err
	}

	flagInfos, ok := v.flagInfos[key]
	if !ok {
		return fmt.Errorf("key %s not set in decision infos", key)
	}

	visitorLogger.Info(fmt.Sprintf("Activating campaign for flag %s for visitor with id : %s", key, v.ID))
	err := v.trackingAPIClient.ActivateCampaign(model.ActivationHit{
		VariationGroupID: flagInfos.Campaign.VariationGroupID,
		VariationID:      flagInfos.Campaign.Variation.ID,
		VisitorID:        v.ID,
		AnonymousID:      v.AnonymousID,
	})
	if err != nil && v.cacheManager != nil {
		campaignsCache, err := v.cacheManager.Get(v.ID)
		if err == nil {
			existingCampaign, ok := campaignsCache[flagInfos.Campaign.ID]
			if ok && !existingCampaign.Activated {
				existingCampaign.Activated = true
				err = v.cacheManager.Set(v.ID, campaignsCache)
			}
		}
		if err != nil {
			visitorLogger.Warnf("error when activating campaign cache for visitor ID: %v", err)
		}
	}
	return err
}

// ActivateModification notifies Flagship that the visitor has seen to modification
func (v *Visitor) ActivateModification(key string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = utils.HandleRecovered(r, visitorLogger)
		}
	}()

	err = v.activateModification(key)
	return err
}

// ActivateCacheModification activates a modification from the cache of assigned visitor campaigns
func (v *Visitor) ActivateCacheModification(key string) (err error) {
	if v.cacheManager != nil {
		cacheCampaigns, err := v.cacheManager.Get(v.ID)
		if err != nil {
			return err
		}

		for _, c := range cacheCampaigns {
			for _, k := range c.FlagKeys {
				if k == key {
					// Key found in cache. Activating it now
					err = v.trackingAPIClient.ActivateCampaign(model.ActivationHit{
						VariationGroupID: c.VariationGroupID,
						VariationID:      c.VariationID,
						VisitorID:        v.ID,
						AnonymousID:      v.AnonymousID,
					})
					return err
				}
			}
		}

		return fmt.Errorf("Cache for flag key %v not found", key)
	}
	return errors.New("No cache manager defined")
}

// SendHit sends a tracking hit to the Data Collect API
func (v *Visitor) SendHit(hit model.HitInterface) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = utils.HandleRecovered(r, visitorLogger)
		}
	}()

	visitorLogger.Info(fmt.Sprintf("Sending hit for visitor with id : %s", v.ID))
	err = v.trackingAPIClient.SendHit(v.ID, hit)

	if err != nil {
		err = fmt.Errorf("Error when registering hit: %s", err.Error())
	}
	return err
}
