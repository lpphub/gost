package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

func Load[T any](configPath, configName, configType string) (*T, error) {
	v := viper.New()

	v.AddConfigPath(configPath)
	v.SetConfigName(configName)
	v.SetConfigType(configType)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := errors.AsType[viper.ConfigFileNotFoundError](err); !ok {
			return nil, fmt.Errorf("read config file failed: %w", err)
		}
	}

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	var cfg T
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config to struct failed: %w", err)
	}

	return &cfg, nil
}

func LoadFile[T any](configFile string) (*T, error) {
	dir := filepath.Dir(configFile)
	file := filepath.Base(configFile)
	ext := filepath.Ext(file)
	name := strings.TrimSuffix(file, ext)
	configType := strings.TrimPrefix(ext, ".")
	return Load[T](dir, name, configType)
}
