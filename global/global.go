package global

import (
	"github.com/cngamesdk/model-transfer-hub/config"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	MTH_DB     *gorm.DB      // 数据库连接
	MTH_CONFIG config.Server // 配置对象
	MTH_VP     *viper.Viper  // Viper实例
	MTH_LOG    *zap.Logger   // 日志实例
)
