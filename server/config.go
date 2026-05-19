package main

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabasePath string
	ServerPort   string
	JWTSecret    string
}

func LoadConfig() Config {
	godotenv.Load()

	return Config{
		DatabasePath: getEnv("DATABASE_PATH", "./pochato.db"),
		ServerPort:   getEnv("SERVER_PORT", ":8080"),
		JWTSecret:    getEnv("JWT_SECRET", "pochato-secret-key-change-in-production"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
