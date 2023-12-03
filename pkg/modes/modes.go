package modes

import (
	"context"
)

// Mode is a mode of operation
type Mode interface {
	// Run runs the mode with the given context
	Run(ctx context.Context) error
}
