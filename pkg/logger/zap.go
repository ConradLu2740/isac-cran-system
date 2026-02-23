package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

var (
	globalLogger *zap.Logger
	sugarLogger  *zap.SugaredLogger
)

func Init(cfg *Config) error {
	level := getLevel(cfg.Level)

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var encoder zapcore.Encoder
	if strings.ToLower(cfg.Format) == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	var cores []zapcore.Core

	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	consoleCore := zapcore.NewCore(
		consoleEncoder,
		zapcore.AddSync(os.Stdout),
		level,
	)
	cores = append(cores, consoleCore)

	if cfg.Filename != "" {
		fileWriter := &Logger{
			Filename:   cfg.Filename,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}
		fileCore := zapcore.NewCore(
			encoder,
			zapcore.AddSync(fileWriter),
			level,
		)
		cores = append(cores, fileCore)
	}

	core := zapcore.NewTee(cores...)
	globalLogger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	sugarLogger = globalLogger.Sugar()

	return nil
}

func getLevel(levelStr string) zapcore.Level {
	switch strings.ToLower(levelStr) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

func L() *zap.Logger {
	if globalLogger == nil {
		globalLogger = zap.NewExample()
	}
	return globalLogger
}

func S() *zap.SugaredLogger {
	if sugarLogger == nil {
		sugarLogger = L().Sugar()
	}
	return sugarLogger
}

func Sync() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}

func Debug(msg string, fields ...zap.Field) {
	L().Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	L().Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	L().Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	L().Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	L().Fatal(msg, fields...)
}

func With(fields ...zap.Field) *zap.Logger {
	return L().With(fields...)
}
