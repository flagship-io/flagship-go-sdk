package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/flagship-io/flagship-go-sdk/v2"
	"github.com/gin-gonic/gin"
)

var defaultValueBtnColor = "rgb(249, 167, 67)"
var defaultValueTxtColor = "#fff"
var defaultValueBtnText = "Buy for 15% off"

// TODO: Those variables are mandatory for Flagship to run without errors
var flagshipEnvID = os.Getenv("FLAGSHIP_ENV_ID")
var flagshipAPIKey = os.Getenv("FLAGSHIP_API_KEY")

func main() {
	// Start the Flagship SDK with the environment ID and API key
	fsClient, err := flagship.Start(flagshipEnvID, flagshipAPIKey)
	if err != nil {
		log.Fatalf("error when starting Flagship: %v", err)
	}

	router := gin.Default()

	router.Static("/static", "public")

	router.SetFuncMap(template.FuncMap{
		"safe": func(s string) template.HTMLAttr {
			return template.HTMLAttr(s)
		},
	})
	router.LoadHTMLGlob("public/*.html")

	router.GET("/", func(c *gin.Context) {

		// TODO: use your own visitor ID from cookie, session, database, uuid, ...
		flagshipVisitorID := "visitorid_1"

		// TODO: use your own visitor context (data that you want to target your visitor with) from cookie, session, database, ...
		flagshipVisitorContext := map[string]interface{}{
			"device": "firefox",
		}

		// Create a Flagship visitor with an ID and a context
		fsVisitor, err := fsClient.NewVisitor(flagshipVisitorID, flagshipVisitorContext)
		if err != nil {
			log.Fatalf("error when creating Flagship visitor: %v", err)
		}

		// Fetch flags and metadata from Flagship
		err = fsVisitor.SynchronizeModifications()
		if err != nil {
			log.Fatalf("error when creating synchronizing Flagship flags: %v", err)
		}

		// Get flags from Flagship to customize the banner (banner.html)
		valueBtnColor, _ := fsVisitor.GetModificationString("btn-color", defaultValueBtnColor, true)
		valueTxtColor, _ := fsVisitor.GetModificationString("txt-color", defaultValueTxtColor, true)
		valueBtnText, _ := fsVisitor.GetModificationString("btn-text", defaultValueBtnText, true)

		// Get all loaded flags from Flagship for debugging purposes
		variablesObj := gin.H{}
		for k, v := range fsVisitor.GetAllModifications() {
			variablesObj[k] = v
		}

		c.HTML(http.StatusOK, "index.html", gin.H{
			"flagship": gin.H{
				"btnStyle":  fmt.Sprintf("style=\"color:%s;background-color:%s\"", valueTxtColor, valueBtnColor),
				"btnText":   valueBtnText,
				"error":     err,
				"variables": variablesObj,
			},
		})
	})

	router.Run(":8080")
}
