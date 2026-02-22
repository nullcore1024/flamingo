package base

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"sync"
)

type ConfigReader struct {
	v *viper.Viper
}

var (
	configInstance *ConfigReader
	configOnce     sync.Once
)

func GetConfigReader() *ConfigReader {
	configOnce.Do(func() {
		v := viper.New()
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./configs")
		v.AddConfigPath(".")

		if err := v.ReadInConfig(); err != nil {
			GetLogger().Error("Failed to read config file", zap.Error(err))
		}

		configInstance = &ConfigReader{v: v}
	})
	return configInstance
}

func (cr *ConfigReader) GetString(key string) string {
	return cr.v.GetString(key)
}

func (cr *ConfigReader) GetInt(key string) int {
	return cr.v.GetInt(key)
}

func (cr *ConfigReader) GetBool(key string) bool {
	return cr.v.GetBool(key)
}

func (cr *ConfigReader) GetStringSlice(key string) []string {
	return cr.v.GetStringSlice(key)
}

func (cr *ConfigReader) GetMapString(key string) map[string]interface{} {
	return cr.v.GetStringMap(key)
}

func (cr *ConfigReader) IsSet(key string) bool {
	return cr.v.IsSet(key)
}
