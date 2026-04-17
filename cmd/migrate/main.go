package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/tabwriter"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/exc-works/migrate/internal/config"
	"github.com/exc-works/migrate/internal/dialect"
	"github.com/exc-works/migrate/internal/logger"
	"github.com/exc-works/migrate/internal/migrate"
	"github.com/exc-works/migrate/internal/source"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/microsoft/go-mssqldb"
	_ "github.com/sijms/go-ora/v2"
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

var (
	buildVersion = "dev"
	openDB       = sql.Open
	pingDB       = func(ctx context.Context, db *sql.DB) error {
		return db.PingContext(ctx)
	}
	newMigrateService  = migrate.NewService
	pingTimeout        = 5 * time.Second
	safeDescPattern    = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	safeVersionPattern = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
)

func main() {
	root := newRootCommand(os.Stdout, os.Stderr)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(root.ErrOrStderr(), err.Error())
		os.Exit(1)
	}
}

func newRootCommand(stdout, stderr io.Writer) *cobra.Command {
	root := &cobra.Command{
		Use:           "migrate",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.SetOut(stdout)
	root.SetErr(stderr)
	cfgPath := root.PersistentFlags().StringP("config", "c", "migration_config.json", "config file path")
	workingDir := root.PersistentFlags().StringP("working-dir", "w", "", "working directory")

	var svc *migrate.Service
	preRun := func(cmd *cobra.Command, args []string) error {
		c, err := config.Read(*cfgPath, *workingDir)
		if err != nil {
			return err
		}
		s, err := newServiceFromConfig(c)
		if err != nil {
			return err
		}
		svc = s
		return nil
	}

	root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print CLI version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), effectiveBuildVersion())
			return err
		},
	})

	root.AddCommand(&cobra.Command{
		Use:     "create",
		Short:   "Create migration schema table",
		PreRunE: preRun,
		RunE: func(cmd *cobra.Command, args []string) error {
			return svc.Create()
		},
	})

	root.AddCommand(&cobra.Command{
		Use:     "baseline",
		Short:   "Mark all pending migration files as baseline",
		PreRunE: preRun,
		RunE: func(cmd *cobra.Command, args []string) error {
			return svc.Baseline()
		},
	})

	statusCmd := &cobra.Command{
		Use:     "status",
		Short:   "Show migration status",
		Example: "  migrate status\n  migrate status --output json",
		PreRunE: preRun,
		RunE: func(cmd *cobra.Command, args []string) error {
			statuses, err := svc.Status()
			if err != nil {
				return err
			}
			output, err := cmd.Flags().GetString("output")
			if err != nil {
				return err
			}
			return printStatus(cmd.OutOrStdout(), statuses, output)
		},
	}
	statusCmd.Flags().StringP("output", "o", "table", "output format: table|json")
	root.AddCommand(statusCmd)

	upCmd := &cobra.Command{
		Use:     "up",
		Short:   "Apply migrations to latest",
		Example: "  migrate up\n  migrate up --dry-run",
		PreRunE: preRun,
	}
	upDryRun := upCmd.Flags().Bool("dry-run", false, "print SQL statements without executing")
	upCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if *upDryRun {
			svcDry, err := withDryRun(svc)
			if err != nil {
				return err
			}
			return svcDry.Up()
		}
		return svc.Up()
	}
	root.AddCommand(upCmd)

	downCmd := &cobra.Command{
		Use:     "down [to-version]",
		Short:   "Rollback migrations",
		Example: "  migrate down 202604140001\n  migrate down --all\n  migrate down --all --dry-run",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
				return err
			}
			rollbackAll, err := cmd.Flags().GetBool("all")
			if err != nil {
				return err
			}
			if rollbackAll && len(args) > 0 {
				return errors.New("to-version and --all are mutually exclusive")
			}
			if !rollbackAll && len(args) == 0 {
				return errors.New("to-version must be set, or use --all")
			}
			return nil
		},
		PreRunE: preRun,
	}
	downDryRun := downCmd.Flags().Bool("dry-run", false, "print SQL statements without executing")
	downAll := downCmd.Flags().Bool("all", false, "rollback all applied versions")
	downCmd.RunE = func(cmd *cobra.Command, args []string) error {
		var toVersion string
		if len(args) > 0 {
			toVersion = args[0]
		}
		targetSvc := svc
		if *downDryRun {
			drySvc, err := withDryRun(svc)
			if err != nil {
				return err
			}
			targetSvc = drySvc
		}
		return targetSvc.Down(toVersion, *downAll)
	}
	root.AddCommand(downCmd)

	newCmd := &cobra.Command{Use: "new", Short: "Generate migration file or config file"}
	root.AddCommand(newCmd)

	var newVersion string
	newVersionCmd := &cobra.Command{
		Use:     "version [description]",
		Short:   "Create a new migration SQL file",
		Example: "  migrate new version init_users\n  migrate new version add_email -v 202604140002",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := readConfigOrDefault(*cfgPath, *workingDir)
			if err != nil {
				return err
			}
			migrationDir := filepath.Join(c.WorkingDirectory, c.MigrationSource)
			if err := os.MkdirAll(migrationDir, 0o755); err != nil {
				return err
			}
			version := strings.TrimSpace(newVersion)
			if version == "" {
				version = time.Now().UTC().Format("20060102150405")
			}
			version, err = sanitizeVersion(version)
			if err != nil {
				return err
			}
			desc, err := sanitizeDescription(args[0])
			if err != nil {
				return err
			}
			filename := fmt.Sprintf("V%s__%s.sql", version, desc)
			fullPath, err := secureJoinWithin(migrationDir, filename)
			if err != nil {
				return err
			}
			if _, err := os.Stat(fullPath); err == nil {
				return fmt.Errorf("migration file already exists: %s", filename)
			}
			f, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o644)
			if err != nil {
				return err
			}
			defer f.Close()
			_, err = f.WriteString(`-- +migrate Up

-- +migrate Down
`)
			return err
		},
	}
	newVersionCmd.Flags().StringVarP(&newVersion, "version", "v", "", "override migration version")
	newCmd.AddCommand(newVersionCmd)

	newConfigCmd := &cobra.Command{
		Use:   "config [filename]",
		Short: "Create a migration config file",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := "migration_config.json"
			if len(args) == 1 {
				name = args[0]
			}
			template := config.DefaultTemplate()
			content, err := json.MarshalIndent(template, "", "  ")
			if err != nil {
				return err
			}
			flags := os.O_CREATE | os.O_WRONLY | os.O_EXCL
			if force, _ := cmd.Flags().GetBool("force"); force {
				flags = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
			}
			f, err := os.OpenFile(name, flags, 0o644)
			if err != nil {
				if os.IsExist(err) {
					return fmt.Errorf("config file already exists: %s (use --force to overwrite)", name)
				}
				return err
			}
			defer f.Close()
			_, err = f.Write(content)
			return err
		},
	}
	newConfigCmd.Flags().Bool("force", false, "overwrite existing config file")
	newCmd.AddCommand(newConfigCmd)

	return root
}

func newServiceFromConfig(c *config.FileConfig) (*migrate.Service, error) {
	d, err := dialect.FromName(c.Dialect)
	if err != nil {
		return nil, err
	}
	db, err := openDB(d.DriverName(), c.DataSourceName)
	if err != nil {
		return nil, err
	}
	pingCtx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()
	if err := pingDB(pingCtx, db); err != nil {
		_ = db.Close()
		return nil, err
	}
	svc, err := newMigrateService(context.Background(), migrate.Config{
		Dialect:           d,
		DB:                db,
		Logger:            logger.NewStd(c.LoggerLevel, os.Stdout),
		SchemaName:        c.SchemaName,
		MigrateOutOfOrder: c.MigrateOutOfOrder,
		MigrationSource: source.DirectorySource{
			Directory: filepath.Join(c.WorkingDirectory, c.MigrationSource),
		},
		DryRunOutput: os.Stdout,
	})
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	return svc, nil
}

func withDryRun(svc *migrate.Service) (*migrate.Service, error) {
	cfg := svc.Config()
	cfg.DryRun = true
	cfg.DryRunOutput = os.Stdout
	return migrate.NewService(context.Background(), cfg)
}

func readConfigOrDefault(path, workingDir string) (*config.FileConfig, error) {
	cfg, err := config.Read(path, workingDir)
	if err == nil {
		return cfg, nil
	}
	if !os.IsNotExist(err) {
		return nil, err
	}
	fallback := config.DefaultTemplate()
	if workingDir != "" {
		fallback.WorkingDirectory = workingDir
	}
	return &fallback, nil
}

func effectiveBuildVersion() string {
	version := strings.TrimSpace(buildVersion)
	if version == "" {
		return "dev"
	}
	return version
}

func printStatus(out io.Writer, items []migrate.MigrationStatus, output string) error {
	switch strings.ToLower(strings.TrimSpace(output)) {
	case "", "table":
		printStatusTable(out, items)
		return nil
	case "json":
		return printStatusJSON(out, items)
	default:
		return fmt.Errorf("unsupported output format %q; supported values: table,json", output)
	}
}

type statusOutputItem struct {
	Version  string         `json:"version"`
	Filename string         `json:"filename"`
	Hash     string         `json:"hash"`
	Status   migrate.Status `json:"status"`
}

func printStatusJSON(out io.Writer, items []migrate.MigrationStatus) error {
	outputItems := make([]statusOutputItem, 0, len(items))
	for _, item := range items {
		outputItems = append(outputItems, statusOutputItem{
			Version:  item.Migration.Version,
			Filename: item.Migration.Filename,
			Hash:     item.Migration.Hash,
			Status:   item.Status,
		})
	}
	encoder := json.NewEncoder(out)
	return encoder.Encode(outputItems)
}

func printStatusTable(out io.Writer, items []migrate.MigrationStatus) {
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Version\tFilename\tHash\tStatus")
	for _, item := range items {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", item.Migration.Version, item.Migration.Filename, item.Migration.Hash, item.Status)
	}
	_ = w.Flush()
}

func sanitizeDescription(raw string) (string, error) {
	desc := strings.ReplaceAll(strings.TrimSpace(raw), " ", "_")
	if desc == "" {
		return "", errors.New("description must not be empty")
	}
	if strings.Contains(desc, "..") || strings.ContainsAny(desc, `/\`) {
		return "", errors.New("description contains unsafe path characters")
	}
	if !safeDescPattern.MatchString(desc) {
		return "", errors.New("description must match [a-zA-Z0-9_]+")
	}
	return desc, nil
}

func sanitizeVersion(raw string) (string, error) {
	version := strings.TrimSpace(raw)
	if version == "" {
		return "", errors.New("version must not be empty")
	}
	if strings.Contains(version, "..") || strings.ContainsAny(version, `/\`) {
		return "", errors.New("version contains unsafe path characters")
	}
	if !safeVersionPattern.MatchString(version) {
		return "", errors.New("version must match [a-zA-Z0-9_]+")
	}
	return version, nil
}

func secureJoinWithin(baseDir, name string) (string, error) {
	baseAbs, err := filepath.Abs(baseDir)
	if err != nil {
		return "", err
	}
	baseClean := filepath.Clean(baseAbs)
	targetClean := filepath.Clean(filepath.Join(baseClean, name))

	basePrefix := baseClean + string(os.PathSeparator)
	if targetClean != baseClean && !strings.HasPrefix(targetClean, basePrefix) {
		return "", fmt.Errorf("migration file path escapes directory: %s", name)
	}
	return targetClean, nil
}
