package transport

import (
	"bank_test/internal/db"
	"bank_test/internal/transport/http"

	"go.uber.org/zap"
)

// Transporter is an interface for the transport layer. It defines the Serve method that
// will be run by any transport layer implementation (HTTP, gRPC, GraphQL, etc.).
type Transporter interface {
	Serve() error       // starts the transport layer
	HealthCheck() error // starts a health check endpoint that verifies the service is up and running
	Close() error       // handles the graceful shutdown of the transport layer
}

// NewTransporter creates a new transport layer based on the provided type.
func NewTransporter(logger *zap.SugaredLogger, db db.DatabaseAdapter) Transporter {
	return http.NewHttpTransport(logger, db)
}
