package logging

import (
	"os"
	"path/filepath"
	"time"

	"naotodoserver/conf"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

// Logger 全局日志实例
var Logger *logrus.Logger

// InitLogger 初始化日志系统
func InitLogger() {
	Logger = logrus.New()

	// 设置日志级别
	level, err := logrus.ParseLevel(conf.Conf.Log.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	Logger.SetLevel(level)

	// 设置输出格式
	Logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// 创建日志目录（如果不存在）
	logDir := filepath.Dir(conf.Conf.Log.FilePath)
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		Logger.Fatalf("创建日志目录失败: %v", err)
	}

	// 配置日志文件滚动（只设置一个限制，避免冲突）
	var writer *rotatelogs.RotateLogs
	var logErr error

	switch {
	case conf.Conf.Log.MaxAge > 0:
		// 使用最大保留天数
		writer, logErr = rotatelogs.New(
			conf.Conf.Log.FilePath+".%Y%m%d",
			rotatelogs.WithLinkName(conf.Conf.Log.FilePath),
			rotatelogs.WithMaxAge(time.Duration(conf.Conf.Log.MaxAge)*24*time.Hour),
			rotatelogs.WithRotationTime(24*time.Hour),
		)
	case conf.Conf.Log.MaxBackups > 0:
		// 使用最大保留数量
		writer, logErr = rotatelogs.New(
			conf.Conf.Log.FilePath+".%Y%m%d",
			rotatelogs.WithLinkName(conf.Conf.Log.FilePath),
			rotatelogs.WithRotationCount(conf.Conf.Log.MaxBackups),
			rotatelogs.WithRotationTime(24*time.Hour),
		)
	default:
		// 默认配置 - 只按时间滚动
		writer, logErr = rotatelogs.New(
			conf.Conf.Log.FilePath+".%Y%m%d",
			rotatelogs.WithLinkName(conf.Conf.Log.FilePath),
			rotatelogs.WithRotationTime(24*time.Hour),
		)
	}
	if logErr != nil {
		Logger.Fatalf("配置日志文件滚动失败: %v", logErr)
	}

	// 配置输出钩子
	lfsHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.DebugLevel: writer,
		logrus.InfoLevel:  writer,
		logrus.WarnLevel:  writer,
		logrus.ErrorLevel: writer,
		logrus.FatalLevel: writer,
		logrus.PanicLevel: writer,
	}, &logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	Logger.AddHook(lfsHook)

	// 同时输出到控制台（如果配置了）
	if conf.Conf.Log.OutputConsole {
		Logger.SetOutput(os.Stdout)
	}

	Logger.Info("日志系统初始化完成")
}

// Debug 输出调试日志
func Debug(args ...any) {
	Logger.Debug(args...)
}

// Debugf 输出格式化调试日志
func Debugf(format string, args ...any) {
	Logger.Debugf(format, args...)
}

// Info 输出信息日志
func Info(args ...any) {
	Logger.Info(args...)
}

// Infof 输出格式化信息日志
func Infof(format string, args ...any) {
	Logger.Infof(format, args...)
}

// Warn 输出警告日志
func Warn(args ...any) {
	Logger.Warn(args...)
}

// Warnf 输出格式化警告日志
func Warnf(format string, args ...any) {
	Logger.Warnf(format, args...)
}

// Error 输出错误日志
func Error(args ...any) {
	Logger.Error(args...)
}

// Errorf 输出格式化错误日志
func Errorf(format string, args ...any) {
	Logger.Errorf(format, args...)
}

// Fatal 输出致命错误日志
func Fatal(args ...any) {
	Logger.Fatal(args...)
}

// Fatalf 输出格式化致命错误日志
func Fatalf(format string, args ...any) {
	Logger.Fatalf(format, args...)
}

// Panic 输出 panic 日志
func Panic(args ...any) {
	Logger.Panic(args...)
}

// Panicf 输出格式化 panic 日志
func Panicf(format string, args ...any) {
	Logger.Panicf(format, args...)
}
