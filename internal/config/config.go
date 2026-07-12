package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	AllowedOrigin string
	MongoDBName   string
	MongoURI      string
	Port          string
	Zoho          ZohoConfig
	ZohoSheets    ZohoSheetsConfig
}

type ZohoConfig struct {
	AccountID    string
	ClientID     string
	ClientSecret string
	Timeout      time.Duration
}

type ZohoSheetConfig struct {
	ID   string
	Name string
}

type ZohoSheetsConfig struct {
	Career     ZohoSheetConfig
	Newsletter ZohoSheetConfig
	Sale       ZohoSheetConfig
}

func Load() Config {
	_ = loadDotEnvFile()
	return Config{
		AllowedOrigin: getEnv("ALLOWED_ORIGIN", "*"),
		MongoDBName:   getEnv("MONGODB_DB_NAME", "xefort"),
		MongoURI:      getEnv("MONGODB_URI", "mongodb://127.0.0.1:27017"),
		Port:          getEnv("PORT", "3000"),
		Zoho: ZohoConfig{
			AccountID:    getEnv("ZOHO_ACCOUNT_ID", ""),
			ClientID:     getEnv("ZOHO_CLIENT_ID", ""),
			ClientSecret: getEnv("ZOHO_CLIENT_SECRET", ""),
			Timeout:      20 * time.Second,
		},
		ZohoSheets: ZohoSheetsConfig{
			Career: ZohoSheetConfig{
				ID:   getEnv("CAREER_SPREADSHEET_ID", ""),
				Name: getEnv("CAREER_SPREADSHEET_NAME", ""),
			},
			Newsletter: ZohoSheetConfig{
				ID:   getEnv("NEWSLETTER_SPREADSHEET_ID", ""),
				Name: getEnv("NEWSLETTER_SPREADSHEET_NAME", ""),
			},
			Sale: ZohoSheetConfig{
				ID:   getEnv("SALE_SPREADSHEET_ID", ""),
				Name: getEnv("SALE_SPREADSHEET_NAME", ""),
			},
		},
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func loadDotEnvFile() error {
	dotEnvPath, err := findDotEnvFile()
	if err != nil {
		return err
	}
	if dotEnvPath == "" {
		return nil
	}
	data, err := os.ReadFile(dotEnvPath)
	if err != nil {
		return err
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key == "" {
			continue
		}
		if existingValue, exists := os.LookupEnv(key); exists && existingValue != "" {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}
	return nil
}

func findDotEnvFile() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("Get working directory: %w", err)
	}
	for dir := currentDir; dir != filepath.Dir(dir); dir = filepath.Dir(dir) {
		candidate := filepath.Join(dir, ".env")
		info, err := os.Stat(candidate)
		if err == nil && !info.IsDir() {
			return candidate, nil
		}
		if err != nil && !os.IsNotExist(err) {
			return "", fmt.Errorf("Check .env file: %w", err)
		}
	}
	return "", nil
}
