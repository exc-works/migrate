package source

import (
	"errors"
	"io/fs"
	"path"
	"path/filepath"
)

// ModuleContext locates an entry and the files available to its compiler.
// Directory is the disk entry directory; FS and EntryPath describe a virtual entry.
// DependencyDir explicitly enables installed-package lookup for virtual sources.
type ModuleContext struct {
	Directory     string
	FS            fs.FS
	EntryPath     string
	DependencyDir string
}

// MigrationEntry keeps optional module context attached to its original source.
type MigrationEntry struct {
	Migration Migration
	Modules   *ModuleContext
}

// MigrationEntriesSource is an optional extension to MigrationSource.
// A service calls this method instead of LoadMigrations, never both for one load.
type MigrationEntriesSource interface {
	LoadMigrationEntries() ([]MigrationEntry, error)
}

func LoadEntries(src MigrationSource) ([]MigrationEntry, error) {
	if extended, ok := src.(MigrationEntriesSource); ok {
		return extended.LoadMigrationEntries()
	}
	items, err := src.LoadMigrations()
	if err != nil {
		return nil, err
	}
	entries := make([]MigrationEntry, len(items))
	for index, item := range items {
		entries[index].Migration = item
	}
	return entries, nil
}

// WithModules adds module files or an explicit disk resolution directory to a source.
// Existing source structs and their field layouts remain unchanged.
type WithModules struct {
	Source        MigrationSource
	FS            fs.FS
	Root          string
	Directory     string
	DependencyDir string
}

func (s WithModules) LoadMigrations() ([]Migration, error) {
	entries, err := s.LoadMigrationEntries()
	return migrationsFromEntries(entries), err
}

func (s WithModules) LoadMigrationEntries() ([]MigrationEntry, error) {
	if s.Source == nil {
		return nil, errors.New("module source is required")
	}
	if s.FS != nil && s.Directory != "" {
		return nil, errors.New("module source must choose a filesystem or disk directory")
	}
	if s.Root != "" && s.FS == nil {
		return nil, errors.New("module root requires a filesystem")
	}
	if s.FS != nil && s.Root != "" && !fs.ValidPath(s.Root) {
		return nil, errors.New("invalid module filesystem root")
	}
	directory, err := absoluteDirectory(s.Directory)
	if err != nil {
		return nil, err
	}
	dependencyDir, err := absoluteDirectory(s.DependencyDir)
	if err != nil {
		return nil, err
	}
	items, err := LoadEntries(s.Source)
	if err != nil {
		return nil, err
	}
	entries := make([]MigrationEntry, len(items))
	for index, item := range items {
		context := ModuleContext{}
		if item.Modules != nil {
			context = *item.Modules
		}
		if s.FS != nil {
			context.FS = s.FS
			context.Directory = ""
			context.EntryPath = path.Join(s.Root, item.Migration.Filename)
		}
		if directory != "" {
			context.Directory = directory
			context.FS = nil
			context.EntryPath = item.Migration.Filename
		}
		if dependencyDir != "" {
			context.DependencyDir = dependencyDir
		}
		entries[index] = MigrationEntry{
			Migration: item.Migration,
			Modules:   &context,
		}
	}
	return entries, nil
}

func absoluteDirectory(directory string) (string, error) {
	if directory == "" {
		return "", nil
	}
	resolved, err := filepath.Abs(directory)
	if err != nil {
		return "", errors.New("resolve module directory failed")
	}
	return resolved, nil
}

func migrationsFromEntries(entries []MigrationEntry) []Migration {
	items := make([]Migration, len(entries))
	for index, entry := range entries {
		items[index] = entry.Migration
	}
	return items
}
