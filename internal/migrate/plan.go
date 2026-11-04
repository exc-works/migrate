package migrate

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/exc-works/migrate/internal/parser"
	"github.com/exc-works/migrate/internal/script"
	"github.com/exc-works/migrate/internal/source"
)

type plannedMigration struct {
	migration Migration
	modules   *source.ModuleContext
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
		if !script.SupportedDialect(s.cfg.Dialect.Name()) {
			return fmt.Errorf("script migration %s: unsupported dialect %s", m.Filename, s.cfg.Dialect.Name())
		}
	}
	return nil
}

// preparePlan validates every selected step before the first transaction. Script
// instances are initialized without DB access and retained, never initialized twice.
func (s *Service) preparePlan(migrations []Migration, direction string) ([]plannedMigration, error) {
	return s.preparePlanWithModules(migrations, direction, nil)
}

func (s *Service) preparePlanWithModules(migrations []Migration, direction string, modules map[string]*source.ModuleContext) ([]plannedMigration, error) {
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
			p, err := script.Compile(m.Filename, m.Source, direction, s.cfg.ScriptOptions, modules[m.Version])
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
			if invocation != nil && !invocation.TransactionEnabled() {
				s.cfg.Logger.Warnf("nontransactional script file=%s: no automatic rollback; failures may leave partial changes without history", m.Filename)
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
			modules:   modules[m.Version],
			step:      step,
		})
	}
	return plan, nil
}

// withStepTransaction uses the selected script's transaction policy for data and
// history. SQL migrations retain their existing transaction behavior.
func (s *Service) withStepTransaction(step migrationStep, fn func(execContextRunner) error) (err error) {
	scriptMigration, isScript := step.(scriptStep)
	disabled := isScript && scriptMigration.invocation != nil && !scriptMigration.invocation.TransactionEnabled()
	nontransactional := disabled
	if isScript {
		capability, _ := script.Capability(s.cfg.Dialect.Name())
		nontransactional = nontransactional || capability.TransactionMode == script.TransactionNone
	}
	defer func() {
		if recover() != nil {
			err = errors.Join(errors.New("migration execution panicked"), err)
		}
		if err != nil && isScript && nontransactional {
			err = fmt.Errorf("%w; nontransactional script may have left partial changes; inspect data and migration history before retrying", err)
		}
	}()
	if disabled && !s.cfg.DryRun {
		return s.withScriptConnection(fn)
	}
	return s.withTx(fn)
}

// One connection preserves manual transactions and session settings across calls.
// Return successful connections to the pool; database/sql and the driver validate
// them before reuse. Failed or panicked steps may leave a manual transaction, so
// discard those connections instead. Authors finish transactions and restore any
// session settings they change before a successful return.
func (s *Service) withScriptConnection(fn func(execContextRunner) error) (err error) {
	conn, err := s.cfg.DB.Conn(s.ctx)
	if err != nil {
		return err
	}
	reusable := false
	defer func() {
		if !reusable {
			discardErr := conn.Raw(func(any) error { return driver.ErrBadConn })
			if discardErr != nil && !errors.Is(discardErr, driver.ErrBadConn) && !errors.Is(discardErr, sql.ErrConnDone) {
				err = errors.Join(err, errors.New("discard migration connection failed"))
			}
		}
		if closeErr := conn.Close(); closeErr != nil && !errors.Is(closeErr, sql.ErrConnDone) {
			err = errors.Join(err, errors.New("close migration connection failed"))
		}
	}()
	err = fn(conn)
	reusable = err == nil
	return err
}
