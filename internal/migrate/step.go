package migrate

import (
	"context"
	"fmt"
	"io"
)

// migrationStep separates execution from planning and history bookkeeping.
type migrationStep interface {
	Execute(context.Context, execContextRunner) error
	WritePlan(io.Writer) error
}

type sqlStep struct {
	migration  Migration
	direction  string
	statements []string
}

func (step sqlStep) Execute(ctx context.Context, runner execContextRunner) error {
	for _, stmt := range step.statements {
		if _, err := runner.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("apply %s %s version=%s: %w", step.direction, step.migration.Filename, step.migration.Version, err)
		}
	}
	return nil
}

func (step sqlStep) WritePlan(w io.Writer) error {
	for _, stmt := range step.statements {
		if _, err := io.WriteString(w, stmt+"\n"); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) executeStep(runner execContextRunner, step migrationStep) error {
	if s.cfg.DryRun {
		return step.WritePlan(s.cfg.DryRunOutput)
	}
	return step.Execute(s.ctx, runner)
}
