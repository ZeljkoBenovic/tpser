package logger

import "go.uber.org/zap"

type Logger interface {
	Error(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
	Fatalln(msg string, args ...interface{})
	Named(name string) *zap.SugaredLogger
}

type log struct {
	log *zap.SugaredLogger
}

func (l *log) Error(msg string, args ...interface{}) {
	l.log.Errorw(msg, args...)
}

func (l *log) Warn(msg string, args ...interface{}) {
	l.log.Warnw(msg, args...)
}

func (l *log) Info(msg string, args ...interface{}) {
	l.log.Infow(msg, args...)
}

func (l *log) Debug(msg string, args ...interface{}) {
	l.log.Debugw(msg, args...)
}

func (l *log) Fatalln(msg string, args ...interface{}) {
	l.log.Fatalw(msg, args...)
}

func (l *log) Named(name string) *zap.SugaredLogger {
	return l.log.Named(name)
}

func New() Logger {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	return &log{
		log: logger.Sugar(),
	}
}
