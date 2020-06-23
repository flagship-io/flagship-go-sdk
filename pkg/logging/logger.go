package logging

import (
	"io"

	"github.com/sirupsen/logrus"
)

var formatter logrus.Formatter
var output io.Writer
var level logrus.Level = logrus.WarnLevel
var loggers map[string]*logrus.Logger = make(map[string]*logrus.Logger)

// SetLevel sets the log level to the given level
func SetLevel(newLevel logrus.Level) {
	level = newLevel
	for _, l := range loggers {
		l.Level = newLevel
	}
}

// LogNameHook is a logrus hook to a
type LogNameHook struct {
	name string
}

// NewNameHook creates a new hook to add name to entry fields
func NewNameHook(name string) *LogNameHook {
	return &LogNameHook{
		name: name,
	}
}

// Levels specifies the levels of the hook
func (h *LogNameHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire is triggered on log entry
func (h *LogNameHook) Fire(entry *logrus.Entry) error {
	entry.Data["name"] = h.name
	return nil
}

// CreateLogger creates a new logger with a name prefix
func CreateLogger(name string) *logrus.Logger {
	// Create a new instance of the logger. You can have any number of instances.
	var logger = logrus.New()
	logger.Level = level
	logger.AddHook(&LogNameHook{
		name: name,
	})

	if output != nil {
		logger.SetOutput(output)
	}

	if formatter != nil {
		logger.SetFormatter(formatter)
	} else {
		// Set default format to text with timestamp
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	loggers[name] = logger

	return logger
}
