package main

import (
	"math/rand"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	infoOnly      struct{}
	infoWithDebug struct{}
	aboveWarn     struct{}
)

func (l infoOnly) Enabled(lv zapcore.Level) bool {
	return lv == zapcore.InfoLevel
}
func (l infoWithDebug) Enabled(lv zapcore.Level) bool {
	return lv == zapcore.InfoLevel || lv == zapcore.DebugLevel
}
func (l aboveWarn) Enabled(lv zapcore.Level) bool {
	return lv >= zapcore.WarnLevel
}

func makeErrorFilter() zapcore.LevelEnabler {
	return aboveWarn{}
}

func init() {
	rand.Seed(time.Now().UnixNano())

	var encoderCfg zapcore.EncoderConfig
	encoderCfg = zap.NewDevelopmentEncoderConfig()
	coreInfo := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.NewMultiWriteSyncer(os.Stdout),
		infoWithDebug{},
	)
	coreError := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.NewMultiWriteSyncer(os.Stderr),
		makeErrorFilter(),
	)

	logger = zap.New(zapcore.NewTee(coreInfo, coreError))
}
