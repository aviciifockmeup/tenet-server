package config

import (
    "github.com/spf13/viper"
)

// Config 全局配置
type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
    Log      LogConfig      `mapstructure:"log"`
}

type ServerConfig struct {
    Port int    `mapstructure:"port"`
    Name string `mapstructure:"name"`
}

type DatabaseConfig struct {
    MySQL MySQLConfig `mapstructure:"mysql"`
    Redis RedisConfig `mapstructure:"redis"`
}

type MySQLConfig struct {
    Host     string `mapstructure:"host"`
    Port     int    `mapstructure:"port"`
    Username string `mapstructure:"username"`
    Password string `mapstructure:"password"`
    Database string `mapstructure:"database"`
}

type RedisConfig struct {
    Host string `mapstructure:"host"`
    Port int    `mapstructure:"port"`
    DB   int    `mapstructure:"db"`
}

type LogConfig struct {
    Level string `mapstructure:"level"`
}

func Load(path string) (*Config, error) {
	viper.SetConfigFile(path)
	
	if err:= viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err:= viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}