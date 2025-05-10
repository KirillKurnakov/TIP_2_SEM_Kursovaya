package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config содержит все переменные окружения приложения
type Config struct {
	AppEnv      string
	ServerPort  string
	DBHost      string
	DBUser      string
	DBPassword  string
	DBName      string
	DBPort      string
}

// Load загружает переменные из .env (если есть) и возвращает структуру конфигурации
func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println(".env файл не найден, загружаем переменные из окружения")
	}

	return &Config{
		AppEnv:     os.Getenv("APP_ENV"),
		ServerPort: os.Getenv("SERVER_PORT"),
		DBHost:     os.Getenv("DB_HOST"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		DBPort:     os.Getenv("DB_PORT"),
	}
}