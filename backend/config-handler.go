package backend

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"gorm.io/gorm/logger"
)

type Config struct {
	LogLevel string `json:"logLevel" doc:"Log level: silent, error, warn, info"`
}

var runtimeConfig = Config{
	LogLevel: "warn",
}

func parseLogLevel(s string) logger.LogLevel {
	switch s {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "info":
		return logger.Info
	default:
		return logger.Warn
	}
}

func applyLogLevel(level string) {
	runtimeConfig.LogLevel = level
	db.Logger = db.Logger.LogMode(parseLogLevel(level))
}

// SetInitialLogLevel is called at startup with the flag value.
func SetInitialLogLevel(level string) {
	applyLogLevel(level)
}

var GetConfigOperation = huma.Operation{
	Method:      http.MethodGet,
	Path:        "/api/v2/config",
	DefaultStatus: http.StatusOK,
}

func GetConfig(_ context.Context, _ *struct{}) (output *struct{ Body Config }, err error) {
	output = &struct{ Body Config }{Body: runtimeConfig}
	return
}

var PutConfigOperation = huma.Operation{
	Method:        http.MethodPut,
	Path:          "/api/v2/config",
	DefaultStatus: http.StatusOK,
}

func PutConfig(_ context.Context, input *struct {
	Body Config
}) (output *struct{ Body Config }, err error) {
	applyLogLevel(input.Body.LogLevel)
	output = &struct{ Body Config }{Body: runtimeConfig}
	return
}
