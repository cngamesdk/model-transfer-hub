package core

import (
	"fmt"
	"github.com/cngamesdk/model-transfer-hub/global"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Viper 初始化配置文件
func Viper(path ...string) *viper.Viper {
	var configFile string
	if len(path) == 0 {
		configFile = "config.yaml"
	} else {
		configFile = path[0]
	}

	v := viper.New()
	v.SetConfigFile(configFile)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		panic(any(fmt.Errorf("Fatal error config file: %s \n", err)))
	}

	// 监听配置文件变化
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("配置文件被修改:", e.Name)
		if err := v.Unmarshal(&global.MTH_CONFIG); err != nil {
			fmt.Println("配置文件解析失败:", err)
		}
	})

	if err := v.Unmarshal(&global.MTH_CONFIG); err != nil {
		panic(any(fmt.Errorf("解析配置文件失败: %s \n", err)))
	}

	return v
}
