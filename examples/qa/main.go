package main

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/flagship-io/flagship-go-sdk/v2/pkg/decisionapi"
	analytics "gopkg.in/segmentio/analytics-go.v3"

	"github.com/flagship-io/flagship-go-sdk/v2/pkg/bucketing"

	"github.com/sirupsen/logrus"

	"github.com/flagship-io/flagship-go-sdk/v2"
	"github.com/flagship-io/flagship-go-sdk/v2/pkg/client"
	"github.com/flagship-io/flagship-go-sdk/v2/pkg/logging"
	"github.com/flagship-io/flagship-go-sdk/v2/pkg/model"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

var fsClients = make(map[string]*client.Client)
var fsVisitors = make(map[string]*client.Visitor)
var segmentClient analytics.Client

// FsSession express infos saved in session
type FsSession struct {
	EnvID           string //"blvo2kijq6pg023l8edg"
	APIKey          string
	UseBucketing    bool //true
	VisitorID       string
	Timeout         int
	PollingInterval int
	SegmentAPIKey   string
}

// FSEnvInfo Binding env from JSON
type FSEnvInfo struct {
	EnvironmentID   string `json:"environment_id" binding:"required"`
	APIKey          string `json:"api_key" binding:"required"`
	Bucketing       bool   `json:"bucketing"`
	Timeout         int    `json:"timeout"`
	PollingInterval int    `json:"polling_interval"`
	SegmentAPIKey   string `json:"segment_api_key"`
}

// FSVisitorInfo Binding visitor from JSON
type FSVisitorInfo struct {
	VisitorID       string                 `json:"visitor_id" binding:"required"`
	IsAuthenticated bool                   `json:"is_authenticated"`
	Context         map[string]interface{} `json:"context"`
}

// FSVisitorAuthInfo Binding visitor auth from JSON
type FSVisitorAuthInfo struct {
	NewVisitorID string `json:"new_visitor_id" binding:"required"`
}

// FSVisitorUnauthInfo Binding visitor unauth from JSON
type FSVisitorUnauthInfo struct {
}

// FSVisitorInfo Binding visitor from JSON
type FSUpdateContextInfo struct {
	Type  string `json:"type" binding:"required"`
	Value string `json:"value" binding:"required"`
}

// FSHitInfo Binding visitor from JSON
type FSHitInfo struct {
	HitType                string  `json:"t" binding:"required"`
	Action                 string  `json:"ea"`
	Category               string  `json:"ec"`
	Value                  int64   `json:"ev"`
	TransactionID          string  `json:"tid"`
	TransactionAffiliation string  `json:"ta"`
	TransactionRevenue     float64 `json:"tr"`
	ItemCode               string  `json:"ic"`
	ItemName               string  `json:"in"`
	ItemQuantity           int     `json:"iq"`
	DocumentLocation       string  `json:"dl"`
}

func printMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func initSession() gin.HandlerFunc {

	return func(c *gin.Context) {
		printMemUsage()
	}
}

func getFsSession(c *gin.Context) *FsSession {
	session := sessions.Default(c)
	fsSessInt := session.Get("fs_session")
	if fsSessInt == nil {
		return nil
	}
	fsSess := fsSessInt.(*FsSession)
	return fsSess
}

func setFsSession(c *gin.Context, fsS *FsSession) {
	session := sessions.Default(c)
	session.Set("fs_session", fsS)
	err := session.Save()

	if err != nil {
		log.Fatalf("Error on saved cookie : %v", err)
	}
}

func returnVisitor(c *gin.Context, fsVisitor *client.Visitor, err error) {
	flagInfos := fsVisitor.GetAllModifications()

	resp := gin.H{
		"flags":       flagInfos,
		"visitorId":   fsVisitor.ID,
		"anonymousId": fsVisitor.AnonymousID,
	}
	if err != nil {
		resp["error"] = err.Error()
	}

	c.JSON(http.StatusOK, resp)
}

func main() {
	log.Println("Setting log level")
	logging.SetLevel(logrus.DebugLevel)
	router := gin.Default()
	store := cookie.NewStore([]byte("fs-go-sdk-demo-secret"))
	router.Use(sessions.Sessions("fs-go-sdk-demo", store))
	gob.Register(&FsSession{})

	router.Use(initSession())

	router.Static("/static", "qa/assets")

	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "static/")
	})

	router.GET("/env", func(c *gin.Context) {
		fsSession := getFsSession(c)
		if fsSession != nil {
			timeout := 2000
			if fsSession.Timeout > 0 {
				timeout = fsSession.Timeout
			}
			pollingInterval := 60000
			if fsSession.PollingInterval > 0 {
				pollingInterval = fsSession.PollingInterval
			}
			c.JSON(http.StatusOK, FSEnvInfo{
				EnvironmentID:   fsSession.EnvID,
				APIKey:          fsSession.APIKey,
				Bucketing:       fsSession.UseBucketing,
				Timeout:         timeout,
				PollingInterval: pollingInterval,
				SegmentAPIKey:   fsSession.SegmentAPIKey,
			})
		}
	})

	router.PUT("/env", func(c *gin.Context) {
		var json FSEnvInfo
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var fsClient *client.Client
		var err error

		timeout := 2000
		if json.Timeout > 0 {
			timeout = json.Timeout
		}
		pollingInterval := 60000
		if json.PollingInterval > 0 {
			pollingInterval = json.PollingInterval
		}

		if json.Bucketing {
			fsClient, err = flagship.Start(json.EnvironmentID, json.APIKey, client.WithBucketing(
				bucketing.PollingInterval(
					time.Duration(pollingInterval)*time.Millisecond)))
		} else {
			fsClient, err = flagship.Start(json.EnvironmentID, json.APIKey, client.WithDecisionAPI(
				decisionapi.Timeout(
					time.Duration(timeout)*time.Millisecond), decisionapi.APIUrl(os.Getenv("DECISION_API_URL"))))
		}

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		fsSession := getFsSession(c)
		if fsSession != nil {
			fsClientExisting := fsClients[fsSession.EnvID]
			if fsClientExisting != nil {
				fsClientExisting.Dispose()
				fsClientExisting = nil
			}
		}
		fsClients[json.EnvironmentID] = fsClient
		setFsSession(c, &FsSession{
			EnvID:           json.EnvironmentID,
			APIKey:          json.APIKey,
			UseBucketing:    json.Bucketing,
			Timeout:         timeout,
			PollingInterval: pollingInterval,
			SegmentAPIKey:   json.SegmentAPIKey,
		})

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/visitor", func(c *gin.Context) {
		fsSession := getFsSession(c)
		if fsSession == nil || fsSession.VisitorID == "" {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "visitor not initialized",
			})
			return
		}
		visitor, ok := fsVisitors[fsSession.EnvID+"-"+fsSession.VisitorID]
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "visitor not initialized",
			})
			return
		}
		c.JSON(http.StatusOK, FSVisitorInfo{
			VisitorID: visitor.ID,
			Context:   visitor.Context,
		})
	})

	router.PUT("/visitor", func(c *gin.Context) {
		var json FSVisitorInfo
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		fsSession := getFsSession(c)
		fsClient := fsClients[fsSession.EnvID]
		if fsClient == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "FS Client not initialized"})
			return
		}

		fsVisitor, err := fsClient.NewVisitor(json.VisitorID, json.Context, client.WithAuthenticated(json.IsAuthenticated))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		err = fsVisitor.SynchronizeModifications()
		fsVisitors[fsSession.EnvID+"-"+json.VisitorID] = fsVisitor
		setFsSession(c, &FsSession{
			EnvID:         fsSession.EnvID,
			APIKey:        fsSession.APIKey,
			UseBucketing:  fsSession.UseBucketing,
			VisitorID:     json.VisitorID,
			SegmentAPIKey: fsSession.SegmentAPIKey,
		})

		returnVisitor(c, fsVisitor, err)
	})

	router.POST("/authenticate", func(c *gin.Context) {
		var json FSVisitorAuthInfo
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		fsSession := getFsSession(c)
		fsClient, _ := fsClients[fsSession.EnvID]
		if fsClient == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "FS Client not initialized"})
			return
		}

		fsVisitor, ok := fsVisitors[fsSession.EnvID+"-"+fsSession.VisitorID]
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Errorf("Visitor ID %v not found", fsSession.VisitorID),
			})
			return
		}

		err := fsVisitor.Authenticate(json.NewVisitorID, nil, true)

		fsVisitors[fsSession.EnvID+"-"+fsVisitor.ID] = fsVisitor
		setFsSession(c, &FsSession{
			EnvID:        fsSession.EnvID,
			APIKey:       fsSession.APIKey,
			UseBucketing: fsSession.UseBucketing,
			VisitorID:    fsVisitor.ID,
		})

		returnVisitor(c, fsVisitor, err)
	})

	router.POST("/unauthenticate", func(c *gin.Context) {
		var json FSVisitorUnauthInfo
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		fsSession := getFsSession(c)
		fsClient, _ := fsClients[fsSession.EnvID]
		if fsClient == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "FS Client not initialized"})
			return
		}

		fsVisitor, ok := fsVisitors[fsSession.EnvID+"-"+fsSession.VisitorID]
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Errorf("Visitor ID %v not found", fsSession.VisitorID),
			})
			return
		}

		err := fsVisitor.Unauthenticate(nil, true)

		fsVisitors[fsSession.EnvID+"-"+fsVisitor.ID] = fsVisitor
		setFsSession(c, &FsSession{
			EnvID:        fsSession.EnvID,
			APIKey:       fsSession.APIKey,
			UseBucketing: fsSession.UseBucketing,
			VisitorID:    fsVisitor.ID,
		})

		returnVisitor(c, fsVisitor, err)
	})

	router.PUT("/visitor/context/:key", func(c *gin.Context) {
		var key = c.Param("key")
		var json FSUpdateContextInfo
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if key == "" || json.Type == "" || json.Value == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing context key, type or value"})
			return
		}

		fsSession := getFsSession(c)
		fsClient := fsClients[fsSession.EnvID]
		if fsClient == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "FS Client not initialized"})
			return
		}

		fsVisitor := fsVisitors[fsSession.EnvID+"-"+fsSession.VisitorID]
		if fsVisitor == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "FS Visitor not initialized"})
			return
		}

		var err error
		switch json.Type {
		case "bool":
			defVal, castErr := strconv.ParseBool(json.Value)
			if castErr != nil {
				err = castErr
				break
			}
			err = fsVisitor.UpdateContextKey(key, defVal)
		case "number":
			defVal, castErr := strconv.ParseFloat(json.Value, 64)
			if castErr != nil {
				err = castErr
				break
			}
			err = fsVisitor.UpdateContextKey(key, defVal)
		case "string":
			err = fsVisitor.UpdateContextKey(key, json.Value)

		default:
			err = fmt.Errorf("Context key type %v not handled", json.Type)
		}

		if err == nil {
			err = fsVisitor.SynchronizeModifications()
		}

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			flagInfos := fsVisitor.GetAllModifications()
			c.JSON(http.StatusOK, gin.H{
				"visitorId": fsVisitor.ID,
				"context":   fsVisitor.Context,
				"flags":     flagInfos,
			})
		}

	})

	//router.LoadHTMLFiles("templates/template1.html", "templates/template2.html")
	router.GET("/flag/:name", func(c *gin.Context) {
		var flag = c.Param("name")
		var flagType = c.Query("type")
		var activate = c.Query("activate")
		var defaultValue = c.Query("defaultValue")

		if flag == "" || flagType == "" || activate == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing flag name, type, activate or defaultValue"})
			return
		}

		fsSession := getFsSession(c)
		fsClient := fsClients[fsSession.EnvID]
		if fsClient == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "FS Client not initialized"})
			return
		}

		fsVisitor := fsVisitors[fsSession.EnvID+"-"+fsSession.VisitorID]
		if fsVisitor == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "FS Visitor not initialized"})
			return
		}

		var value interface{}
		var err error
		shouldActivate, err := strconv.ParseBool(activate)

		if err == nil {
			switch flagType {
			case "bool":
				defVal, castErr := strconv.ParseBool(defaultValue)
				if castErr != nil {
					err = castErr
					break
				}

				value, err = fsVisitor.GetModificationBool(flag, defVal, shouldActivate)
			case "number":
				defVal, castErr := strconv.ParseFloat(defaultValue, 64)
				if castErr != nil {
					err = castErr
					break
				}

				value, err = fsVisitor.GetModificationNumber(flag, defVal, shouldActivate)
			case "string":
				value, err = fsVisitor.GetModificationString(flag, defaultValue, shouldActivate)
			case "object":
				defVal := map[string]interface{}{}
				if defaultValue != "" {
					castErr := json.Unmarshal([]byte(defaultValue), &defVal)
					if castErr != nil {
						err = castErr
						break
					}
				}
				value, err = fsVisitor.GetModificationObject(flag, defVal, shouldActivate)
			case "array":
				defVal := []interface{}{}
				if defaultValue != "" {
					castErr := json.Unmarshal([]byte(defaultValue), &defVal)
					if castErr != nil {
						err = castErr
						break
					}
				}
				value, err = fsVisitor.GetModificationArray(flag, defVal, shouldActivate)
			default:
				err = fmt.Errorf("Flag type %v not handled", flagType)
			}
		}

		errString := ""
		status := http.StatusOK
		if err != nil {
			status = http.StatusBadRequest
			errString = err.Error()
		}

		c.JSON(status, gin.H{"value": value, "error": errString})
	})

	router.GET("/flag/:name/activate", func(c *gin.Context) {
		var flag = c.Param("name")

		if flag == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing flag key"})
			return
		}

		fsSession := getFsSession(c)
		fsClient := fsClients[fsSession.EnvID]
		if fsClient == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "FS Client not initialized"})
			return
		}

		fsVisitor := fsVisitors[fsSession.EnvID+"-"+fsSession.VisitorID]
		if fsVisitor == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "FS Visitor not initialized"})
			return
		}

		err := fsVisitor.ActivateModification(flag)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	router.GET("/flag/:name/info", func(c *gin.Context) {
		var flag = c.Param("name")

		if flag == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing flag key"})
			return
		}

		fsSession := getFsSession(c)
		fsClient := fsClients[fsSession.EnvID]
		if fsClient == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "FS Client not initialized"})
			return
		}

		fsVisitor := fsVisitors[fsSession.EnvID+"-"+fsSession.VisitorID]
		if fsVisitor == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "FS Visitor not initialized"})
			return
		}

		modifInfos, err := fsVisitor.GetModificationInfo(flag)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Track segment
		if fsSession.SegmentAPIKey != "" {
			segmentClient = analytics.New(fsSession.SegmentAPIKey)
			defer segmentClient.Close()

			data := analytics.Track{
				UserId: fsVisitor.ID,
				Event:  "Flagship_Source_Go",
				Properties: analytics.NewProperties().
					Set("cid", modifInfos.CampaignID).
					Set("vgid", modifInfos.VariationGroupID).
					Set("vid", modifInfos.VariationID).
					Set("isref", modifInfos.IsReference).
					Set("val", modifInfos.Value),
			}
			fmt.Println("Track to segment", data)
			segmentClient.Enqueue(data)
		}

		c.JSON(http.StatusOK, modifInfos)
	})

	//router.LoadHTMLFiles("templates/template1.html", "templates/template2.html")
	router.POST("/hit", func(c *gin.Context) {
		var json FSHitInfo
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		fsSession := getFsSession(c)
		if fsSession == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "FS Session not initialized"})
			return
		}

		fsClient := fsClients[fsSession.EnvID]
		if fsClient == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "FS Client not initialized"})
			return
		}

		fsVisitor := fsVisitors[fsSession.EnvID+"-"+fsSession.VisitorID]
		if fsVisitor == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "FS Visitor not initialized"})
			return
		}

		hitType := json.HitType

		var hit model.HitInterface

		switch hitType {
		case "EVENT":
			hit = &model.EventHit{Action: json.Action, Value: json.Value, Category: json.Category}
		case "PAGE":
			hit = &model.PageHit{BaseHit: model.BaseHit{DocumentLocation: json.DocumentLocation}}
		case "SCREEN":
			hit = &model.ScreenHit{BaseHit: model.BaseHit{DocumentLocation: json.DocumentLocation}}
		case "TRANSACTION":
			rand.Seed(time.Now().UnixNano())
			hit = &model.TransactionHit{TransactionID: json.TransactionID, Affiliation: json.TransactionAffiliation, Revenue: json.TransactionRevenue}
		case "ITEM":
			hit = &model.ItemHit{TransactionID: json.TransactionID, Name: json.ItemName, Quantity: json.ItemQuantity, Code: json.ItemCode}
		}

		err := fsVisitor.SendHit(hit)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok", "hitType": hitType})
	})

	router.Run(":8080")
}
