module github.com/flagship-io/flagship-go-sdk/examples

go 1.12

replace github.com/flagship-io/flagship-go-sdk/v2 => ../

require (
	github.com/flagship-io/flagship-go-sdk/v2 v2.0.5
	github.com/gin-contrib/sessions v0.0.3
	github.com/gin-gonic/gin v1.6.2
	github.com/segmentio/backo-go v0.0.0-20200129164019-23eae7c10bd3 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/xtgo/uuid v0.0.0-20140804021211-a0b114877d4c // indirect
	google.golang.org/protobuf v1.27.1
	gopkg.in/segmentio/analytics-go.v3 v3.1.0
)
