package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/exc-works/migrate/internal/batch"
	"github.com/exc-works/migrate/internal/config"
	"github.com/exc-works/migrate/internal/migrate"
	"github.com/exc-works/migrate/internal/parser"
	"github.com/exc-works/migrate/internal/tunnel"
	"github.com/spf13/cobra"
)

const defaultParallel = 8

type targetOutput struct {
	plan       bytes.Buffer
	migrations []migrate.MigrationStatus
}

type batchOutputItem struct {
	DataSource string              `json:"data_source"`
	Status     batch.Status        `json:"status"`
	Error      string              `json:"error,omitempty"`
	Migrations *[]statusOutputItem `json:"migrations,omitempty"`
}

type commandOperation struct {
	name      string
	dryRun    bool
	all       bool
	toVersion string
}

func runMultiCommand(cmd *cobra.Command, args []string, cfg *config.FileConfig, parallel int, noninteractive bool) error {
	op := commandOperation{
		name: cmd.Name(),
	}
	if cmd.Flags().Lookup("dry-run") != nil {
		op.dryRun, _ = cmd.Flags().GetBool("dry-run")
	}
	if cmd.Flags().Lookup("all") != nil {
		op.all, _ = cmd.Flags().GetBool("all")
	}
	if len(args) > 0 {
		op.toVersion = args[0]
	}
	output := "table"
	if op.name == commandStatus {
		output, _ = cmd.Flags().GetString("output")
		output = strings.ToLower(strings.TrimSpace(output))
		if output != "json" && output != "table" && output != "" {
			return errors.New("unsupported output format; supported values: table,json")
		}
	}
	stderr := &synchronizedWriter{
		writer: cmd.ErrOrStderr(),
	}
	promptGate := make(chan struct{}, 1)
	configs := make(map[string]config.FileConfig)
	payloads := make(map[string]*targetOutput)
	var names []string
	for _, target := range cfg.Targets() {
		names = append(names, target.Name)
		configs[target.Name] = target.Config
		payloads[target.Name] = &targetOutput{}
	}
	results, err := batch.Run(cmd.Context(), names, parallel, func(ctx context.Context, name string) (err error) {
		c := configs[name]
		payload := payloads[name]
		logs := &targetLogWriter{
			name:   name,
			writer: stderr,
		}
		var prompt tunnel.Prompt
		var hostKeyPrompt tunnel.HostKeyPrompt
		if !noninteractive {
			prompt = func(ctx context.Context, label string) ([]byte, error) {
				if err := acquirePrompt(ctx, promptGate); err != nil {
					return nil, err
				}
				defer func() { <-promptGate }()
				return terminalSecret(ctx, stderr, fmt.Sprintf("[%q] %s", name, label))
			}
			hostKeyPrompt = func(ctx context.Context, info tunnel.HostKeyInfo) (bool, error) {
				if err := acquirePrompt(ctx, promptGate); err != nil {
					return false, err
				}
				defer func() { <-promptGate }()
				return terminalHostKey(ctx, stderr, info)
			}
		}
		svc, cleanup, err := newCommandService(ctx, &c, prompt, hostKeyPrompt, logs, &payload.plan)
		if err != nil {
			return err
		}
		defer func() { err = errors.Join(err, cleanup()) }()
		if op.dryRun {
			svc, err = withDryRun(svc, svc.Context(), &payload.plan)
			if err != nil {
				return err
			}
		}
		payload.migrations, err = executeOperation(svc, op)
		return err
	})
	if err != nil {
		return err
	}
	return printBatchResults(cmd.OutOrStdout(), results, payloads, op, output)
}

func executeOperation(svc *migrate.Service, op commandOperation) ([]migrate.MigrationStatus, error) {
	switch op.name {
	case commandCreate:
		return nil, svc.Create()
	case commandBaseline:
		return nil, svc.Baseline()
	case commandStatus:
		return svc.Status()
	case parser.DirectionUpName:
		return nil, svc.Up()
	case parser.DirectionDownName:
		return nil, svc.Down(op.toVersion, op.all)
	default:
		return nil, errors.New("unsupported database command")
	}
}

func printBatchResults(out io.Writer, results []batch.Result, payloads map[string]*targetOutput, op commandOperation, output string) error {
	items := make([]batchOutputItem, 0, len(results))
	var overall error
	for _, result := range results {
		item := batchOutputItem{
			DataSource: result.Name,
			Status:     result.Status,
		}
		if result.Err != nil {
			overall = errors.New("one or more data sources failed or were canceled")
			item.Error = "operation failed; inspect this database and migration history before retrying"
			if result.Status == batch.StatusCanceled {
				item.Error = "operation canceled; inspect this database and migration history before retrying"
			}
		}
		if op.name == commandStatus && result.Status == batch.StatusSucceeded {
			statuses := statusOutputItems(payloads[result.Name].migrations)
			item.Migrations = &statuses
		}
		items = append(items, item)
	}
	if output == "json" {
		return errors.Join(overall, json.NewEncoder(out).Encode(items))
	}
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, "Data source\tStatus\tError"); err != nil {
		return errors.Join(overall, err)
	}
	for _, item := range items {
		if _, err := fmt.Fprintf(w, "%q\t%s\t%s\n", item.DataSource, item.Status, item.Error); err != nil {
			return errors.Join(overall, err)
		}
	}
	if err := w.Flush(); err != nil {
		return errors.Join(overall, err)
	}
	for _, item := range items {
		payload := payloads[item.DataSource]
		if op.name != commandStatus && !op.dryRun {
			continue
		}
		if _, err := fmt.Fprintf(out, "\nData source %q\n", item.DataSource); err != nil {
			return errors.Join(overall, err)
		}
		if op.name == commandStatus && item.Status == batch.StatusSucceeded {
			if err := printStatus(out, payload.migrations, "table"); err != nil {
				return errors.Join(overall, err)
			}
		} else if _, err := out.Write(payload.plan.Bytes()); err != nil {
			return errors.Join(overall, err)
		}
	}
	return overall
}

func acquirePrompt(ctx context.Context, gate chan struct{}) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case gate <- struct{}{}:
		if err := ctx.Err(); err != nil {
			<-gate
			return err
		}
		return nil
	}
}

type synchronizedWriter struct {
	mu     sync.Mutex
	writer io.Writer
}

func (w *synchronizedWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.writer.Write(p)
}

type targetLogWriter struct {
	name   string
	writer io.Writer
}

func (w *targetLogWriter) Write(p []byte) (int, error) {
	_, err := fmt.Fprintf(w.writer, "[data_source=%q] %s", w.name, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}
