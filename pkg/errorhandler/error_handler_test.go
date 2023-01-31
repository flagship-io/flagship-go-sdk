package errorhandler

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flagship-io/flagship-go-sdk/v3/pkg/logging"
)

var testLogger = logging.CreateLogger("FS Test")

func TestHandleRecovered(t *testing.T) {
	func() {
		defer func() {
			if r := recover(); r != nil {
				err := HandleRecovered(r, testLogger)

				assert.NotNil(t, err)
				assert.Equal(t, err.Error(), "Flagship SDK recovered from the error : test")
			}
		}()
		panic("test")
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				err := HandleRecovered(r, testLogger)

				assert.NotNil(t, err)
				assert.Equal(t, err.Error(), "Flagship SDK recovered from the error : test")
			}
		}()
		panic(errors.New("test"))
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				err := HandleRecovered(r, testLogger)

				assert.NotNil(t, err)
				assert.Equal(t, err.Error(), "Flagship SDK recovered from the error : false")
			}
		}()
		panic(false)
	}()
}
