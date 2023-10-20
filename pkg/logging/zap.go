package logging

import (
	"context"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ Logger = (*zapLogger)(nil)

type zapLogger struct {
	logger *zap.SugaredLogger
}

func NewZapLogger(level string) *zapLogger {
	var l zapcore.Level
	switch strings.ToLower(level) {
	case "error":
		l = zapcore.ErrorLevel
	case "warn":
		l = zapcore.WarnLevel
	case "info":
		l = zapcore.InfoLevel
	case "debug":
		l = zapcore.DebugLevel
	default:
		l = zapcore.InfoLevel
	}

	// create logger config
	config := zap.Config{
		Development:      false,
		Encoding:         "json",
		Level:            zap.NewAtomicLevelAt(l),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			EncodeDuration: zapcore.SecondsDurationEncoder,
			LevelKey:       "severity",
			EncodeLevel:    zapcore.CapitalLevelEncoder, // e.g. "Info"
			CallerKey:      "caller",
			EncodeCaller:   zapcore.ShortCallerEncoder, // e.g. package/file:line
			TimeKey:        "timestamp",
			EncodeTime:     zapcore.ISO8601TimeEncoder, // e.g. 2020-05-05T03:24:36.903+0300
			NameKey:        "name",
			EncodeName:     zapcore.FullNameEncoder, // e.g. GetSiteGeneralHandler
			MessageKey:     "message",
			StacktraceKey:  "",
			LineEnding:     "\n",
		},
	}

	// build logger from config
	logger, _ := config.Build()

	return &zapLogger{
		logger: logger.Sugar(),
	}
}

func (l *zapLogger) Named(name string) Logger {
	return &zapLogger{
		logger: l.logger.Named(name),
	}
}

func (l *zapLogger) With(args ...interface{}) Logger {
	return &zapLogger{
		logger: l.logger.With(args...),
	}
}

func (l *zapLogger) WithContext(ctx context.Context) Logger {
	return l.With("RequestID", ctx.Value("RequestID"))
}

func (l *zapLogger) Debug(message string, args ...interface{}) {
	l.logger.Debugw(message, args...)
}

func (l *zapLogger) Info(message string, args ...interface{}) {
	l.logger.Infow(message, args...)
}

func (l *zapLogger) Warn(message string, args ...interface{}) {
	l.logger.Warnw(message, args...)
}

func (l *zapLogger) Error(message string, args ...interface{}) {
	l.logger.Errorw(message, args...)
}

func (l *zapLogger) Fatal(message string, args ...interface{}) {
	l.logger.Fatalw(message, args...)
	os.Exit(1)
}
