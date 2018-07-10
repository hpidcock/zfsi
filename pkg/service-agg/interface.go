package service_agg

import (
	"context"
	"time"
)

// Host
type Host struct {
	Hostname string
	Port     int
}

// Service
type Service struct {
	Name string

	Prefixes []string

	Hosts []Host

	Timeout time.Duration

	// TODO: Add support for TLS backends
}

// ServiceAggregate provides access to discover services.
type ServiceAggregate interface {
	Services(context.Context) ([]Service, error)
}
