package config

import (
	"github.com/plaid/go-envvar/envvar"
)

// App holds the application-wide configuration.
type App struct {
	HTTPServer *HTTPServer
	GRPCServer *GRPCServer
}

// HTTPServer holds the configuration for the HTTP server.
type HTTPServer struct {
	Host string
	Port string
}

// GRPCServer holds the configuration for the gRPC server.
type GRPCServer struct {
	Host string
	Port string
}

// New returns the application-wide configuration.
func New() (*App, error) {
	var cfg App

	if err := envvar.Parse(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
