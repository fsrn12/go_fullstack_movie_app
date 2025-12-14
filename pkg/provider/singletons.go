package provider

import (
	"fmt"
	"log/slog"
	"sync"

	"multipass/pkg/logging"
	"multipass/pkg/response"
)

var (
	appLoggerInstance  logging.Logger
	jsonWriterInstance response.Writer
	once               sync.Once
	initErr            error // To store any initialization errors
)

func GetLogger() (logging.Logger, error) {
	var err error
	once.Do(func() {
		// Ensure this block is executed
		appLoggerInstance, err = logging.NewAppLogger("app.log", slog.LevelInfo)
		if err != nil {
			appLoggerInstance.Error("Failed to initialize logger", err)
		}
	})
	if appLoggerInstance == nil {
		appLoggerInstance.Error("Logger is still nil after initialization", nil)
		return nil, fmt.Errorf("logger is not initialized")
	}
	return appLoggerInstance, err
}

func GetWriter() (response.Writer, error) {
	var err error
	once.Do(func() {
		appLoggerInstance.Info("Initializing JSONWriter...")
		logger, err := GetLogger() // Reuse the logger
		if err != nil {
			appLoggerInstance.Error("Failed to get logger for JSONWriter", err)
			return
		}
		jsonWriterInstance = response.NewJSONWriter(logger)
	})

	// Verify initialization after once.Do()
	if jsonWriterInstance == nil {
		appLoggerInstance.Error("JSONWriter is still nil after initialization", nil)
		return nil, fmt.Errorf("JSONWriter is not initialized")
	}

	return jsonWriterInstance, err
}
