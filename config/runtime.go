package config

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// yamlConfig mirrors the YAML config file structure.
// Only non-secret, environment-specific defaults belong here.
type yamlConfig struct {
	AppEnv             string   `yaml:"app_env"`
	Port               string   `yaml:"port"`
	CookieDomain       string   `yaml:"cookie_domain"`
	CookieSecure       bool     `yaml:"cookie_secure"`
	CorsAllowedOrigins []string `yaml:"cors_allowed_origins"`
	GoogleRedirectURL  string   `yaml:"google_redirect_url"`
	RedisEnabled       bool     `yaml:"redis_enabled"`
	RedisUseTLS        bool     `yaml:"redis_use_tls"`
}

type Runtime struct {
	AppEnv             string
	Port               string
	DatabaseURL        string
	CookieDomain       string
	CookieSecure       bool
	CorsAllowedOrigins []string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	RedisEnabled       bool
	RedisAddr          string
	RedisPassword      string
	RedisDB            int
	RedisUseTLS        bool
	RazorpayKey        string
	RazorpaySecret     string
}

var (
	runtimeConfig Runtime
	runtimeOnce   sync.Once
)

// loadYAMLConfig reads config/{env}.yaml and returns the parsed defaults.
// Returns a zero-value yamlConfig if the file does not exist (no error).
func loadYAMLConfig(env string) yamlConfig {
	var cfg yamlConfig

	// Try multiple search paths so it works from project root and from
	// inside Docker (/app/config/*.yaml).
	candidates := []string{
		filepath.Join("config", env+".yaml"),
		filepath.Join("/app", "config", env+".yaml"),
	}

	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			log.Printf("warning: failed to parse %s: %v", path, err)
			return yamlConfig{}
		}
		log.Printf("loaded config from %s", path)
		return cfg
	}

	log.Printf("no YAML config found for env %q (checked %v); using env vars only", env, candidates)
	return cfg
}

func RuntimeConfig() Runtime {
	runtimeOnce.Do(func() {
		_ = godotenv.Load()

		appEnv := getEnv("APP_ENV", "development")

		// --- Load YAML defaults for this environment ---
		yml := loadYAMLConfig(appEnv)

		// --- Build runtime: env vars override YAML defaults ---
		port := getEnv("PORT", yamlDefault(yml.Port, "8080"))

		runtimeConfig = Runtime{
			AppEnv:             appEnv,
			Port:               port,
			DatabaseURL:        strings.TrimSpace(os.Getenv("DATABASE_URL")),
			CookieDomain:       getEnv("COOKIE_DOMAIN", yml.CookieDomain),
			CookieSecure:       getEnvBool("COOKIE_SECURE", yml.CookieSecure),
			CorsAllowedOrigins: getEnvCSV("CORS_ALLOWED_ORIGINS", yamlDefaultSlice(yml.CorsAllowedOrigins, []string{"http://localhost:3010"})),
			GoogleClientID:     strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID")),
			GoogleClientSecret: strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_SECRET")),
			GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", yamlDefault(yml.GoogleRedirectURL, fmt.Sprintf("http://localhost:%s/google_callback", port))),
			RedisEnabled:       resolveRedisEnabledWithDefault(yml.RedisEnabled),
			RedisAddr:          strings.TrimSpace(os.Getenv("REDIS_ADDR")),
			RedisPassword:      os.Getenv("REDIS_PASSWORD"),
			RedisDB:            getEnvInt("REDIS_DB", 0),
			RedisUseTLS:        getEnvBool("REDIS_USE_TLS", yml.RedisUseTLS),
			RazorpayKey:        strings.TrimSpace(os.Getenv("RAZORPAY_KEY")),
			RazorpaySecret:     strings.TrimSpace(os.Getenv("RAZORPAY_SECRET")),
		}
	})

	return runtimeConfig
}

// yamlDefault returns the YAML value if non-empty, otherwise the hard-coded fallback.
func yamlDefault(yamlVal, fallback string) string {
	if yamlVal != "" {
		return yamlVal
	}
	return fallback
}

// yamlDefaultSlice returns the YAML slice if non-empty, otherwise the hard-coded fallback.
func yamlDefaultSlice(yamlVal, fallback []string) []string {
	if len(yamlVal) > 0 {
		return yamlVal
	}
	return fallback
}

func DatabaseDSN() string {
	if databaseURL := RuntimeConfig().DatabaseURL; databaseURL != "" {
		return databaseURL
	}

	return fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		getEnv("DB_HOST", ""),
		getEnv("DB_PORT", ""),
		getEnv("DB_USER", ""),
		getEnv("DB_NAME", ""),
		getEnv("DB_PASSWORD", ""),
		getEnv("DB_SSLMODE", "disable"),
	)
}

func RedisTLSConfig() *tls.Config {
	if !RuntimeConfig().RedisUseTLS {
		return nil
	}

	return &tls.Config{MinVersion: tls.VersionTLS12}
}

func getEnv(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

func getEnvBool(key string, fallback bool) bool {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvInt(key string, fallback int) int {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvCSV(key string, fallback []string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parts := strings.Split(value, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			origins = append(origins, trimmed)
		}
	}

	if len(origins) == 0 {
		return fallback
	}

	return origins
}

func resolveRedisEnabledWithDefault(yamlDefault bool) bool {
	value, ok := os.LookupEnv("REDIS_ENABLED")
	if !ok {
		// If env var is not set, check REDIS_ADDR; if also empty, use YAML default.
		if strings.TrimSpace(os.Getenv("REDIS_ADDR")) != "" {
			return true
		}
		return yamlDefault
	}

	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return false
	}

	return parsed
}
