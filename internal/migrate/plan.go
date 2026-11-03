package migrate

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/exc-works/migrate/internal/parser"
	"github.com/exc-works/migrate/internal/script"
	"github.com/exc-works/migrate/internal/source"
)

type plannedMigration struct {
	migration Migration
	step      migrationStep
}

type scriptStep struct {
	migration  Migration
	direction  string
	dialect    string
	invocation *script.Invocation
}

func (step scriptStep) Execute(ctx context.Context, runner execContextRunner) error {
	return step.invocation.Execute(ctx, runner, step.dialect)
}

func (step scriptStep) WritePlan(w io.Writer) error {
	_, err := fmt.Fprintf(w, "-- script %s %s (not executed)\n", step.direction, step.migration.Filename)
	return err
}

func (s *Service) checkCapabilities(migrations []Migration) error {
	for _, m := range migrations {
		if !source.IsSupportedFilename(m.Filename) {
			return fmt.Errorf("unsupported migration filename: %s", m.Filename)
		}
		if filepath.Ext(m.Filename) == source.ExtensionSQL {
			continue
		}
		if len(script.Formats()) == 1 {
			return fmt.Errorf("migration %s: %w", m.Filename, script.ErrRequiresFull)
		}
		if !script.SupportedDialect(s.cfg.Dialect.Name()) {
			return fmt.Errorf("script migration %s: unsupported dialect %s", m.Filename, s.cfg.Dialect.Name())
		}
	}
	return nil
}

// preparePlan validates every selected step before the first transaction. Script
// instances are initialized without DB access and retained, never initialized twice.
func (s *Service) preparePlan(migrations []Migration, direction string) ([]plannedMigration, error) {
	plan := make([]plannedMigration, 0, len(migrations))
	warned := false
	for _, m := range migrations {
		var step migrationStep
		if filepath.Ext(m.Filename) == source.ExtensionSQL {
			statements := m.UpStatements
			if direction == parser.DirectionDownName {
				statements = m.DownStatements
			}
			if len(statements) == 0 {
				return nil, fmt.Errorf("no %s statements found for %s", direction, m.Filename)
			}
			step = sqlStep{
				migration:  m,
				direction:  direction,
				statements: statements,
			}
		} else {
			if !warned {
				capability, _ := script.Capability(s.cfg.Dialect.Name())
				if capability.Verification == script.VerificationExperimental {
					s.cfg.Logger.Warnf("experimental script dialect=%s: %s", capability.Name, capability.Notes)
				}
				warned = true
			}
			p, err := script.Compile(m.Filename, m.Source, direction, s.cfg.ScriptOptions)
			if err != nil {
				return nil, err
			}
			var invocation *script.Invocation
			if !s.cfg.DryRun {
				invocation, err = p.Instantiate(s.ctx)
				if err != nil {
					return nil, err
				}
			}
			step = scriptStep{
				migration:  m,
				direction:  direction,
				dialect:    s.cfg.Dialect.Name(),
				invocation: invocation,
			}
		}
		plan = append(plan, plannedMigration{
			migration: m,
			step:      step,
		})
	}
	return plan, nil
}

// withStepTransaction includes history writes in the failure boundary. A failed
// nontransactional script can have persisted data even without a history record.
func (s *Service) withStepTransaction(step migrationStep, fn func(execContextRunner) error) error {
	err := s.withTx(fn)
	if _, isScript := step.(scriptStep); err != nil && isScript {
		capability, _ := script.Capability(s.cfg.Dialect.Name())
		if capability.TransactionMode == script.TransactionNone {
			return fmt.Errorf("%w; nontransactional script may have left partial changes; inspect data and migration history before retrying", err)
		}
	}
	return err
}
