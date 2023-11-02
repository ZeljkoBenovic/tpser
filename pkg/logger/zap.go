package logger

import "go.uber.org/zap"

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

func (l *zapLogger) Fatalln(msg string, args ...interface{}) {
	l.log.Fatalw(msg, args...)
}

func (l *zapLogger) Named(name string) Logger {
	l.log = l.log.Named(name)
	return l
}

func NewZapLogger() Logger {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	return &zapLogger{log: logger.Sugar()}
}
