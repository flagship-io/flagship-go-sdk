# Flagship GO SDK Ecommerce example

This example code shows off how to build a simple ecommerce website using Golang,
and customizing its experience with Flagship

## Dependencies

The example use the following libraries and tooling to run the website:
- [Gin](https://github.com/gin-gonic/gin): to serve the assets and the main page
- [html/template](https://pkg.go.dev/html/template): to inject Flagship variables into the HTML
- Flagship Go SDK: to connect to Flagship and get the flags

## Install

1. Run `go mod download` to fetch all the dependencies
2. Get your Flagship Environment ID and API Key from your environment settings in the Flagship Platform
3. (Optional) Setup your own visitor ID and visitor context logic in the main.go file (see TODO comments)
3. Run `FLAGSHIP_ENV_ID={your_env_id} FLAGSHIP_API_KEY={your_api_key} go run *.go`
4. Go to http://localhost:8080 to access the website

## How it works

This websites hosts a single page application on the `/` path.
When loading the home page, the server will fetch flags from Flagship in order to customize the Banner element in the home page.
The banner part of the page can be found here: `./public/banner.html`

Here are the flags used to customize the banner call to action:
- `btn-color`: used to change the background color of the banner click to action button
- `txt-color`: used to change the text color of the banner click to action button
- `btn-text`: used to change the text of the banner click to action button

Additionaly, a debugging banner is displayed on top to show all the flags fetched by Flagship.
