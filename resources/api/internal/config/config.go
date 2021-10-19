package config

import (
	"api/pkg/log"
	"context"

	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
)

type ApiConfig struct {
	LogLevel string `envconfig:"LOG_LEVEL" default:"INFO"`
}

// our config
var apiConfig ApiConfig

// The key type is unexported to prevent collisions with context keys defined in
// other packages.
type contextKey string

func (c contextKey) String() string {
	return "context key " + string(c)
}

//
var (
	contextKeyConfig = contextKey("config")
)

// ReadEnvConfig is
func ReadEnvConfig(ctx context.Context, namespace string) context.Context {

	logger := log.Logger(ctx)

	err := envconfig.Process(namespace, &apiConfig)

	if err != nil {
		logger.Panic("unable to process environment", zap.Error(err))
	}

	return context.WithValue(ctx, contextKeyConfig, &apiConfig)
}

// GetConfig is
func GetConfig(ctx context.Context) *ApiConfig {

	logger := log.Logger(ctx)

	econfig, ok := ctx.Value(contextKeyConfig).(*ApiConfig)

	if !ok {
		logger.Panic("unable to retrieve config value")
	}

	return econfig
}
