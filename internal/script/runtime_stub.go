//go:build !full

package script

import (
	"context"

	migrationsource "github.com/exc-works/migrate/internal/source"
)

func Formats() []string { return []string{migrationsource.FormatSQL} }

type Program struct{}
type Invocation struct{}

func Compile(string, string, string, Options) (*Program, error)          { return nil, ErrRequiresFull }
func (*Program) Instantiate(context.Context) (*Invocation, error)        { return nil, ErrRequiresFull }
func (*Invocation) Run(context.Context, func(context.Context) any) error { return ErrRequiresFull }
