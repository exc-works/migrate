//go:build full

package script

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
	migrationsource "github.com/exc-works/migrate/internal/source"
)

const entryNamespace = "migration-entry"

// bundleModule resolves dependencies without evaluating JavaScript. The entry always
// comes from the bytes used to calculate its history checksum, including back imports.
func bundleModule(filename, source string, modules *migrationsource.ModuleContext) ([]byte, error) {
	entry, namespace, directory := filename, entryNamespace, ""
	if modules != nil && modules.Directory != "" {
		if modules.FS != nil || !filepath.IsAbs(modules.Directory) {
			return nil, errors.New("invalid disk module context")
		}
		directory = modules.Directory
		entry = filepath.Join(directory, filename)
		if resolved, err := filepath.EvalSymlinks(entry); err == nil {
			entry = resolved
			directory = filepath.Dir(entry)
		}
		namespace = "file"
	}
	result := api.Build(api.BuildOptions{
		EntryPoints:   []string{entry},
		AbsWorkingDir: directory,
		Bundle:        true,
		Write:         false,
		Format:        api.FormatIIFE,
		GlobalName:    "migration",
		Platform:      api.PlatformNeutral,
		MainFields:    []string{"module", "main"},
		Conditions:    []string{},
		Target:        api.ES2017,
		// Preserve async generators so initialization rejects their original kind.
		Supported:      map[string]bool{"async-generator": true},
		Sourcemap:      api.SourceMapInline,
		SourcesContent: api.SourcesContentExclude,
		TsconfigRaw:    "{}",
		LogLevel:       api.LogLevelSilent,
		LogOverride: map[string]api.LogLevel{
			"unsupported-dynamic-import": api.LogLevelError,
			"unsupported-require-call":   api.LogLevelError,
		},
		Plugins: []api.Plugin{{
			Name: "migration-modules",
			Setup: func(build api.PluginBuild) {
				build.OnResolve(api.OnResolveOptions{Filter: ".*"}, func(args api.OnResolveArgs) (api.OnResolveResult, error) {
					if args.Kind == api.ResolveEntryPoint {
						return api.OnResolveResult{
							Path:      entry,
							Namespace: namespace,
						}, nil
					}
					if args.Kind == api.ResolveJSDynamicImport || args.Kind == api.ResolveJSRequireResolve {
						return moduleError("asynchronous or runtime module loading is unsupported"), nil
					}
					if directory == "" {
						return moduleError("module import requires source context"), nil
					}
					if strings.HasPrefix(args.Path, "node:") {
						return moduleError("Node built-in modules are unavailable in Goja"), nil
					}
					// Handle back imports even when a supplied string entry has no disk file.
					if (strings.HasPrefix(args.Path, "./") || strings.HasPrefix(args.Path, "../") || filepath.IsAbs(args.Path)) && diskEntryAlias(filepath.Clean(filepath.Join(args.ResolveDir, args.Path)), entry) {
						return api.OnResolveResult{
							Path:      entry,
							Namespace: namespace,
						}, nil
					}
					return api.OnResolveResult{}, nil
				})
				build.OnLoad(api.OnLoadOptions{Filter: ".*"}, func(args api.OnLoadArgs) (api.OnLoadResult, error) {
					text, loaderName := source, filename
					if args.Path != entry || args.Namespace != namespace {
						content, err := os.ReadFile(args.Path)
						if err != nil {
							return api.OnLoadResult{Errors: []api.Message{{Text: "read module failed"}}}, nil
						}
						text, loaderName = string(content), args.Path
					}
					loader, err := moduleLoader(loaderName)
					if err != nil {
						return api.OnLoadResult{Errors: []api.Message{{Text: err.Error()}}}, nil
					}
					resolveDir := ""
					if args.Namespace == "file" {
						resolveDir = filepath.Dir(args.Path)
					}
					return api.OnLoadResult{
						Contents:   &text,
						Loader:     loader,
						ResolveDir: resolveDir,
					}, nil
				})
			},
		}},
	})
	if len(result.Errors) > 0 {
		if loc := result.Errors[0].Location; loc != nil {
			return nil, fmt.Errorf("compile script %s: module %s:%d:%d: module resolution or compilation failed", filename, moduleDisplayName(loc.File), loc.Line, loc.Column+1)
		}
		return nil, fmt.Errorf("compile script %s: module resolution or compilation failed", filename)
	}
	if len(result.OutputFiles) != 1 {
		return nil, fmt.Errorf("compile script %s: expected one module", filename)
	}
	return result.OutputFiles[0].Contents, nil
}

func moduleError(message string) api.OnResolveResult {
	return api.OnResolveResult{Errors: []api.Message{{Text: message}}}
}

func moduleLoader(filename string) (api.Loader, error) {
	switch filepath.Ext(filename) {
	case migrationsource.ExtensionJS, migrationsource.ExtensionMJS, ".cjs":
		return api.LoaderJS, nil
	case migrationsource.ExtensionTS:
		return api.LoaderTS, nil
	case ".json":
		return api.LoaderJSON, nil
	default:
		return api.LoaderNone, errors.New("unsupported module file type")
	}
}

func moduleDisplayName(filename string) string {
	return strings.TrimPrefix(filepath.Base(strings.ReplaceAll(filename, "\\", "/")), entryNamespace+":")
}

// Overlay the loaded entry at the position where the native extension search would
// find it. Existing higher-priority files still resolve through esbuild normally.
func diskEntryAlias(candidate, entry string) bool {
	for _, base := range []string{candidate, filepath.Join(candidate, "index")} {
		for _, extension := range []string{"", ".tsx", ".ts", ".jsx", ".js", ".css", ".json"} {
			name := base + extension
			if name == entry {
				return true
			}
			if info, err := os.Stat(name); err == nil && !info.IsDir() {
				return false
			}
		}
	}
	return false
}
