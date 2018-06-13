package service_agg

import (
	"context"
	"time"
)

// Service
type Service struct {
	Name string

	Prefixes []string

	Hostname string
	Port     int

	Timeout time.Duration

	// TODO: Add support for TLS backends
}

// ServiceAggregate provides access to discover services.
type ServiceAggregate interface {
	Services(context.Context) ([]Service, error)
}
