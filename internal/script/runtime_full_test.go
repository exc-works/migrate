//go:build full

package script

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestRuntimeContracts(t *testing.T) {
	for _, tc := range []struct {
		name, source, direction, filename string
		badCompile, badInit, badRun       bool
	}{
		{name: "MJS", filename: "V2__test.mjs", source: `export function up(db){}`, direction: "up"},
		{name: "MJS down", filename: "V2__test.mjs", source: `export function down(db){}`, direction: "down"},
		{name: "MJS import", filename: "V2__test.mjs", source: `import x from './helper.js'; export function up(){x()}`, direction: "up", badCompile: true},
		{name: "MJS async", filename: "V2__test.mjs", source: `export async function up(){}`, direction: "up", badInit: true},
		{name: "MJS rejects TS syntax", filename: "V2__test.mjs", source: `export function up(db:unknown){}`, direction: "up", badCompile: true},
		{name: "JS", source: `export function up(db){}`, direction: "up"},
		{name: "TS", source: `type DB={}; export function up(db:DB):void {}`, direction: "up"},
		{name: "down", source: `export function down(db){}`, direction: "down"},
		{name: "missing down", source: `export function up(db){}`, direction: "down", badInit: true},
		{name: "not callable", source: `export const up=3;`, direction: "up", badInit: true},
		{name: "async", source: `export async function up(db){}`, direction: "up", badInit: true},
		{name: "generator", source: `export function* up(db){yield 1}`, direction: "up", badInit: true},
		{name: "promise", source: `export function up(db){return Promise.resolve()}`, direction: "up", badRun: true},
		{name: "rejection", source: `export function up(db){Promise.reject('private data')}`, direction: "up", badRun: true},
		{name: "import", source: `import {x} from './helper';export function up(){x()}`, direction: "up", badCompile: true},
		{name: "dynamic import", source: `export function up(){import('./helper')}`, direction: "up", badCompile: true},
		{name: "reexport", source: `export {up} from './helper'`, direction: "up", badCompile: true},
		{name: "syntax", source: `export function up(`, direction: "up", badCompile: true},
		{name: "throw", source: `export function up(){throw Error('private data')}`, direction: "up", badRun: true},
		{name: "no DB on init", source: `if(typeof db!=='undefined')throw Error('unexpected');export function up(){}`, direction: "up"},
		{name: "no Node", source: `export function up(){if(typeof process!=='undefined'||typeof globalThis.require!=='undefined'||typeof fetch!=='undefined')throw Error('unexpected')}`, direction: "up"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			filename := tc.filename
			if filename == "" {
				filename = "V2__test.ts"
			}
			p, err := Compile(filename, tc.source, tc.direction, Options{})
			if tc.badCompile {
				if err == nil {
					t.Fatal("expected compile failure")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			i, err := p.Instantiate(context.Background())
			if tc.badInit {
				if err == nil {
					t.Fatal("expected initialization failure")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			err = i.Run(context.Background(), nil)
			if (err != nil) != tc.badRun {
				t.Fatalf("%v", err)
			}
			if err != nil && strings.Contains(err.Error(), "private data") {
				t.Fatal("exception data leaked")
			}
		})
	}
}

func TestCompileDoesNotEvaluate(t *testing.T) {
	if _, err := Compile("V1__init.js", `while(true){};export function up(){}`, "up", Options{}); err != nil {
		t.Fatal(err)
	}
}

func TestRuntimeTimeoutAndFreshState(t *testing.T) {
	p, err := Compile("V1__loop.js", `export function up(){try{while(true){}}catch(e){}}`, "up", Options{Timeout: 20 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	i, err := p.Instantiate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err = i.Run(context.Background(), nil); err == nil || !strings.Contains(err.Error(), "deadline exceeded") {
		t.Fatalf("%v", err)
	}
	p, err = Compile("V1__state.js", `globalThis.n=(globalThis.n||0)+1;export function up(){if(globalThis.n!==1)throw Error('state')}`, "up", Options{})
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		i, err = p.Instantiate(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if err = i.Run(context.Background(), nil); err != nil {
			t.Fatal(err)
		}
		if err = i.Run(context.Background(), nil); err == nil {
			t.Fatal("invocation reused")
		}
	}
}

func TestRuntimeSourceMapAndCallbackPanic(t *testing.T) {
	p, err := Compile("V1__location.ts", "type DB = unknown;\nexport function up(db:DB){\nthrow Error('private data');\n}", "up", Options{})
	if err != nil {
		t.Fatal(err)
	}
	i, err := p.Instantiate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err = i.Run(context.Background(), nil); err == nil || !strings.Contains(err.Error(), "V1__location.ts:3:") || strings.Contains(err.Error(), "private data") {
		t.Fatalf("%v", err)
	}
	p, err = Compile("V1__callback.js", `export function up(db){db.fail()}`, "up", Options{})
	if err != nil {
		t.Fatal(err)
	}
	i, err = p.Instantiate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err = i.Run(context.Background(), func(context.Context) any { return map[string]any{"fail": func() { panic("private data") }} }); err == nil || strings.Contains(err.Error(), "private data") {
		t.Fatalf("%v", err)
	}
}

func TestModuleInitializationTimeout(t *testing.T) {
	p, err := Compile("V1__init.js", `while(true){};export function up(){}`, "up", Options{Timeout: 20 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.Instantiate(context.Background()); err == nil || !strings.Contains(err.Error(), "deadline exceeded") {
		t.Fatalf("%v", err)
	}
}

func TestTSIsTranspiledWithoutTypeChecking(t *testing.T) {
	p, err := Compile("V1__types.ts", `export function up(){const x:number='text';if(typeof x!=='string')throw Error('type')}`, "up", Options{})
	if err != nil {
		t.Fatal(err)
	}
	i, err := p.Instantiate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err = i.Run(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
}

func TestAsyncGeneratorRejectedBeforeExecution(t *testing.T) {
	for _, source := range []string{`export async function* down(db){yield 1}`, `export const down=async function*(db){yield 1}`} {
		p, err := Compile("V1__generator.ts", source, "down", Options{})
		if err != nil {
			continue
		}
		if _, err = p.Instantiate(context.Background()); err == nil {
			t.Fatal("async generator passed preflight")
		}
	}
}

func TestScriptTransactionConfig(t *testing.T) {
	for _, tc := range []struct {
		config    string
		want, bad bool
	}{
		{"", true, false}, {`export const config={}`, true, false},
		{`export const config={transaction:true}`, true, false}, {`export const config={transaction:false}`, false, false},
		{`export const config={transaction:undefined}`, true, false},
		{`export const config=null`, false, true}, {`export const config=[]`, false, true},
		{`export const config=false`, false, true}, {`export const config={transaction:'false'}`, false, true},
		{`export const config={transaction:0}`, false, true},
		{`export const config={transaction:{get nested(){throw Error('private value')}}}`, false, true}, {`export const config={transaction:null}`, false, true},
		{`export const config={get transaction(){throw Error('private value')}}`, false, true},
	} {
		for _, direction := range []string{"up", "down"} {
			t.Run(direction+"/"+tc.config, func(t *testing.T) {
				p, err := Compile("V1__config.ts", tc.config+`;export function up(){};export function down(){}`, direction, Options{})
				if err != nil {
					t.Fatal(err)
				}
				i, err := p.Instantiate(context.Background())
				if (err != nil) != tc.bad || err != nil && strings.Contains(err.Error(), "private value") {
					t.Fatalf("%v", err)
				}
				if err == nil && i.TransactionEnabled() != tc.want {
					t.Fatalf("transaction=%v", i.TransactionEnabled())
				}
			})
		}
	}
	p, err := Compile("V1__snapshot.js", `let reads=0;export const config={get transaction(){if(++reads>1)throw Error('read twice');return false}};export function up(){Object.defineProperty(config,'transaction',{value:true})}`, "up", Options{})
	if err != nil {
		t.Fatal(err)
	}
	i, err := p.Instantiate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err = i.Run(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if i.TransactionEnabled() {
		t.Fatal("entrypoint changed preflight transaction policy")
	}
}
