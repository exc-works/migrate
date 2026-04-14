package migrate

import (
	"context"
	"crypto/md5"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/exc-works/migrate/internal/dialect"
	"github.com/exc-works/migrate/internal/logger"
	"github.com/exc-works/migrate/internal/parser"
	"github.com/exc-works/migrate/internal/source"
)

type Config struct {
	Dialect dialect.Dialect
	DB      *sql.DB
	Logger  logger.Logger

	SchemaName        string
	MigrateOutOfOrder bool
	MigrationSource   source.MigrationSource
	DryRun            bool
	DryRunOutput      io.Writer
}

type Service struct {
	ctx context.Context
	cfg Config
}

type execContextRunner interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func (s *Service) Config() Config {
	return s.cfg
}

func NewService(ctx context.Context, cfg Config) (*Service, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if cfg.Dialect == nil {
		return nil, errors.New("dialect is required")
	}
	if cfg.DB == nil {
		return nil, errors.New("db is required")
	}
	if cfg.SchemaName == "" {
		cfg.SchemaName = "migration_schema"
	}
	if err := dialect.ValidateIdentifier(cfg.SchemaName); err != nil {
		return nil, err
	}
	if cfg.MigrationSource == nil {
		return nil, errors.New("migration source is required")
	}
	if cfg.Logger == nil {
		cfg.Logger = logger.NoopLogger{}
	}
	if cfg.DryRunOutput == nil {
		cfg.DryRunOutput = io.Discard
	}
	return &Service{ctx: ctx, cfg: cfg}, nil
}

func (s *Service) Create() error {
	stmt, err := s.cfg.Dialect.CreateSchemaSQL(s.cfg.SchemaName)
	if err != nil {
		return err
	}
	_, err = s.cfg.DB.ExecContext(s.ctx, stmt)
	return err
}

func (s *Service) Status() ([]MigrationStatus, error) {
	migrations, err := s.loadMigrations(false, false)
	if err != nil {
		return nil, err
	}
	schemas, _, err := s.loadSchemas()
	if err != nil {
		return nil, err
	}
	statuses, _ := buildStatuses(migrations, schemas)
	return statuses, nil
}

func (s *Service) Baseline() error {
	migrations, err := s.loadMigrations(false, false)
	if err != nil {
		return err
	}
	schemas, maxID, err := s.loadSchemas()
	if err != nil {
		return err
	}
	statuses, _ := buildStatuses(migrations, schemas)
	if len(statuses) == 0 {
		return errors.New("no migration files")
	}
	for _, st := range statuses {
		if st.Status != StatusPending {
			return fmt.Errorf("baseline only supports pending migration files")
		}
	}

	insertSQL, err := s.cfg.Dialect.InsertSchemaSQL(s.cfg.SchemaName)
	if err != nil {
		return err
	}

	nextID := maxID
	err = s.withTx(func(runner execContextRunner) error {
		for _, st := range statuses {
			nextID++
			record := SchemaRecord{
				ID:        nextID,
				Version:   st.Migration.Version,
				Filename:  st.Migration.Filename,
				Hash:      st.Migration.Hash,
				Status:    StatusBaseline,
				CreatedAt: time.Now().UTC(),
			}
			if _, err := runner.ExecContext(s.ctx, insertSQL, record.ID, record.Version, record.Filename, record.Hash, record.Status, record.CreatedAt); err != nil {
				return fmt.Errorf("baseline %s: %w", record.Filename, err)
			}
		}
		return nil
	})
	return err
}

func (s *Service) Up() error {
	migrations, err := s.loadMigrations(true, false)
	if err != nil {
		return err
	}
	schemas, maxID, err := s.loadSchemas()
	if err != nil {
		return err
	}
	statuses, _ := buildStatuses(migrations, schemas)
	if err := validateStatuses(statuses); err != nil {
		return err
	}

	insertSQL, err := s.cfg.Dialect.InsertSchemaSQL(s.cfg.SchemaName)
	if err != nil {
		return err
	}

	nextID := maxID
	for _, st := range statuses {
		if st.Status == StatusApplied || st.Status == StatusBaseline {
			continue
		}
		if st.Status == StatusOutOfOrder && !s.cfg.MigrateOutOfOrder {
			s.cfg.Logger.Warnf("skip out-of-order migration file=%s version=%s", st.Migration.Filename, st.Migration.Version)
			continue
		}
		if err := s.applyUp(st.Migration, insertSQL, &nextID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) Down(toVersion string, all bool) error {
	if !all && strings.TrimSpace(toVersion) == "" {
		return errors.New("to-version must be set when --all is false")
	}
	migrations, err := s.loadMigrations(false, true)
	if err != nil {
		return err
	}
	schemas, _, err := s.loadSchemas()
	if err != nil {
		return err
	}
	statuses, migrationByVersion := buildStatuses(migrations, schemas)
	if err := validateStatuses(statuses); err != nil {
		return err
	}
	if !all {
		if _, ok := migrationByVersion[toVersion]; !ok {
			return fmt.Errorf("no migration file with version %s found", toVersion)
		}
	}

	deleteSQL, err := s.cfg.Dialect.DeleteSchemaSQL(s.cfg.SchemaName)
	if err != nil {
		return err
	}
	for i := len(statuses) - 1; i >= 0; i-- {
		st := statuses[i]
		if st.Status != StatusApplied {
			continue
		}
		if !all && !CompareVersion(toVersion, st.Migration.Version) {
			continue
		}
		if err := s.applyDown(st.Migration, deleteSQL); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) applyUp(m Migration, insertSQL string, nextID *uint64) error {
	if len(m.UpStatements) == 0 {
		return fmt.Errorf("no up statements found for %s", m.Filename)
	}
	return s.withTx(func(runner execContextRunner) error {
		for _, stmt := range m.UpStatements {
			if s.cfg.DryRun {
				if _, err := io.WriteString(s.cfg.DryRunOutput, stmt+"\n"); err != nil {
					return err
				}
				continue
			}
			if _, err := runner.ExecContext(s.ctx, stmt); err != nil {
				return fmt.Errorf("apply up %s version=%s: %w", m.Filename, m.Version, err)
			}
		}
		if s.cfg.DryRun {
			return nil
		}
		*nextID = *nextID + 1
		record := SchemaRecord{
			ID:        *nextID,
			Version:   m.Version,
			Filename:  m.Filename,
			Hash:      m.Hash,
			Status:    StatusApplied,
			CreatedAt: time.Now().UTC(),
		}
		if _, err := runner.ExecContext(s.ctx, insertSQL, record.ID, record.Version, record.Filename, record.Hash, record.Status, record.CreatedAt); err != nil {
			return fmt.Errorf("insert migration history %s: %w", m.Filename, err)
		}
		return nil
	})
}

func (s *Service) applyDown(m Migration, deleteSQL string) error {
	if len(m.DownStatements) == 0 {
		return fmt.Errorf("no down statements found for %s", m.Filename)
	}
	return s.withTx(func(runner execContextRunner) error {
		for _, stmt := range m.DownStatements {
			if s.cfg.DryRun {
				if _, err := io.WriteString(s.cfg.DryRunOutput, stmt+"\n"); err != nil {
					return err
				}
				continue
			}
			if _, err := runner.ExecContext(s.ctx, stmt); err != nil {
				return fmt.Errorf("apply down %s version=%s: %w", m.Filename, m.Version, err)
			}
		}
		if s.cfg.DryRun {
			return nil
		}
		if _, err := runner.ExecContext(s.ctx, deleteSQL, m.Filename); err != nil {
			return fmt.Errorf("delete migration history %s: %w", m.Filename, err)
		}
		return nil
	})
}

func (s *Service) withTx(fn func(runner execContextRunner) error) (err error) {
	// ClickHouse commonly runs without transaction semantics in database/sql.
	if s.cfg.Dialect != nil && s.cfg.Dialect.Name() == "clickhouse" {
		return fn(s.cfg.DB)
	}
	tx, err := s.cfg.DB.BeginTx(s.ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = errors.Join(err, fmt.Errorf("rollback failed: %w", rbErr))
			}
			return
		}
		if cmErr := tx.Commit(); cmErr != nil {
			err = fmt.Errorf("commit failed: %w", cmErr)
		}
	}()
	err = fn(tx)
	return err
}

func (s *Service) loadMigrations(parseUp, parseDown bool) ([]Migration, error) {
	items, err := s.cfg.MigrationSource.LoadMigrations()
	if err != nil {
		return nil, err
	}
	migrations := make([]Migration, 0, len(items))
	for _, item := range items {
		version, err := SplitFilename(item.Filename)
		if err != nil {
			return nil, err
		}
		m := Migration{
			Filename: item.Filename,
			Source:   item.Source,
			Version:  version,
			Hash:     fmt.Sprintf("%x", md5.Sum([]byte(item.Source))),
		}
		if parseUp {
			stmts, err := parser.SplitSQLStatements(strings.NewReader(item.Source), parser.DirectionUp)
			if err != nil {
				return nil, fmt.Errorf("parse up migration %s: %w", item.Filename, err)
			}
			m.UpStatements = stmts
		}
		if parseDown {
			stmts, err := parser.SplitSQLStatements(strings.NewReader(item.Source), parser.DirectionDown)
			if err != nil {
				return nil, fmt.Errorf("parse down migration %s: %w", item.Filename, err)
			}
			m.DownStatements = stmts
		}
		migrations = append(migrations, m)
	}
	sort.Slice(migrations, func(i, j int) bool {
		return CompareVersion(migrations[i].Version, migrations[j].Version)
	})

	for i := 1; i < len(migrations); i++ {
		if migrations[i-1].Version == migrations[i].Version {
			return nil, fmt.Errorf("duplicate version: %s", migrations[i].Version)
		}
	}
	return migrations, nil
}

func (s *Service) loadSchemas() ([]SchemaRecord, uint64, error) {
	selectSQL, err := s.cfg.Dialect.SelectSchemaSQL(s.cfg.SchemaName)
	if err != nil {
		return nil, 0, err
	}
	rows, err := s.cfg.DB.QueryContext(s.ctx, selectSQL)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	records := make([]SchemaRecord, 0)
	var maxID uint64
	for rows.Next() {
		var rec SchemaRecord
		if err := rows.Scan(&rec.ID, &rec.Version, &rec.Filename, &rec.Hash, &rec.Status, &rec.CreatedAt); err != nil {
			return nil, 0, err
		}
		records = append(records, rec)
		if rec.ID > maxID {
			maxID = rec.ID
		}
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return records, maxID, nil
}

func buildStatuses(migrations []Migration, schemas []SchemaRecord) ([]MigrationStatus, map[string]Migration) {
	versionToSchema := make(map[string]SchemaRecord, len(schemas))
	for _, s := range schemas {
		versionToSchema[s.Version] = s
	}
	latestSchemaVersion := ""
	if len(schemas) > 0 {
		sortedSchemas := append([]SchemaRecord(nil), schemas...)
		sort.Slice(sortedSchemas, func(i, j int) bool {
			return CompareVersion(sortedSchemas[i].Version, sortedSchemas[j].Version)
		})
		latestSchemaVersion = sortedSchemas[len(sortedSchemas)-1].Version
	}

	statuses := make([]MigrationStatus, 0, len(migrations))
	migrationByVersion := make(map[string]Migration, len(migrations))
	for _, migration := range migrations {
		migrationByVersion[migration.Version] = migration
		schema, ok := versionToSchema[migration.Version]
		if !ok {
			st := StatusPending
			if latestSchemaVersion != "" && CompareVersion(migration.Version, latestSchemaVersion) {
				st = StatusOutOfOrder
			}
			statuses = append(statuses, MigrationStatus{Migration: migration, Status: st})
			continue
		}
		if schema.Filename != migration.Filename {
			sc := schema
			statuses = append(statuses, MigrationStatus{Migration: migration, Schema: &sc, Status: StatusFilenameMismatch})
			continue
		}
		if schema.Hash != migration.Hash {
			sc := schema
			statuses = append(statuses, MigrationStatus{Migration: migration, Schema: &sc, Status: StatusHashMismatch})
			continue
		}
		sc := schema
		statuses = append(statuses, MigrationStatus{Migration: migration, Schema: &sc, Status: schema.Status})
	}

	return statuses, migrationByVersion
}

func validateStatuses(statuses []MigrationStatus) error {
	for _, st := range statuses {
		switch st.Status {
		case StatusHashMismatch:
			return fmt.Errorf("filename: %s, version: %s, hash mismatch", st.Migration.Filename, st.Migration.Version)
		case StatusFilenameMismatch:
			return fmt.Errorf("filename: %s, version: %s, filename mismatch", st.Migration.Filename, st.Migration.Version)
		}
	}
	return nil
}
