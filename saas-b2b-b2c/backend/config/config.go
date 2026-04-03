package config

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config содержит все настройки приложения
type Config struct {
	ServerPort  string
	DatabaseURL string
	JWTSecret   string
	JWTExpires  time.Duration
	RedisURL    string
	Debug       bool
	
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
}

// LoadConfig загружает конфигурацию из различных источников
func LoadConfig() *Config {
	viper.SetDefault("server_port", "8080")
	viper.SetDefault("jwt_expires", "24h")
	viper.SetDefault("debug", false)

	// Загружаем из переменных окружения
	viper.AutomaticEnv()

	// Создаем экземпляр конфигурации
	config := &Config{
		ServerPort: viper.GetString("SERVER_PORT"),
		JWTSecret:  viper.GetString("JWT_SECRET"),
		RedisURL:   viper.GetString("REDIS_URL"),
		Debug:      viper.GetBool("DEBUG"),
		
		DBHost:     viper.GetString("DB_HOST"),
		DBPort:     viper.GetString("DB_PORT"),
		DBUser:     viper.GetString("DB_USER"),
		DBPassword: viper.GetString("DB_PASSWORD"),
		DBName:     viper.GetString("DB_NAME"),
	}

	// Устанавливаем значения по умолчанию, если они не заданы
	if config.ServerPort == "" {
		config.ServerPort = "8080"
	}
	
	if config.JWTSecret == "" {
		config.JWTSecret = "CHANGE_ME_JWT_SECRET"
	}
	
	if config.RedisURL == "" {
		config.RedisURL = "redis://localhost:6379"
	}
	
	if config.DBHost == "" {
		config.DBHost = "localhost"
	}
	
	if config.DBPort == "" {
		config.DBPort = "5432"
	}
	
	if config.DBUser == "" {
		config.DBUser = "postgres"
	}
	
	if config.DBPassword == "" {
		config.DBPassword = "password"
	}
	
	if config.DBName == "" {
		config.DBName = "franchise_db"
	}

	// Формируем URL для подключения к базе данных
	config.DatabaseURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.DBUser, config.DBPassword, config.DBHost, config.DBPort, config.DBName)

	// Парсим продолжительность жизни JWT токена
	jwtExpiresStr := viper.GetString("JWT_EXPIRES")
	if jwtExpiresStr == "" {
		jwtExpiresStr = "24h"
	}
	
	jwtExpires, err := time.ParseDuration(jwtExpiresStr)
	if err != nil {
		jwtExpires = 24 * time.Hour // значение по умолчанию
	}
	config.JWTExpires = jwtExpires

	return config
}

// DBContext возвращает контекст с таймаутом для операций с базой данных
func (c *Config) DBContext() context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	go func() {
		time.Sleep(10 * time.Second)
		cancel()
	}()
	return ctx
}