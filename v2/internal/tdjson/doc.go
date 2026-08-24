// Package tdjson is the isolated cgo edge to libtdjson (the TDLib JSON
// interface). It is the only package in the module that requires libtdjson
// to be installed.
//
// The cgo implementation lives behind the build constraint
//
//	//go:build cgo && tdjson
//
// so a plain `go build ./...` / `go test ./...` succeeds on machines without
// libtdjson: untagged, this package is empty and the rest of the module
// (transport fakes, tests) never links against C. On a machine with
// libtdjson installed under /usr/local (headers in
// /usr/local/include/td/telegram, library in /usr/local/lib), build with:
//
//	go build -tags tdjson ./...
package tdjson
