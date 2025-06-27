// Package api provides the handlers/middlewares
package api

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/maniac-en/chirpstack/internal/database"
)

type Platform string

const (
	PlatformDev  Platform = "dev"
	PlatformProd Platform = "prod"
)

func (p Platform) IsValid() bool {
	switch p {
	case PlatformDev, PlatformProd:
		return true
	default:
		return false
	}
}

func ValidPlatforms() []Platform {
	return []Platform{PlatformDev, PlatformProd}
}

func ParsePlatform(s string) (Platform, error) {
	p := Platform(s)
	if !p.IsValid() {
		return "", fmt.Errorf("invalid platform '%s', must be one of: %v", s, ValidPlatforms())
	}
	return p, nil
}

type APIConfig struct {
	fileserverHits atomic.Int32
	DB             *database.Queries
	Platform       Platform
	JWTTokenSecret string
	PolkaAPIKey    string
}

func (cfg *APIConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		w.Header().Add("Cache-Control", "no-cache")
		next.ServeHTTP(w, r)
	})
}

func (cfg *APIConfig) LogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
