package main

import (
	"go.uber.org/zap"
)

func setupLogger(logLevel string) *zap.SugaredLogger {
	zapLevel, err := zap.ParseAtomicLevel(logLevel)
	if err != nil {
		zapLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	log := zap.Must(zap.Config{
		Level:            zapLevel,
		Development:      false,
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}.Build()).Sugar()
	zap.ReplaceGlobals(log.Desugar())
	return log
}
