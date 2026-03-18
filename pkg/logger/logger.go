package logger

import (
	"go.uber.org/zap"
)

var Log *zap.Logger

func Init(dev bool) {

	var err error
	var cfg zap.Config

	if dev {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}

	cfg.EncoderConfig.CallerKey = ""
	cfg.DisableCaller = true

	Log, err = cfg.Build()
	if err != nil {
		panic(err)
	}
}

func Sync() {
	if Log != nil {
		Log.Sync()
	}
}
