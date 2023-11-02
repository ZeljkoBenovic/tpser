package logger

import (
	"github.com/sirupsen/logrus"
)

type log struct {
	log *logrus.Logger
}

func (l *log) Error(msg string, args ...interface{}) {
	fields := map[string]any{}

	for i := 0; i < len(args); i += 2 {
		fields[args[i].(string)] = args[i+1]
	}

	l.log.WithFields(fields).Error(msg)
}

func (l *log) Warn(msg string, args ...interface{}) {
	fields := map[string]any{}

	for i := 0; i < len(args); i += 2 {
		fields[args[i].(string)] = args[i+1]
	}

	l.log.WithFields(fields).Warn(msg)
}

func (l *log) Info(msg string, args ...interface{}) {
	fields := map[string]any{}

	for i := 0; i < len(args); i += 2 {
		fields[args[i].(string)] = args[i+1]
	}

	l.log.WithFields(fields).Info(msg)
}

func (l *log) Debug(msg string, args ...interface{}) {
	fields := map[string]any{}

	for i := 0; i < len(args); i += 2 {
		fields[args[i].(string)] = args[i+1]
	}

	l.log.WithFields(fields).Debug(msg)
}

func (l *log) Fatalln(msg string, args ...interface{}) {
	fields := map[string]any{}

	for i := 0; i < len(args); i += 2 {
		fields[args[i].(string)] = args[i+1]
	}

	l.log.WithFields(fields).Fatal(msg)
}

func (l *log) Named(_ string) Logger {
	return l
}

func NewLogrusLogger() Logger {
	return &log{
		log: logrus.StandardLogger(),
	}
}
