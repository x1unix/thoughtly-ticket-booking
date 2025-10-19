package config

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogConfig struct {
	Level        zapcore.Level `default:"info"`
	IsProduction bool          `envconfig:"IS_PROD"`
}

func (logCfg LogConfig) BuildZapLogger() (*zap.Logger, error) {
	var cfg zap.Config
	if logCfg.IsProduction {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
	}

	cfg.Level = zap.NewAtomicLevelAt(logCfg.Level)
	l, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build Zap logger: %w", err)
	}

	return l, nil
}
