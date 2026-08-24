//go:build !tdjson

package main

import "log"

// The real bootstrap lives in main.go behind the tdjson tag; without the
// tag there is no TDLib transport to wire the server to.
func main() {
	log.Fatal("tdlib-mcp requires the tdjson build tag: go build -tags tdjson ./cmd/tdlib-mcp")
}
