package source

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
)

var filenamePattern = regexp.MustCompile(`^V.+__.*\.sql$`)

type Migration struct {
	Filename string
	Source   string
}

type MigrationSource interface {
	LoadMigrations() ([]Migration, error)
}

type StringSource struct {
	Migrations []Migration
}

func (s StringSource) LoadMigrations() ([]Migration, error) {
	return append([]Migration(nil), s.Migrations...), nil
}

type DirectorySource struct {
	Directory string
}

func (d DirectorySource) LoadMigrations() ([]Migration, error) {
	entries, err := os.ReadDir(d.Directory)
	if err != nil {
		return nil, err
	}
	migrations := make([]Migration, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !IsSupportedFilename(name) {
			continue
		}
		path := filepath.Join(d.Directory, name)
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read migration %s: %w", name, err)
		}
		migrations = append(migrations, Migration{Filename: name, Source: string(content)})
	}
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Filename < migrations[j].Filename
	})
	return migrations, nil
}

// FSSource loads migrations from any fs.FS (embed.FS, os.DirFS, fstest.MapFS).
// Root is the directory within FS that holds the migration files; leave empty
// for the FS root. Only direct children are read; subdirectories are ignored.
type FSSource struct {
	FS   fs.FS
	Root string
}

func (f FSSource) LoadMigrations() ([]Migration, error) {
	root := f.Root
	if root == "" {
		root = "."
	}
	entries, err := fs.ReadDir(f.FS, root)
	if err != nil {
		return nil, err
	}
	migrations := make([]Migration, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !IsSupportedFilename(name) {
			continue
		}
		content, err := fs.ReadFile(f.FS, path.Join(root, name))
		if err != nil {
			return nil, fmt.Errorf("read migration %s: %w", name, err)
		}
		migrations = append(migrations, Migration{Filename: name, Source: string(content)})
	}
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Filename < migrations[j].Filename
	})
	return migrations, nil
}

type CombinedSource struct {
	Sources []MigrationSource
}

func (c CombinedSource) LoadMigrations() ([]Migration, error) {
	all := make([]Migration, 0)
	for _, src := range c.Sources {
		items, err := src.LoadMigrations()
		if err != nil {
			return nil, err
		}
		all = append(all, items...)
	}
	return all, nil
}

func IsSupportedFilename(filename string) bool {
	return filenamePattern.MatchString(filename)
}
