package zaplogger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Level                    zapcore.Level
	Development              bool
	DisableRedirectStdLog    bool
	DisableReplaceGlobals    bool
	DisableSampling          bool
	Encoding                 string
	SamplingConfigInitial    int
	SamplingConfigThereafter int
	Options                  []zap.Option
}

func New(cfg Config) (*zap.Logger, func(), error) {
	prodCfg := zap.NewProductionConfig()
	prodCfg.Level = zap.NewAtomicLevelAt(cfg.Level)
	prodCfg.Development = cfg.Development
	if cfg.DisableSampling {
		prodCfg.Sampling = nil
	} else {
		if cfg.SamplingConfigInitial != 0 {
			// Default Initial Sampling is 100
			prodCfg.Sampling.Initial = cfg.SamplingConfigInitial
		}
		if cfg.SamplingConfigThereafter != 0 {
			// Default Thereafter Sampling is 100
			prodCfg.Sampling.Thereafter = cfg.SamplingConfigThereafter
		}
	}

	// Default encoding is JSON
	if cfg.Encoding != "" {
		prodCfg.Encoding = cfg.Encoding
	}

	logger, err := prodCfg.Build(cfg.Options...)
	if err != nil {
		return nil, nil, err
	}
	restoreStdLogFunc := func() {}
	if !cfg.DisableRedirectStdLog {
		restoreStdLogFunc = zap.RedirectStdLog(logger)
	}
	restoreGlobalsFunc := func() {}
	if !cfg.DisableReplaceGlobals {
		restoreGlobalsFunc = zap.ReplaceGlobals(logger)
	}

	return logger, func() {
		restoreGlobalsFunc()
		restoreStdLogFunc()
	}, nil
}
