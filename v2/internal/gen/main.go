// Command gen parses scheme/td_api.tl (the TDLib TL schema) and emits
// generated Go code into the td package: entities, unions, requests and
// update registries. Run it via `go generate ./...` from the v2 module root
// or directly with `go run ./internal/gen`.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func main() {
	_, src, _, ok := runtime.Caller(0)
	if !ok {
		fail("runtime.Caller failed")
	}
	dir := filepath.Dir(src)
	schemePath := filepath.Join(dir, "scheme", "td_api.tl")
	rootDir := facadeRootDir(dir)
	outDir := filepath.Join(rootDir, "td")

	s, err := parseScheme(schemePath)
	if err != nil {
		fail("parse %s: %v", schemePath, err)
	}
	if err := s.validate(); err != nil {
		fail("validate: %v", err)
	}
	if err := emitAll(s, outDir); err != nil {
		fail("emit: %v", err)
	}
	emitted, skipped, err := emitFacade(s, rootDir)
	if err != nil {
		fail("emit facade: %v", err)
	}

	fmt.Printf("gen: %d entities, %d unions, %d requests, %d update variants -> %s\n",
		len(s.entities), len(s.classes), len(s.functions), len(s.updateTypes), outDir)
	fmt.Printf("gen: facade: %d methods -> %s\n", emitted, filepath.Join(rootDir, "facade_gen.go"))
	for _, name := range skipped {
		fmt.Printf("gen: facade: skipped %s (collides with an existing Client member)\n", name)
	}
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "gen: "+format+"\n", args...)
	os.Exit(1)
}
