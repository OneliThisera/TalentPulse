package config

import (
	"fmt"
	"net/url"
	"os"
)

type Config struct {
	Port         string
	DBDriver     string
	DBUser       string
	DBPassword   string
	DBHost       string
	DBPort       string
	DBName       string
	DBPath       string
	JWTSecret    string
	GeminiAPIKey string
}

func LoadConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	driver := os.Getenv("DB_DRIVER")
	if driver == "" {
		driver = "sqlite" // supports "sqlite" and "postgres"
	}

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "talentpulse_admin"
	}

	dbPass := os.Getenv("DB_PASSWORD")
	if dbPass == "" {
		dbPass = "TalentPulse#SecurePass2026"
	}

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "talentpulse_db"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./talentpulse.db"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "talentpulse-jwt-supersecret-token"
	}

	return &Config{
		Port:         port,
		DBDriver:     driver,
		DBUser:       dbUser,
		DBPassword:   dbPass,
		DBHost:       dbHost,
		DBPort:       dbPort,
		DBName:       dbName,
		DBPath:       dbPath,
		JWTSecret:    jwtSecret,
		GeminiAPIKey: os.Getenv("GEMINI_API_KEY"),
	}
}

// GetDSN returns driver and authenticated connection string
func (c *Config) GetDSN() (string, string) {
	if c.DBDriver == "postgres" {
		dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			url.QueryEscape(c.DBUser),
			url.QueryEscape(c.DBPassword),
			c.DBHost,
			c.DBPort,
			c.DBName,
		)
		return "postgres", dsn
	}

	// SQLite connection with authenticated pragma key
	dsn := fmt.Sprintf("%s?_pragma_key=%s&_pragma_cipher=aes256cbc", c.DBPath, url.QueryEscape(c.DBPassword))
	return "sqlite", dsn
}
