package bootstrap

import (
	"context"
	"errors"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/config"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/db"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/logger"
	"gorm.io/gorm"
)

func BootstrapPlugin(ctx context.Context, cfg *config.Config) (*gorm.DB, error) {
	if cfg == nil {
		return nil, errors.New("config is required")
	}
	if cfg.Database == nil {
		return nil, errors.New("database config is required")
	}

	// 初始化日志
	logLevel := cfg.LogLevel
	logFormat := "json"
	logOutput := "stdout"
	logFile := ""
	maxSize := 100
	maxBackups := 3
	maxAge := 28
	httpAccess := true
	if cfg.Logging != nil {
		if cfg.Logging.Level != "" {
			logLevel = cfg.Logging.Level
		}
		if cfg.Logging.Format != "" {
			logFormat = cfg.Logging.Format
		}
		if cfg.Logging.Output != "" {
			logOutput = cfg.Logging.Output
		}
		logFile = cfg.Logging.FilePath
		maxSize = cfg.Logging.MaxSize
		maxBackups = cfg.Logging.MaxBackups
		maxAge = cfg.Logging.MaxAge
	}
	logger.InitWithHostMode(logLevel, logFormat, logOutput, logFile, maxSize, maxBackups, maxAge, httpAccess, envTruthy("POWERX_PROXY"))
	logger.Info("Starting PowerX Note Plugin...")

	// 初始化 schema
	models.InitSchemaFrom(cfg.Database.Schema)

	// 连接数据库（在进程生命周期内保持打开；在优雅退出时关闭）
	queryDB, err := db.Connect(cfg.Database)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to database")
	}

	return queryDB, nil
}
