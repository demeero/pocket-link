package zaplogger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Dev                      bool
	DisableRedirectStdLog    bool
	DisableReplaceGlobals    bool
	Encoding                 string
	DisableSampling          bool
	SamplingConfigInitial    int
	SamplingConfigThereafter int
	Level                    zapcore.Level
	Options                  []zap.Option
}

func New(cfg Config) (*zap.Logger, func(), error) {
	logCfg := zap.NewProductionConfig()
	if cfg.SamplingConfigInitial != 0 {
		logCfg.Sampling.Initial = cfg.SamplingConfigInitial
	}
	if cfg.SamplingConfigThereafter != 0 {
		logCfg.Sampling.Thereafter = cfg.SamplingConfigThereafter
	}
	if cfg.DisableSampling {
		logCfg.Sampling = nil
	}
	if cfg.Encoding != "" {
		logCfg.Encoding = cfg.Encoding
	}

	logger, err := logCfg.Build(cfg.Options...)
	if err != nil {
		return nil, nil, err
	}
	restoreGlobalsFunc := func() {}
	if !cfg.DisableReplaceGlobals {
		restoreGlobalsFunc = zap.ReplaceGlobals(logger)
	}
	restoreStdLogFunc := func() {}
	if !cfg.DisableRedirectStdLog {
		restoreStdLogFunc = zap.RedirectStdLog(logger)
	}
	logCfg.Development = cfg.Dev
	logCfg.Level = zap.NewAtomicLevelAt(cfg.Level)
	return logger, func() {
		restoreGlobalsFunc()
		restoreStdLogFunc()
	}, nil
}
