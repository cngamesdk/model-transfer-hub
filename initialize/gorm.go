package initialize

import (
	"fmt"
	"github.com/cngamesdk/model-transfer-hub/global"
	"github.com/cngamesdk/model-transfer-hub/model"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

// Gorm 初始化数据库连接
func Gorm() *gorm.DB {
	switch global.MTH_CONFIG.Database.Type {
	case "mysql":
		return GormMysql()
	default:
		return GormMysql()
	}
}

// GormMysql 初始化MySQL连接
func GormMysql() *gorm.DB {
	config := global.MTH_CONFIG.Database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.DbName,
	)

	// GORM配置
	gormConfig := &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   getGormLogger(config.LogMode),
	}

	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		global.MTH_LOG.Error("数据库连接失败", zap.Error(err))
		panic(any(fmt.Sprintf("数据库连接失败: %v", err)))
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 自动迁移表结构
	if err := autoMigrate(db); err != nil {
		global.MTH_LOG.Error("数据库迁移失败", zap.Error(err))
	}

	global.MTH_LOG.Info("数据库连接成功")
	return db
}

// getGormLogger 获取GORM日志配置
func getGormLogger(logMode string) logger.Interface {
	var logLevel logger.LogLevel
	switch logMode {
	case "silent":
		logLevel = logger.Silent
	case "error":
		logLevel = logger.Error
	case "warn":
		logLevel = logger.Warn
	case "info":
		logLevel = logger.Info
	default:
		logLevel = logger.Error
	}

	return logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)
}

// autoMigrate 自动迁移表结构
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.AiToken{},
		&model.AiUsageLog{},
	)
}
