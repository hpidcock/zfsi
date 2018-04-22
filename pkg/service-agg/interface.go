package service_agg

import (
	"context"
)

// Service
type Service struct {
	Name string

	Prefixes []string

	Hostname string
	Port     int

	// TODO: Add support for TLS backends
}

// ServiceAggregate provides access to discover services.
type ServiceAggregate interface {
	Services(context.Context) ([]Service, error)
}
