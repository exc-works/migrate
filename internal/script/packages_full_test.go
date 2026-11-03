//go:build full

package script

import (
	"os"
	"path/filepath"
	"testing"

	migrationsource "github.com/exc-works/migrate/internal/source"
)

func TestInstalledPackagesUseNativeEntryResolution(t *testing.T) {
	root := t.TempDir()
	writeModuleFiles(t, root, map[string]string{
		"node_modules/plain/package.json":          `{"main":"common.cjs","module":"module.js"}`,
		"node_modules/plain/module.js":             `export const value='module';`,
		"node_modules/plain/common.cjs":            `exports.value='main';`,
		"node_modules/@scope/helpers/package.json": `{"exports":{".":{"browser":"./bad.js","node":"./bad.js","import":"./esm.js","require":"./common.cjs","default":"./bad.js"},"./sub":"./sub.ts"}}`,
		"node_modules/@scope/helpers/esm.js":       `export const value='import';`,
		"node_modules/@scope/helpers/common.cjs":   `exports.value='require';`,
		"node_modules/@scope/helpers/sub.ts":       `export const sub:number=42;`,
		"node_modules/@scope/helpers/bad.js":       `throw Error('wrong condition')`,
		"node_modules/common/package.json":         `{"main":"index.cjs"}`,
		"node_modules/common/index.cjs":            `const helper=require('@scope/helpers');exports.value=helper.value;`,
		"node_modules/default/package.json":        `{"exports":{"module":"./bad.js","default":"./index.js"}}`,
		"node_modules/default/index.js":            `export const value='default';`,
		"node_modules/default/bad.js":              `throw Error('implicit module condition')`,
		"package-lock.json":                        "not parsed by migrate",
	})
	text := `import {value as plain} from 'plain';import {value as esm} from '@scope/helpers';import {sub} from '@scope/helpers/sub';import cjs from 'common';import {value as fallback} from 'default';export function up(){if(plain!=='module'||esm!=='import'||sub!==42||cjs.value!=='require'||fallback!=='default')throw Error('incorrect package entry')}`
	p, err := Compile("V1__packages.ts", text, "up", Options{}, &migrationsource.ModuleContext{Directory: root})
	if err != nil {
		t.Fatal(err)
	}
	runModuleProgram(t, p)
}

func TestPackageVersionsFollowEachImporter(t *testing.T) {
	for _, version := range []string{"first", "second"} {
		root := t.TempDir()
		writeModuleFiles(t, root, map[string]string{
			"node_modules/parent/package.json":                     `{"main":"index.js"}`,
			"node_modules/parent/index.js":                         `export {value} from 'nested';`,
			"node_modules/parent/node_modules/nested/package.json": `{"main":"index.js"}`,
			"node_modules/parent/node_modules/nested/index.js":     `export const value='` + version + `';`,
			"node_modules/nested/package.json":                     `{"main":"index.js"}`,
			"node_modules/nested/index.js":                         `throw Error('wrong dependency version')`,
		})
		p, err := Compile("V1__version.js", `import {value} from 'parent';export function up(){if(value!=='`+version+`')throw Error('wrong version')}`, "up", Options{}, &migrationsource.ModuleContext{Directory: root})
		if err != nil {
			t.Fatal(err)
		}
		runModuleProgram(t, p)
	}
}

func TestLinkedInstalledPackage(t *testing.T) {
	root := t.TempDir()
	writeModuleFiles(t, root, map[string]string{
		"node_modules/.pnpm/linked@1/node_modules/linked/package.json": `{"main":"index.js"}`,
		"node_modules/.pnpm/linked@1/node_modules/linked/index.js":     `export const value=42;`,
	})
	if err := os.Symlink(filepath.FromSlash(".pnpm/linked@1/node_modules/linked"), filepath.Join(root, "node_modules", "linked")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	p, err := Compile("V1__linked.mjs", `import {value} from 'linked';export function up(){if(value!==42)throw Error('linked package')}`, "up", Options{}, &migrationsource.ModuleContext{Directory: root})
	if err != nil {
		t.Fatal(err)
	}
	runModuleProgram(t, p)
}

func TestPackageExportsAndRuntimeBoundaries(t *testing.T) {
	root := t.TempDir()
	writeModuleFiles(t, root, map[string]string{
		"node_modules/node-only/package.json": `{"exports":{"node":"./index.js"}}`,
		"node_modules/node-only/index.js":     `export const value=1;`,
		"node_modules/private/package.json":   `{"exports":{".":"./index.js"}}`,
		"node_modules/private/index.js":       `export const value=1;`,
		"node_modules/private/hidden.js":      `export const value=1;`,
		"node_modules/native/package.json":    `{"main":"addon.node"}`,
		"node_modules/native/addon.node":      "native fixture",
		"node_modules/builtin/package.json":   `{"main":"index.js"}`,
		"node_modules/builtin/index.js":       `import fs from 'node:fs';export const value=fs.readFileSync('file');`,
	})
	for _, specifier := range []string{"node-only", "private/hidden.js", "native", "builtin", "missing-package"} {
		if _, err := Compile("V1__unsupported.js", `import {value} from '`+specifier+`';export function up(){return value}`, "up", Options{}, &migrationsource.ModuleContext{Directory: root}); err == nil {
			t.Fatalf("unsupported package resolved: %s", specifier)
		}
	}
}
