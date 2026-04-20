package migrate

import (
	"github.com/exc-works/migrate/internal/source"
)

// MigrationSource loads the raw migration files a Service should consider.
type MigrationSource = source.MigrationSource

// SourceFile is one unparsed migration file (filename plus raw SQL text).
// It is named SourceFile rather than Migration to avoid collision with the parsed migrate.Migration.
type SourceFile = source.Migration

// DirectorySource loads migrations from files on disk under Directory.
type DirectorySource = source.DirectorySource

// StringSource serves a caller-supplied slice of migrations (useful with //go:embed or tests).
type StringSource = source.StringSource

// CombinedSource concatenates multiple MigrationSources into one.
type CombinedSource = source.CombinedSource

// IsSupportedFilename reports whether a filename matches the V<version>__<desc>.sql pattern.
func IsSupportedFilename(filename string) bool {
	return source.IsSupportedFilename(filename)
}
