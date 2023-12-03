package logger

import (
	"github.com/ZeljkoBenovic/tpser/pkg/logger"
	"go.uber.org/zap"
)

type zapLogger struct {
	log *zap.SugaredLogger
}

func (l *zapLogger) Error(msg string, args ...interface{}) {
	l.log.Errorw(msg, args...)
}

func (l *zapLogger) Warn(msg string, args ...interface{}) {
	l.log.Warnw(msg, args...)
}

func (l *zapLogger) Info(msg string, args ...interface{}) {
	l.log.Infow(msg, args...)
}

func (l *zapLogger) Debug(msg string, args ...interface{}) {
	l.log.Debugw(msg, args...)
}

func NewZapLogger() logger.Logger {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	return &zapLogger{log: logger.Sugar()}
}
