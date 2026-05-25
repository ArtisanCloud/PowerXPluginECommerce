package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"

	runtimelogging "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/runtime/logging"
	"github.com/sirupsen/logrus"
)

// Logger 全局日志实例。
var Logger *logrus.Logger
var httpAccessEnabled = true

var (
	backendMu sync.RWMutex
	backend   *slog.Logger
)

// Fields 日志字段类型别名。
type Fields = logrus.Fields

type Entry = logrus.Entry
type LoggerType = logrus.Logger
type Level = logrus.Level

const (
	PanicLevel Level = logrus.PanicLevel
	FatalLevel Level = logrus.FatalLevel
	ErrorLevel Level = logrus.ErrorLevel
	WarnLevel  Level = logrus.WarnLevel
	InfoLevel  Level = logrus.InfoLevel
	DebugLevel Level = logrus.DebugLevel
	TraceLevel Level = logrus.TraceLevel
)

// Init 初始化统一日志后端与 logrus 兼容层。
func Init(level, format, output, filePath string, maxSize, maxBackups, maxAge int, httpAccess bool) {
	InitWithHostMode(level, format, output, filePath, maxSize, maxBackups, maxAge, httpAccess, runtimelogging.IsHostProxyMode())
}

func InitWithHostMode(level, format, output, filePath string, maxSize, maxBackups, maxAge int, httpAccess bool, hostMode bool) {
	Logger = logrus.New()
	httpAccessEnabled = httpAccess
	runtimelogging.SetHostModeOverride(hostMode)
	if hostMode {
		format = "json"
		output = string(runtimelogging.SinkStdout)
	}
	policy := runtimelogging.ResolveWithHostDefaults(runtimelogging.Policy{
		Format: strings.TrimSpace(format),
		Level:  strings.TrimSpace(level),
		Sinks:  []runtimelogging.SinkType{runtimelogging.SinkType(strings.TrimSpace(output))},
	})
	if err := runtimelogging.ValidatePolicy(policy); err == nil {
		format = policy.Format
		level = policy.Level
		output = runtimelogging.PrimaryOutput(policy)
	}

	logLevel, err := logrus.ParseLevel(strings.ToLower(strings.TrimSpace(level)))
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	Logger.SetLevel(logLevel)

	logWriter, warnMsg := resolveLogWriter(strings.ToLower(strings.TrimSpace(output)), strings.TrimSpace(filePath))
	if warnMsg != "" {
		_, _ = fmt.Fprintln(os.Stderr, warnMsg)
	}

	handlerOpts := &slog.HandlerOptions{Level: mapLogrusLevel(logLevel)}
	logFormat := strings.ToLower(strings.TrimSpace(format))
	if logFormat == "" {
		logFormat = "json"
	}
	var handler slog.Handler
	if logFormat == "text" {
		handler = slog.NewTextHandler(logWriter, handlerOpts)
	} else {
		handler = slog.NewJSONHandler(logWriter, handlerOpts)
	}
	setBackendLogger(slog.New(handler))
	slog.SetDefault(Slog())

	// logrus 仅作为兼容层，统一转发到 slog backend。
	Logger.SetOutput(io.Discard)
	Logger.SetFormatter(&logrus.JSONFormatter{})
	Logger.ReplaceHooks(make(logrus.LevelHooks))
	Logger.AddHook(&slogForwardHook{backend: Slog()})

	if logLevel == logrus.DebugLevel || logLevel == logrus.TraceLevel {
		Logger.SetReportCaller(true)
	}

	std := logrus.StandardLogger()
	std.SetLevel(logLevel)
	std.SetOutput(io.Discard)
	std.SetReportCaller(Logger.ReportCaller)
	std.ReplaceHooks(make(logrus.LevelHooks))
	std.AddHook(&slogForwardHook{backend: Slog()})

	_ = maxSize
	_ = maxBackups
	_ = maxAge
}

func resolveLogWriter(output, filePath string) (io.Writer, string) {
	switch output {
	case "stderr":
		return os.Stderr, ""
	case "file":
		if filePath == "" {
			return os.Stdout, "logger output=file but file_path is empty, fallback to stdout"
		}
		if err := ensureLogParentDir(filePath); err != nil {
			return os.Stdout, fmt.Sprintf("logger output=file fallback to stdout: mkdir failed file_path=%s err=%v", filePath, err)
		}
		f, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o640)
		if err != nil {
			return os.Stdout, fmt.Sprintf("logger output=file fallback to stdout: open failed file_path=%s err=%v", filePath, err)
		}
		return f, ""
	default:
		return os.Stdout, ""
	}
}

func ensureLogParentDir(filePath string) error {
	parent := strings.TrimSpace(filepathDir(strings.TrimSpace(filePath)))
	if parent == "" || parent == "." {
		return nil
	}
	return os.MkdirAll(parent, 0o755)
}

func filepathDir(p string) string {
	idx := strings.LastIndex(p, "/")
	if idx < 0 {
		idx = strings.LastIndex(p, "\\")
	}
	if idx < 0 {
		return "."
	}
	if idx == 0 {
		return p[:1]
	}
	return p[:idx]
}

func mapLogrusLevel(level logrus.Level) slog.Leveler {
	switch level {
	case logrus.TraceLevel:
		return slog.LevelDebug - 4
	case logrus.DebugLevel:
		return slog.LevelDebug
	case logrus.InfoLevel:
		return slog.LevelInfo
	case logrus.WarnLevel:
		return slog.LevelWarn
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func mapEntryLevel(level logrus.Level) slog.Level {
	switch level {
	case logrus.TraceLevel:
		return slog.LevelDebug - 4
	case logrus.DebugLevel:
		return slog.LevelDebug
	case logrus.InfoLevel:
		return slog.LevelInfo
	case logrus.WarnLevel:
		return slog.LevelWarn
	default:
		return slog.LevelError
	}
}

type slogForwardHook struct {
	backend *slog.Logger
}

func (h *slogForwardHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *slogForwardHook) Fire(entry *logrus.Entry) error {
	logger := h.backend
	if logger == nil {
		logger = Slog()
	}
	attrs := make([]slog.Attr, 0, len(entry.Data)+2)
	for k, v := range entry.Data {
		attrs = append(attrs, slog.Any(k, v))
	}
	if entry.Caller != nil {
		attrs = append(attrs, slog.String("caller", fmt.Sprintf("%s:%d", path.Base(entry.Caller.File), entry.Caller.Line)))
		attrs = append(attrs, slog.String("func", path.Base(entry.Caller.Function)))
	}

	ctx := entry.Context
	if ctx == nil {
		ctx = context.Background()
	}
	logger.LogAttrs(ctx, mapEntryLevel(entry.Level), entry.Message, attrs...)
	return nil
}

func setBackendLogger(l *slog.Logger) {
	backendMu.Lock()
	defer backendMu.Unlock()
	backend = l
}

// Slog 返回统一日志后端。
func Slog() *slog.Logger {
	backendMu.RLock()
	defer backendMu.RUnlock()
	if backend == nil {
		return slog.Default()
	}
	return backend
}

// SetOutput 设置日志输出。
func SetOutput(output io.Writer) {
	if Logger != nil {
		Logger.SetOutput(output)
	}
}

// HTTPAccessEnabled controls request logging in middleware.
func HTTPAccessEnabled() bool {
	return httpAccessEnabled
}

// WithFields 创建带字段的日志条目。
func WithFields(fields logrus.Fields) *logrus.Entry {
	if Logger == nil {
		return logrus.WithFields(fields)
	}
	return Logger.WithFields(fields)
}

// WithField 创建带单个字段的日志条目。
func WithField(key string, value interface{}) *logrus.Entry {
	if Logger == nil {
		return logrus.WithField(key, value)
	}
	return Logger.WithField(key, value)
}

// WithError 创建带错误的日志条目。
func WithError(err error) *logrus.Entry {
	if Logger == nil {
		return logrus.WithError(err)
	}
	return Logger.WithError(err)
}

// New returns a logrus-compatible logger from logger package.
func New() *logrus.Logger {
	if Logger != nil {
		return Logger
	}
	return logrus.New()
}

// StandardLogger returns the global standard logger.
func StandardLogger() *logrus.Logger {
	if Logger != nil {
		return Logger
	}
	return logrus.StandardLogger()
}

// NewEntry creates a new log entry with a target logger.
func NewEntry(l *logrus.Logger) *logrus.Entry {
	if l == nil {
		l = StandardLogger()
	}
	return logrus.NewEntry(l)
}

// Debug 调试日志。
func Debug(args ...interface{}) {
	if Logger == nil {
		logrus.Debug(args...)
		return
	}
	Logger.Debug(args...)
}

// Debugf 格式化调试日志。
func Debugf(format string, args ...interface{}) {
	if Logger == nil {
		logrus.Debugf(format, args...)
		return
	}
	Logger.Debugf(format, args...)
}

// Info 信息日志。
func Info(args ...interface{}) {
	if Logger == nil {
		logrus.Info(args...)
		return
	}
	Logger.Info(args...)
}

// Infof 格式化信息日志。
func Infof(format string, args ...interface{}) {
	if Logger == nil {
		logrus.Infof(format, args...)
		return
	}
	Logger.Infof(format, args...)
}

// Warn 警告日志。
func Warn(args ...interface{}) {
	if Logger == nil {
		logrus.Warn(args...)
		return
	}
	Logger.Warn(args...)
}

// Warnf 格式化警告日志。
func Warnf(format string, args ...interface{}) {
	if Logger == nil {
		logrus.Warnf(format, args...)
		return
	}
	Logger.Warnf(format, args...)
}

// Error 错误日志。
func Error(args ...interface{}) {
	if Logger == nil {
		logrus.Error(args...)
		return
	}
	Logger.Error(args...)
}

// Errorf 格式化错误日志。
func Errorf(format string, args ...interface{}) {
	if Logger == nil {
		logrus.Errorf(format, args...)
		return
	}
	Logger.Errorf(format, args...)
}

// Fatal 致命错误日志。
func Fatal(args ...interface{}) {
	if Logger == nil {
		logrus.Fatal(args...)
		return
	}
	Logger.Fatal(args...)
}

// Fatalf 格式化致命错误日志。
func Fatalf(format string, args ...interface{}) {
	if Logger == nil {
		logrus.Fatalf(format, args...)
		return
	}
	Logger.Fatalf(format, args...)
}

// Panic panic 日志。
func Panic(args ...interface{}) {
	if Logger == nil {
		logrus.Panic(args...)
		return
	}
	Logger.Panic(args...)
}

// Panicf 格式化 panic 日志。
func Panicf(format string, args ...interface{}) {
	if Logger == nil {
		logrus.Panicf(format, args...)
		return
	}
	Logger.Panicf(format, args...)
}

// HTTPMiddleware 创建 HTTP 中间件日志。
func HTTPMiddleware() *logrus.Entry {
	return WithFields(logrus.Fields{"component": "http"})
}

// DBMiddleware 创建数据库中间件日志。
func DBMiddleware() *logrus.Entry {
	return WithFields(logrus.Fields{"component": "database"})
}

// AuthMiddleware 创建认证中间件日志。
func AuthMiddleware() *logrus.Entry {
	return WithFields(logrus.Fields{"component": "auth"})
}

// ServiceLogger 创建服务层日志。
func ServiceLogger(service string) *logrus.Entry {
	return WithFields(logrus.Fields{
		"component": "service",
		"service":   service,
	})
}

// RepoLogger 创建仓储层日志。
func RepoLogger(repo string) *logrus.Entry {
	return WithFields(logrus.Fields{
		"component": "repository",
		"repo":      repo,
	})
}

// HandlerLogger 创建处理器日志。
func HandlerLogger(handler string) *logrus.Entry {
	return WithFields(logrus.Fields{
		"component": "handler",
		"handler":   handler,
	})
}

// callerPrettyForDebug 保留旧行为，便于排错定位。
func callerPrettyForDebug(f *runtime.Frame) (function string, file string) {
	funcName := path.Base(f.Function)
	fileName := path.Base(f.File)
	return funcName, fmt.Sprintf("%s:%d", fileName, f.Line)
}
