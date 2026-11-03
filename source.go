package migrate

import (
	"github.com/exc-works/migrate/internal/source"
)

// MigrationSource loads the raw migration files a Service should consider.
type MigrationSource = source.MigrationSource

// SourceFile is one unparsed migration file (filename plus raw source text).
// It is named SourceFile rather than Migration to avoid collision with the parsed migrate.Migration.
type SourceFile = source.Migration

// DirectorySource loads migrations from files on disk under Directory.
type DirectorySource = source.DirectorySource

// StringSource serves a caller-supplied slice of migrations (useful for tests).
type StringSource = source.StringSource

// FSSource loads migrations from an fs.FS (e.g. a //go:embed embed.FS,
// os.DirFS, or fstest.MapFS). Set Root to the subdirectory that holds
// the migration files, or leave empty for the FS root.
type FSSource = source.FSSource

// CombinedSource concatenates multiple MigrationSources into one.
type CombinedSource = source.CombinedSource

// IsSupportedFilename reports whether a filename is a versioned SQL, JS, or TS migration.
func IsSupportedFilename(filename string) bool {
	return source.IsSupportedFilename(filename)
}

// MigrationEntry carries a source file and its optional module context.
type MigrationEntry = source.MigrationEntry

// ModuleContext identifies the disk or virtual files available to compilation.
type ModuleContext = source.ModuleContext

// MigrationEntriesSource optionally supplies entries with module context.
type MigrationEntriesSource = source.MigrationEntriesSource

// WithModules adds module files or a disk dependency directory to a source.
type WithModules = source.WithModules
