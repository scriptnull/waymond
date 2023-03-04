package trigger

import "context"

type Type string

type Interface interface {
	Type() Type

	// Register is used to register the trigger with waymond core
	// It is executed exactly once for a given trigger
	// i.e. when waymond boots up
	Register(ctx context.Context) error
}
