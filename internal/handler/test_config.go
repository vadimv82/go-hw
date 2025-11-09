package handler

import (
	"bufio"
	"os"
	"strings"
)

type TestConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
	DSN      string
}

func LoadTestConfig() *TestConfig {
	// Try to load test.env file from project root (ignore error if file doesn't exist)
	// Environment variables take precedence over .env file values
	envFiles := []string{"test.env", "../test.env", "../../test.env"}
	for _, envFile := range envFiles {
		if err := loadEnvFile(envFile); err == nil {
			break
		}
	}

	config := &TestConfig{
		Host:     getEnvOrDefault("TEST_POSTGRES_HOST", "localhost"),
		Port:     getEnvOrDefault("TEST_POSTGRES_PORT", "5433"),
		User:     getEnvOrDefault("TEST_POSTGRES_USER", "postgres"),
		Password: getEnvOrDefault("TEST_POSTGRES_PASSWORD", "postgres"),
		DBName:   getEnvOrDefault("TEST_POSTGRES_DB", "postgres_test"),
		SSLMode:  getEnvOrDefault("TEST_POSTGRES_SSLMODE", "disable"),
		DSN:      os.Getenv("TEST_POSTGRES_DSN"),
	}

	return config
}

// loadEnvFile reads a .env file and sets environment variables
// Only sets variables that are not already set (env vars take precedence)
func loadEnvFile(filename string) error {
	// #nosec G304 -- File paths are hardcoded in LoadTestConfig, not user input
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') ||
			(value[0] == '\'' && value[len(value)-1] == '\'')) {
			value = value[1 : len(value)-1]
		}

		if os.Getenv(key) == "" {
			if err := os.Setenv(key, value); err != nil {
				// Ignore Setenv errors in test configuration (e.g., invalid key format)
				// This is test-only code and errors here won't affect test execution
				continue
			}
		}
	}

	return scanner.Err()
}

func (c *TestConfig) BuildDSN() string {
	if c.DSN != "" {
		return c.DSN
	}

	return "host=" + c.Host +
		" port=" + c.Port +
		" user=" + c.User +
		" password=" + c.Password +
		" dbname=" + c.DBName +
		" sslmode=" + c.SSLMode
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
