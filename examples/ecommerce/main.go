package main

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"

	"github.com/flagship-io/flagship-go-sdk/v2"
	"github.com/flagship-io/flagship-go-sdk/v2/pkg/client"
	"github.com/flagship-io/flagship-go-sdk/v2/pkg/model"
	"github.com/gin-gonic/gin"
)

var fsClients = make(map[string]*client.Client)
var fsVisitors = make(map[string]*client.Visitor)

// FSEnvInfo Binding env from JSON
type FSEnvInfo struct {
	EnvironmentID string                 `json:"environment_id" binding:"required"`
	VisitorID     string                 `json:"visitor_id" binding:"required"`
	Context       map[string]interface{} `json:"context" binding:"required"`
}

func main() {
	router := gin.Default()

	router.Static("/static", "ecommerce/public")
	router.LoadHTMLGlob("ecommerce/public/*.html")

	router.GET("/", func(c *gin.Context) {
		fsCookie, err := c.Cookie("fscookie")

		var fsInfo FSEnvInfo = FSEnvInfo{}
		if fsCookie != "" {
			data, err := base64.StdEncoding.DecodeString(fsCookie)
			if err == nil {
				err = json.Unmarshal(data, &fsInfo)
				log.Println(err)
				log.Println(fsInfo)
			}
			if err != nil {
				log.Println(err)
			}
		}

		valueBtnColor := "rgb(249, 167, 67)"
		valueTxtColor := "#fff"
		valueBtnText := "Buy for 15% off"

		var variables = make(map[string]model.FlagInfos)
		if fsInfo.EnvironmentID != "" && fsInfo.VisitorID != "" {
			fsClient, _ := fsClients[fsInfo.EnvironmentID]
			if fsClient == nil {
				fsClient, err = flagship.Start(fsInfo.EnvironmentID, client.WithBucketing())
			}
			fsClients[fsInfo.EnvironmentID] = fsClient
			fsVisitor, _ := fsVisitors[fsInfo.EnvironmentID+"-"+fsInfo.VisitorID]
			if fsVisitor == nil {
				fsVisitor, err = fsClient.NewVisitor(fsInfo.VisitorID, fsInfo.Context)
			}
			fsClients[fsInfo.EnvironmentID] = fsClient

			if fsClient != nil && fsVisitor != nil {
				fsVisitor.SynchronizeModifications()
				valueBtnColor, _ = fsVisitor.GetModificationString("btn-color", "rgb(249, 167, 67)", true)
				valueTxtColor, _ = fsVisitor.GetModificationString("txt-color", "#fff", true)
				valueBtnText, _ = fsVisitor.GetModificationString("btn-text", "Buy for 15% off", true)
			}
			variables = fsVisitor.GetAllModifications()
		}

		variablesObj := gin.H{}

		for k, v := range variables {
			variablesObj[k] = v.Value
		}

		variablesJSON, _ := json.Marshal(variablesObj)

		c.HTML(http.StatusOK, "index.html", gin.H{
			"flagship": gin.H{
				"btnColor":  valueBtnColor,
				"txtColor":  valueTxtColor,
				"btnText":   valueBtnText,
				"error":     err,
				"variables": string(variablesJSON),
			},
		})
	})

	router.Run(":8080")
}
