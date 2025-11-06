package config

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

type (
	Config struct {
		LogLevel  zerolog.Level
		Server    *ServerConfig
		Postgres  *PostgresConfig
		Handler   *HandlerConfig
		JWT       *JWTConfig
		Embedding *EmbeddingConfig
	}

	PostgresConfig struct {
		Host     string
		User     string
		Password string
		DB       string
		Port     int
		SSLMode  string
	}

	ServerConfig struct {
		Port           int
		ReadTimeout    time.Duration
		WriteTimeout   time.Duration
		MaxHeaderBytes int
	}

	HandlerConfig struct {
		RequestTimeout time.Duration
	}

	JWTConfig struct {
		Secret string
	}

	EmbeddingConfig struct {
		Host string
		Port int
	}
)

func Init(cfgPath, globalCfgPath, envPath string) (*Config, error) {
	v := viper.New()

	v.AddConfigPath(filepath.Dir(globalCfgPath))
	v.SetConfigName(filepath.Base(globalCfgPath))
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("ошибка чтения глобального конфига '%s': %w", globalCfgPath, err)
	}

	v.AddConfigPath(filepath.Dir(cfgPath))
	v.SetConfigName(filepath.Base(cfgPath))
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("ошибка чтения локального конфига '%s': %w", cfgPath, err)
		}
	}

	v.SetConfigFile(envPath)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read .env file: %w", err)
		}
	}

	v.AutomaticEnv()

	level, err := zerolog.ParseLevel(v.GetString("log_level"))
	if err != nil {
		return nil, fmt.Errorf("не удалось распознать log_level: '%s': %w", v.GetString("log_level"), err)
	}

	return &Config{
		LogLevel: level,
		Server: &ServerConfig{
			Port:           v.GetInt("backend.port"),
			ReadTimeout:    v.GetDuration("server.readTimeout"),
			WriteTimeout:   v.GetDuration("server.writeTimeout"),
			MaxHeaderBytes: v.GetInt("server.maxHeaderBytes"),
		},
		Postgres: &PostgresConfig{
			Host:     v.GetString("POSTGRES_HOST"),
			User:     v.GetString("POSTGRES_USER"),
			Password: v.GetString("POSTGRES_PASSWORD"),
			DB:       v.GetString("POSTGRES_DB"),
			Port:     v.GetInt("POSTGRES_PORT"),
			SSLMode:  v.GetString("POSTGRES_SSLMODE"),
		},
		Handler: &HandlerConfig{
			RequestTimeout: v.GetDuration("handler.requestTimeout"),
		},
		JWT: &JWTConfig{
			Secret: v.GetString("JWT_SECRET"),
		},
		Embedding: &EmbeddingConfig{
			Host: v.GetString("embedding-service.host"),
			Port: v.GetInt("embedding-service.port"),
		},
	}, nil
}

func (p *PostgresConfig) GetUrl() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.Host, p.Port, p.User, p.Password, p.DB, p.SSLMode)
}

func (e *EmbeddingConfig) GetUrl() string {
	return fmt.Sprintf("http://%s:%d",
		e.Host, e.Port)
}
