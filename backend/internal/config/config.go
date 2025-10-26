package config

import (
	"fmt"
	"os"
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
		SSLMode  bool
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
		Url string
	}
)

func Init(configPath string) (*Config, error) {
	v := viper.New()

	v.AddConfigPath(filepath.Dir(configPath))
	v.SetConfigName(filepath.Base(configPath))
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("ошибка чтения базового конфига '%s': %w", configPath, err)
	}

	v.SetConfigFile(".env")
	if err := v.MergeInConfig(); err != nil {
		if _, ok := err.(*os.PathError); !ok {
			return nil, fmt.Errorf("ошибка парсинга .env файла: %w", err)
		}
	}

	v.AutomaticEnv()

	// // --- ОТЛАДОЧНЫЙ БЛОК ---
	// // Получаем все настройки после слияния всех источников
	// allSettings := v.AllSettings()

	// // Преобразуем в красивый JSON для вывода
	// settingsJson, err := json.MarshalIndent(allSettings, "", "  ")
	// if err != nil {
	// 	log.Fatalf("Ошибка при преобразовании настроек в JSON: %v", err)
	// }

	// // Выводим в консоль
	// fmt.Println("--- Загруженная конфигурация Viper ---")
	// fmt.Println(string(settingsJson))
	// fmt.Println("------------------------------------")
	// // --- КОНЕЦ ОТЛАДОЧНОГО БЛОКА ---

	level, err := zerolog.ParseLevel(v.GetString("log_level"))
	if err != nil {
		return nil, fmt.Errorf("не удалось распознать log_level: %w", err)
	}

	return &Config{
		LogLevel: level,
		Server: &ServerConfig{
			Port:           v.GetInt("server.port"),
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
			SSLMode:  v.GetBool("POSTGRES_SSLMODE"),
		},
		Handler: &HandlerConfig{
			RequestTimeout: v.GetDuration("handler.requestTimeout"),
		},
		JWT: &JWTConfig{
			Secret: v.GetString("JWT_SECRET"),
		},
		Embedding: &EmbeddingConfig{Url: v.GetString("embedding.url")},
	}, nil
}

func (p *PostgresConfig) GetUrl() string {
	sslMode := "disable"
	if p.SSLMode {
		sslMode = "require"
	}
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.Host, p.Port, p.User, p.Password, p.DB, sslMode)
}
