package script

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/dop251/goja"
	"github.com/dop251/goja/parser"
	migrationparser "github.com/exc-works/migrate/internal/parser"
	migrationsource "github.com/exc-works/migrate/internal/source"
)

func Formats() []string {
	return []string{migrationsource.FormatSQL, migrationsource.FormatJS, migrationsource.FormatTS, migrationsource.FormatMJS}
}

type Program struct {
	code                *goja.Program
	filename, direction string
	options             Options
}

type Invocation struct {
	vm          *goja.Runtime
	fn          goja.Callable
	program     *Program
	rejected    map[*goja.Promise]bool
	used        bool
	transaction bool
}

// Compile bundles and compiles without evaluating scripts or dependency modules.
func Compile(filename, source, direction string, options Options, contexts ...*migrationsource.ModuleContext) (*Program, error) {
	options, err := options.Normalize()
	if err != nil {
		return nil, err
	}
	if filepath.Base(filename) != filename || strings.ContainsAny(filename, "/\\") {
		return nil, errors.New("script filename must be a base filename")
	}
	if direction != migrationparser.DirectionUpName && direction != migrationparser.DirectionDownName {
		return nil, errors.New("invalid script direction")
	}
	if len(contexts) > 1 {
		return nil, errors.New("expected at most one module context")
	}
	var modules *migrationsource.ModuleContext
	if len(contexts) == 1 {
		modules = contexts[0]
	}
	switch filepath.Ext(filename) {
	case migrationsource.ExtensionJS, migrationsource.ExtensionMJS, migrationsource.ExtensionTS:
	default:
		return nil, errors.New("unsupported script extension")
	}
	bundled, err := bundleModule(filename, source, modules)
	if err != nil {
		return nil, err
	}
	ast, err := parser.ParseFile(nil, filename, string(bundled), 0, parser.WithSourceMapLoader(func(string) ([]byte, error) { return nil, errors.New("external source maps disabled") }))
	if err != nil {
		return nil, fmt.Errorf("compile script %s: invalid source map", filename)
	}
	code, err := goja.CompileAST(ast, false)
	if err != nil {
		return nil, fmt.Errorf("compile script %s: unsupported JavaScript syntax", filename)
	}
	return &Program{
		code:      code,
		filename:  filename,
		direction: direction,
		options:   options,
	}, nil
}

// Instantiate evaluates the module without a database, then validates the entrypoint.
// A plan retains each instance so initialization happens exactly once.
func (p *Program) Instantiate(ctx context.Context) (instance *Invocation, err error) {
	i := &Invocation{
		vm:          goja.New(),
		program:     p,
		rejected:    make(map[*goja.Promise]bool),
		transaction: true,
	}
	i.vm.SetMaxCallStackSize(512)
	i.vm.SetParserOptions(parser.WithSourceMapLoader(func(string) ([]byte, error) { return nil, errors.New("external source maps disabled") }))
	i.vm.SetPromiseRejectionTracker(func(p *goja.Promise, op goja.PromiseRejectionOperation) {
		if op == goja.PromiseRejectionReject {
			i.rejected[p] = true
		} else {
			delete(i.rejected, p)
		}
	})
	err = i.guarded(ctx, func(context.Context) error {
		if _, e := i.vm.RunProgram(p.code); e != nil {
			return e
		}
		var validateErr error
		if e := i.vm.Try(func() {
			exports := i.vm.Get("migration")
			if exports == nil || goja.IsUndefined(exports) || goja.IsNull(exports) {
				validateErr = errors.New("missing exports")
				return
			}
			module := exports.ToObject(i.vm)
			config := module.Get("config")
			if config != nil && !goja.IsUndefined(config) {
				object, ok := config.(*goja.Object)
				if !ok || object.ClassName() != "Object" {
					validateErr = errors.New("config must be an object")
					return
				}
				value := object.Get("transaction")
				if value != nil && !goja.IsUndefined(value) {
					if _, objectValue := value.(*goja.Object); objectValue {
						validateErr = errors.New("config.transaction must be a boolean")
						return
					}
					var valid bool
					i.transaction, valid = value.Export().(bool)
					if !valid {
						validateErr = errors.New("config.transaction must be a boolean")
						return
					}
				}
			}
			entry := module.Get(p.direction)
			var ok bool
			i.fn, ok = goja.AssertFunction(entry)
			if !ok {
				validateErr = errors.New("missing synchronous entrypoint")
				return
			}
			constructor := entry.ToObject(i.vm).Get("constructor")
			if constructor == nil || goja.IsUndefined(constructor) || constructor.ToObject(i.vm).Get("name").String() != "Function" {
				validateErr = errors.New("entrypoint must be synchronous")
			}
		}); e != nil {
			return e
		}
		if validateErr != nil {
			return validateErr
		}
		if len(i.rejected) > 0 {
			return errors.New("unhandled promise rejection")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return i, nil
}

// TransactionEnabled snapshots the module's config.transaction before execution.
// The default requests the framework transaction where the dialect supports it.
func (i *Invocation) TransactionEnabled() bool { return i.transaction }

// Run is single-use. Bindings are created with the same deadline as the script.
func (i *Invocation) Run(ctx context.Context, bindings func(context.Context) any) error {
	if i.used {
		return errors.New("script invocation is single-use")
	}
	i.used = true
	return i.guarded(ctx, func(runCtx context.Context) error {
		argument := goja.Undefined()
		if bindings != nil {
			argument = i.vm.ToValue(bindings(runCtx))
		}
		result, err := i.fn(goja.Undefined(), argument)
		if err != nil {
			return err
		}
		// Ignore synchronous return values without exporting them or reading
		// properties. Promise results still require an unsupported async runtime.
		if result != nil && result.ExportType() == reflect.TypeFor[*goja.Promise]() {
			return errors.New("entrypoint must be synchronous")
		}
		if len(i.rejected) > 0 {
			return errors.New("unhandled promise rejection")
		}
		return nil
	})
}

func (i *Invocation) guarded(parent context.Context, fn func(context.Context) error) (err error) {
	ctx, cancel := context.WithTimeout(parent, i.program.options.Timeout)
	done := make(chan struct{})
	stop := context.AfterFunc(ctx, func() { i.vm.Interrupt(ctx.Err()); close(done) })
	defer func() {
		if !stop() {
			<-done
		}
		cancel()
		i.vm.ClearInterrupt()
		if recover() != nil {
			err = fmt.Errorf("execute %s script %s: host callback failed", i.program.direction, i.program.filename)
		}
	}()
	if ctx.Err() != nil {
		return fmt.Errorf("execute %s script %s: %w", i.program.direction, i.program.filename, ctx.Err())
	}
	err = fn(ctx)
	if ctx.Err() != nil {
		return fmt.Errorf("execute %s script %s: %w", i.program.direction, i.program.filename, ctx.Err())
	}
	if err != nil {
		var exception *goja.Exception
		if errors.As(err, &exception) {
			for _, frame := range exception.Stack() {
				pos := frame.Position()
				if pos.Line > 0 {
					module := moduleDisplayName(pos.Filename)
					if module != i.program.filename {
						return fmt.Errorf("execute %s script %s: module %s:%d:%d: script failed", i.program.direction, i.program.filename, module, pos.Line, pos.Column)
					}
					return fmt.Errorf("execute %s script %s:%d:%d: script failed", i.program.direction, i.program.filename, pos.Line, pos.Column)
				}
			}
		}
		return fmt.Errorf("execute %s script %s: invalid synchronous entrypoint or script failure", i.program.direction, i.program.filename)
	}
	return nil
}
