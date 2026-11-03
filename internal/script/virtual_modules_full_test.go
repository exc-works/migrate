//go:build full

package script

import (
	"context"
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	migrationsource "github.com/exc-works/migrate/internal/source"
)

//go:embed testdata/modules
var embeddedModules embed.FS

func TestEmbeddedRelativeModules(t *testing.T) {
	entry := "testdata/modules/migrations/V1__embedded.ts"
	text, err := fs.ReadFile(embeddedModules, entry)
	if err != nil {
		t.Fatal(err)
	}
	p, err := Compile(filepath.Base(entry), string(text), "up", Options{}, &migrationsource.ModuleContext{FS: embeddedModules, EntryPath: entry})
	if err != nil {
		t.Fatal(err)
	}
	runModuleProgram(t, p)
}

func TestVirtualEntryCyclesAndStateIsolation(t *testing.T) {
	files := fstest.MapFS{
		"migrations/helper.ts": {Data: []byte(`import {marker} from './V1__virtual';globalThis.initialized=(globalThis.initialized||0)+1;export function read(){return marker}`)},
		"shared/barrel.mjs":    {Data: []byte(`export {read} from '../migrations/helper.ts';`)},
	}
	text := `import {read} from '../shared/barrel.mjs';export const marker=42;export function up(){if(read()!==42||globalThis.initialized!==1)throw Error('virtual cycle or state')}`
	ctx := &migrationsource.ModuleContext{FS: files, EntryPath: "migrations/V1__virtual.ts"}
	p, err := Compile("V1__virtual.ts", text, "up", Options{}, ctx)
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		runModuleProgram(t, p)
	}
}

func TestVirtualModulesCanUseExplicitInstalledPackages(t *testing.T) {
	directory := t.TempDir()
	writeModuleFiles(t, directory, map[string]string{
		"node_modules/helper/package.json": `{"main":"index.js"}`,
		"node_modules/helper/index.js":     `export {value} from './nested.js';`,
		"node_modules/helper/nested.js":    `export const value=42;`,
		"node_modules/pure/package.json":   `{"main":"index.js","sideEffects":false}`,
		"node_modules/pure/index.js":       `throw Error('unused package side effects')`,
	})
	files := fstest.MapFS{
		"shared/index.ts": {Data: []byte(`import 'pure';export {value} from 'helper';`)},
	}
	text := `import {value} from './shared';export function up(){if(value!==42)throw Error('package lookup')}`
	ctx := &migrationsource.ModuleContext{FS: files, DependencyDir: directory}
	p, err := Compile("V1__virtual.ts", text, "up", Options{}, ctx)
	if err != nil {
		t.Fatal(err)
	}
	runModuleProgram(t, p)
	ctx.DependencyDir = ""
	if _, err := Compile("V1__virtual.ts", text, "up", Options{}, ctx); err == nil {
		t.Fatal("virtual package lookup implicitly used disk")
	}
}

func TestVirtualModulesRespectSuppliedFilesystem(t *testing.T) {
	directory := t.TempDir()
	writeModuleFiles(t, directory, map[string]string{"helper.ts": "export const value=42;"})
	t.Chdir(directory)
	ctx := &migrationsource.ModuleContext{FS: fstest.MapFS{}}
	for _, specifier := range []string{"./helper.ts", "../helper.ts", "./missing", "node:fs"} {
		_, err := Compile("V1__virtual.ts", `import {value} from '`+specifier+`';export function up(){return value}`, "up", Options{}, ctx)
		if err == nil || strings.Contains(err.Error(), directory) {
			t.Fatalf("invalid virtual lookup %s: %v", specifier, err)
		}
	}
	if _, err := os.Stat(filepath.Join(directory, "helper.ts")); err != nil {
		t.Fatal(err)
	}
}

func TestImportedVirtualExceptionLocation(t *testing.T) {
	files := fstest.MapFS{"helper.ts": {Data: []byte("export function fail(){\nthrow Error('private data')\n}")}}
	p, err := Compile("V1__entry.ts", `import {fail} from './helper.ts';export function up(){fail()}`, "up", Options{}, &migrationsource.ModuleContext{FS: files})
	if err != nil {
		t.Fatal(err)
	}
	i, err := p.Instantiate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := i.Run(context.Background(), nil); err == nil || !strings.Contains(err.Error(), "helper.ts:2:") || strings.Contains(err.Error(), "private data") {
		t.Fatalf("virtual module location: %v", err)
	}
}

func TestVirtualTypeScriptJavaScriptAliases(t *testing.T) {
	files := fstest.MapFS{
		"helper.ts": {Data: []byte(`import {marker} from './V1__virtual.js';export function read(){return marker}`)},
	}
	text := `import {read} from './helper.js';export const marker=42;export function up(){if(read()!==42)throw Error('TS alias')}`
	p, err := Compile("V1__virtual.ts", text, "up", Options{}, &migrationsource.ModuleContext{FS: files})
	if err != nil {
		t.Fatal(err)
	}
	runModuleProgram(t, p)
}

func TestVirtualDotImportsStayWithinFilesystem(t *testing.T) {
	files := fstest.MapFS{
		"index.js":        {Data: []byte(`export const value=42;`)},
		"nested/index.js": {Data: []byte(`export const value=42;`)},
	}
	disk := t.TempDir()
	writeModuleFiles(t, disk, map[string]string{"index.js": `throw Error('host fallback')`})
	for _, specifier := range []string{".", ".."} {
		p, err := Compile("V1__virtual.js", `import {value} from '`+specifier+`';export function up(){if(value!==42)throw Error('dot import')}`, "up", Options{}, &migrationsource.ModuleContext{FS: files, EntryPath: "nested/V1__virtual.js", DependencyDir: disk})
		if err != nil {
			t.Fatal(err)
		}
		runModuleProgram(t, p)
	}
	if _, err := Compile("V1__virtual.js", `import {value} from '..';export function up(){return value}`, "up", Options{}, &migrationsource.ModuleContext{FS: files, DependencyDir: disk}); err == nil {
		t.Fatal("relative traversal fell back to disk package lookup")
	}
}
