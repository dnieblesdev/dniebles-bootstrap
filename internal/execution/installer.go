package execution

import (
	"context"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
)

// Installer is the execution contract for a single resource kind. Implementations
// perform the actual installation work for plan steps matching SupportedKind.
type Installer interface {
	SupportedKind() planning.ResourceKind
	Install(context.Context, planning.PlanStep) StepResult
}
