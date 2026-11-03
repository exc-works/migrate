//go:build full

package script

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	migrationsource "github.com/exc-works/migrate/internal/source"
)

func writeModuleFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for name, text := range files {
		filename := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(filename), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filename, []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func runModuleProgram(t *testing.T, p *Program) *Invocation {
	t.Helper()
	i, err := p.Instantiate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := i.Run(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	return i
}

func TestDiskModulesResolveFromEachImporter(t *testing.T) {
	root := t.TempDir()
	writeModuleFiles(t, root, map[string]string{
		"shared/value.ts":            "import { value } from './nested/value.mjs'; export const answer:number=value;",
		"shared/nested/value.mjs":    "export const value=42;",
		"migrations/shared/index.js": "export const local=8;",
	})
	context := &migrationsource.ModuleContext{Directory: filepath.Join(root, "migrations")}
	for _, extension := range []string{"js", "ts", "mjs"} {
		name := "V1__modules." + extension
		text := `import {answer} from '../shared/value.ts';import {local} from './shared';export function up(){if(answer!==42||local!==8)throw Error('bad imports')}`
		p, err := Compile(name, text, "up", Options{}, context)
		if err != nil {
			t.Fatal(err)
		}
		runModuleProgram(t, p)
	}
}

func TestDiskModulesRetainEntryBytesAndVMIsolation(t *testing.T) {
	root := t.TempDir()
	writeModuleFiles(t, root, map[string]string{
		"V1__entry.js": "throw Error('disk entry must never replace loaded bytes')",
		"helper.js":    `import {marker} from './V1__entry.js';globalThis.initialized=(globalThis.initialized||0)+1;export function read(){return marker}`,
		"barrel.js":    `export {read} from './helper.js';`,
	})
	text := `import {read} from './barrel.js';import {read as again} from './helper.js';export const marker='loaded';export function up(){if(read()!=='loaded'||again()!=='loaded'||globalThis.initialized!==1)throw Error('entry identity or module state')}`
	p, err := Compile("V1__entry.js", text, "up", Options{}, &migrationsource.ModuleContext{Directory: root})
	if err != nil {
		t.Fatal(err)
	}
	writeModuleFiles(t, root, map[string]string{"helper.js": "throw Error('prepared bundle was replaced')"})
	for range 2 {
		runModuleProgram(t, p)
	}
	fresh, err := Compile("V1__entry.js", text, "up", Options{}, &migrationsource.ModuleContext{Directory: root})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = fresh.Instantiate(context.Background()); err == nil {
		t.Fatal("new compilation reused a stale dependency")
	}
}

func TestDiskReexportsValidateOriginalEntrypointKind(t *testing.T) {
	root := t.TempDir()
	for _, entry := range []string{
		`export function up(){}`, `export async function up(){}`, `export function* up(){yield 1}`,
	} {
		writeModuleFiles(t, root, map[string]string{"helper.ts": entry})
		p, err := Compile("V1__entry.ts", `export {up} from './helper.ts'`, "up", Options{}, &migrationsource.ModuleContext{Directory: root})
		if err != nil {
			t.Fatal(err)
		}
		_, err = p.Instantiate(context.Background())
		if (err != nil) != strings.Contains(entry, "async") && !strings.Contains(entry, "function*") {
			t.Fatalf("entry=%s err=%v", entry, err)
		}
		if strings.Contains(entry, "function*") && err == nil {
			t.Fatal("imported generator accepted")
		}
	}
}

func TestImportedModuleDiagnosticsAreSanitized(t *testing.T) {
	root := t.TempDir()
	ctx := &migrationsource.ModuleContext{Directory: root}
	for _, helper := range []string{
		"export function fail(){\nthrow Error('private module data')\n}",
		"export function fail( private module data",
	} {
		writeModuleFiles(t, root, map[string]string{"helper.ts": helper})
		p, err := Compile("V1__entry.ts", `import {fail} from './helper.ts';export function up(){fail()}`, "up", Options{}, ctx)
		if err == nil {
			var i *Invocation
			i, err = p.Instantiate(context.Background())
			if err == nil {
				err = i.Run(context.Background(), nil)
			}
		}
		if err == nil || !strings.Contains(err.Error(), "helper.ts:") || strings.Contains(err.Error(), root) || strings.Contains(err.Error(), "private module data") {
			t.Fatalf("unsafe or missing module diagnostic: %v", err)
		}
	}
}

func TestDiskModuleUnsupportedLoading(t *testing.T) {
	root := t.TempDir()
	writeModuleFiles(t, root, map[string]string{"helper.js": "export const value=1"})
	for _, text := range []string{
		`import x from './missing.js';export function up(){x()}`,
		`import fs from 'node:fs';export function up(){fs.readFileSync('anything')}`,
		`export function up(){import('./helper.js')}`,
		`export function up(){const name='./helper.js';return require(name)}`,
	} {
		if _, err := Compile("V1__entry.js", text, "up", Options{}, &migrationsource.ModuleContext{Directory: root}); err == nil {
			t.Fatalf("accepted unsupported loading: %s", text)
		}
	}
}

func TestEntryAliasesUseSuppliedTypeScriptBytes(t *testing.T) {
	for _, specifier := range []string{"./V1__entry", "./V1__entry.js", "./V1__entry.jsx"} {
		root := t.TempDir()
		writeModuleFiles(t, root, map[string]string{
			"helper.ts": `import {marker} from '` + specifier + `';export function read(){return marker}`,
		})
		text := `import {read} from './helper.ts';export const marker=42;export function up(){if(read()!==42)throw Error('entry overlay')}`
		p, err := Compile("V1__entry.ts", text, "up", Options{}, &migrationsource.ModuleContext{Directory: root})
		if err != nil {
			t.Fatal(err)
		}
		runModuleProgram(t, p)
	}
}

func TestEntryAliasPreservesExactJavaScriptPrecedence(t *testing.T) {
	root := t.TempDir()
	writeModuleFiles(t, root, map[string]string{
		"V1__entry.js": `export const marker=7;`,
		"helper.ts":    `import {marker} from './V1__entry.js';export function read(){return marker}`,
	})
	text := `import {read} from './helper.ts';export const marker=42;export function up(){if(read()!==7)throw Error('exact JS file precedence')}`
	p, err := Compile("V1__entry.ts", text, "up", Options{}, &migrationsource.ModuleContext{Directory: root})
	if err != nil {
		t.Fatal(err)
	}
	runModuleProgram(t, p)
}

func TestDiskDotDirectoryImports(t *testing.T) {
	root := t.TempDir()
	writeModuleFiles(t, root, map[string]string{
		"index.js":        `export const value=42;`,
		"nested/index.js": `export const value=42;`,
	})
	for _, specifier := range []string{".", ".."} {
		p, err := Compile("V1__directory.js", `import {value} from '`+specifier+`';export function up(){if(value!==42)throw Error('directory import')}`, "up", Options{}, &migrationsource.ModuleContext{Directory: filepath.Join(root, "nested")})
		if err != nil {
			t.Fatal(err)
		}
		runModuleProgram(t, p)
	}
}
