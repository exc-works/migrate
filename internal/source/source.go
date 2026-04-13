package source

import (
	"fmt"
	"os"
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
