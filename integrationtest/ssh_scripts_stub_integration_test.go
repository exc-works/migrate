//go:build integration && !full

package integrationtest

import "github.com/exc-works/migrate/internal/source"

func sshScriptMigrations(string) []source.Migration { return nil }
