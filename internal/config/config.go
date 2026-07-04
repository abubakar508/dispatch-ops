package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port             string
	OSRMBaseURL      string
	OSRMTimeout      time.Duration
	OptimizerTimeout time.Duration
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	ShutdownTimeout  time.Duration
	LogLevel         string
	Environment      string
	AllowedOrigin    string
	NominatimBaseURL string
	NominatimAgent   string
	GeocodeTimeout   time.Duration
	H3Resolution     int
}

func Load() (Config, error) {
	cfg := Config{
		Port:             getString("PORT", "8080"),
		OSRMBaseURL:      getString("OSRM_BASE_URL", "https://router.project-osrm.org"),
		OSRMTimeout:      getDuration("OSRM_TIMEOUT", 20*time.Second),
		OptimizerTimeout: getDuration("OPTIMIZER_TIMEOUT", 10*time.Second),
		ReadTimeout:      getDuration("HTTP_READ_TIMEOUT", 15*time.Second),
		WriteTimeout:     getDuration("HTTP_WRITE_TIMEOUT", 60*time.Second),
		ShutdownTimeout:  getDuration("SHUTDOWN_TIMEOUT", 15*time.Second),
		LogLevel:         getString("LOG_LEVEL", "info"),
		Environment:      getString("ENVIRONMENT", "production"),
		AllowedOrigin:    getString("ALLOWED_ORIGIN", ""),
		NominatimBaseURL: getString("NOMINATIM_BASE_URL", "https://nominatim.openstreetmap.org"),
		NominatimAgent:   getString("NOMINATIM_USER_AGENT", "kasi-delivery/1.0 (contact: ops@kasi.delivery)"),
		GeocodeTimeout:   getDuration("GEOCODE_TIMEOUT", 10*time.Second),
		H3Resolution:     getInt("H3_RESOLUTION", 9),
	}

	if cfg.OSRMBaseURL == "" {
		return Config{}, fmt.Errorf("config: OSRM_BASE_URL must not be empty")
	}

	return cfg, nil
}

func getString(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
		if secs, err := strconv.Atoi(v); err == nil {
			return time.Duration(secs) * time.Second
		}
	}
	return fallback
}
