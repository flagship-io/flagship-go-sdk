package errorhandler

import (
	"fmt"
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

// HandleRecovered logs a recovered panic
func HandleRecovered(r interface{}, logger *logrus.Logger) (err error) {
	err = fmt.Errorf("Flagship SDK recovered from the error : %v", r)
	logger.Error(err)
	logger.Debug(string(debug.Stack()))

	return err
}
