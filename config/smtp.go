package config

import (
	"os"
	"strconv"
)

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

func LoadSMTPConfig() SMTPConfig {
	return SMTPConfig{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     getEnvAsInt("SMTP_PORT", 587),
		Username: os.Getenv("SMTP_USERNAME"),
		Password: os.Getenv("SMTP_PASSWORD"),
	}
}

func getEnvAsInt(name string, defaultVal int) int {
	valueStr := os.Getenv(name)
	if valueStr == "" {
		return defaultVal
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultVal
	}
	return value
}
