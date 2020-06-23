package logging

import (
	"bytes"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sirupsen/logrus"
)

func TestSetLevel(t *testing.T) {
	assert.Equal(t, level, logrus.WarnLevel)
	SetLevel(logrus.InfoLevel)
	assert.Equal(t, level, logrus.InfoLevel)
}

func TestCreateLogger(t *testing.T) {
	SetLevel(logrus.DebugLevel)

	var buf bytes.Buffer
	output = &buf

	logger := CreateLogger("test")

	logger.Debugf("debug log : %v", "test")

	assert.Contains(t, buf.String(), "level=debug msg=\"debug log : test\" name=test\n")

	log.SetOutput(os.Stderr)
}

func TestCreateLoggerFormatter(t *testing.T) {
	SetLevel(logrus.DebugLevel)

	var buf bytes.Buffer
	output = &buf

	formatter = &logrus.JSONFormatter{}

	logger := CreateLogger("test")

	logger.Debugf("debug log : %v", "test")

	assert.Contains(t, buf.String(), `{"level":"debug","msg":"debug log : test","name":"test"`)

	log.SetOutput(os.Stderr)
}
