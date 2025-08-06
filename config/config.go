package config

import (
	"strings"

	"github.com/spf13/viper"
)

// ReadConfig 读取配置文件
// 支持'$'及'${}'环境变量
func ReadConfig(file string, config interface{}) (err error) {
	v := viper.New()
	v.SetConfigFile(file)
	v.AutomaticEnv()
	v.AllowEmptyEnv(true)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	if err := v.ReadInConfig(); err != nil {
		return err
	}
	return v.Unmarshal(config)
}
