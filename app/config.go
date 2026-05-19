package main

import (
	"os"

	"github.com/joho/godotenv"
)

type ClientConfig struct {
	ServerURL string
	DataDir   string
}

func LoadClientConfig() ClientConfig {
	godotenv.Load()

	return ClientConfig{
		ServerURL: getEnvClient("SERVER_URL", "http://localhost:8080"),
		DataDir:   getEnvClient("DATA_DIR", "./.pochato"),
	}
}

func getEnvClient(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
