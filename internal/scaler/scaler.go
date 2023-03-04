package scaler

import "context"

type Type string

type Interface interface {
	Type() Type

	// Register is used to register the scaler with waymond core
	// It is executed exactly once for a given scaler
	// i.e. when waymond boots up
	Register(ctx context.Context) error
}
