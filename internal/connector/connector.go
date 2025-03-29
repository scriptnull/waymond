package connector

import "context"

type Type string

type Interface interface {
	Type() Type

	// Register is used to register the connector with waymond core
	// It is executed exactly once for a given connection
	// i.e. when waymond boots up
	Register(ctx context.Context) error

	From() string
	To() string
}
