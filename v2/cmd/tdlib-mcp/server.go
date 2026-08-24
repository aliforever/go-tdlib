// Command tdlib-mcp is an MCP (Model Context Protocol) server over stdio
// that exposes a live go-tdlib v2 session to AI agents: TDLib method
// search/inspection, raw method calls, update streaming and login-code
// submission.
//
// This file contains the protocol machinery — JSON-RPC framing, tool
// dispatch, update buffering and call validation. It is deliberately
// free of any go-tdlib imports (stdlib only) so the protocol logic is
// unit-testable without the tdjson build tag; main.go (tagged tdjson)
// wires it to a real TDLib client through the small function types below.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

// maxLineSize bounds one JSON-RPC frame; TDLib updates can be large.
const maxLineSize = 16 << 20

func newLineScanner(r io.Reader) *bufio.Scanner {
	s := bufio.NewScanner(r)
	s.Buffer(make([]byte, 64<<10), maxLineSize)
	return s
}

// ---------------------------------------------------------------------------
// Backend contract
// ---------------------------------------------------------------------------

// ParamSchema mirrors td.ParamSchema. It is redeclared here so this file
// (and its tests) need no go-tdlib imports; main.go adapts the generated
// registry to these types.
type ParamSchema struct {
	Name        string   `json:"name"`
	GoType      string   `json:"goType"`
	IsUnion     bool     `json:"isUnion,omitempty"`
	Variants    []string `json:"variants,omitempty"`
	Description string   `json:"description,omitempty"`
}

// MethodSchema mirrors td.MethodSchema; see ParamSchema.
type MethodSchema struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Params      []ParamSchema `json:"params"`
	ReturnType  string        `json:"returnType"`
}

// LookupFunc resolves a TDLib method name to its schema.
type LookupFunc func(name string) (MethodSchema, bool)

// SearchFunc returns up to limit schemas matching query.
type SearchFunc func(query string, limit int) []MethodSchema

// CallFunc invokes a TDLib method with raw JSON-able params and returns the
// raw response. TDLib errors must be returned as *CallError so the server
// can surface {code, message} to the caller.
type CallFunc func(ctx context.Context, method string, params map[string]any) (json.RawMessage, error)

// CallError is a TDLib error transported across the backend boundary.
type CallError struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

func (e *CallError) Error() string { return fmt.Sprintf("tdlib error %d: %s", e.Code, e.Message) }

// Config wires a Server to its backend. Lookup, Search and Call are
// required; SubmitCode may be nil (the tool then reports it unsupported).
type Config struct {
	Lookup     LookupFunc
	Search     SearchFunc
	Call       CallFunc
	SubmitCode func(code string) error

	// CallTimeout bounds one tdlib_call; defaults to 60s.
	CallTimeout time.Duration
	// UpdateBufferCap bounds the update ring buffer; defaults to 500.
	UpdateBufferCap int
}

// ---------------------------------------------------------------------------
// JSON-RPC framing
// ---------------------------------------------------------------------------

const (
	protocolVersion = "2024-11-05"
	serverName      = "tdlib-mcp"
	serverVersion   = "0.1.0"
)

const (
	errInvalidRequest = -32600
	errMethodNotFound = -32601
	errInvalidParams  = -32602
)

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// rpcResponse carries Result pre-marshalled so that "result" can be omitted
// entirely on error responses (and vice versa).
type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

// Server is the MCP protocol endpoint. It is safe for concurrent use.
type Server struct {
	cfg Config

	outMu sync.Mutex
	out   io.Writer

	updates *updateBuffer
}

// NewServer builds a Server, applying Config defaults.
func NewServer(cfg Config) *Server {
	if cfg.CallTimeout <= 0 {
		cfg.CallTimeout = 60 * time.Second
	}
	if cfg.UpdateBufferCap <= 0 {
		cfg.UpdateBufferCap = 500
	}
	return &Server{cfg: cfg, updates: newUpdateBuffer(cfg.UpdateBufferCap)}
}

// HandleUpdate feeds one incoming TDLib update into the ring buffer and any
// matching waiters. It never blocks; raw is copied. Register it as the
// client's single OnRaw hook.
func (s *Server) HandleUpdate(tdType string, raw json.RawMessage) {
	s.updates.add(tdType, raw)
}

// Serve runs the newline-delimited JSON-RPC loop until in hits EOF, ctx is
// done or a read error occurs. In-flight tool calls are awaited before
// returning. Logs must never go to out; out carries protocol frames only.
func (s *Server) Serve(ctx context.Context, in io.Reader, out io.Writer) error {
	s.outMu.Lock()
	s.out = out
	s.outMu.Unlock()

	scanner := newLineScanner(in)
	var wg sync.WaitGroup
	for scanner.Scan() {
		line := append([]byte(nil), scanner.Bytes()...)
		if len(strings.TrimSpace(string(line))) == 0 {
			continue
		}
		s.dispatch(ctx, line, &wg)
	}
	wg.Wait()
	return scanner.Err()
}

// dispatch parses one frame. Notifications (no id) are dropped silently —
// the protocol forbids responding to them. Requests are handled on their
// own goroutine; responses are written atomically under outMu.
func (s *Server) dispatch(ctx context.Context, line []byte, wg *sync.WaitGroup) {
	var req rpcRequest
	if err := json.Unmarshal(line, &req); err != nil {
		s.writeError(nil, errInvalidRequest, "malformed JSON request")
		return
	}
	if len(req.ID) == 0 {
		return // notification (e.g. notifications/initialized): never respond
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.handleRequest(ctx, req)
	}()
}

func (s *Server) handleRequest(ctx context.Context, req rpcRequest) {
	if req.JSONRPC != "2.0" || req.Method == "" {
		s.writeError(req.ID, errInvalidRequest, `invalid request: need "jsonrpc":"2.0" and "method"`)
		return
	}
	switch req.Method {
	case "initialize":
		s.writeResult(req.ID, map[string]any{
			"protocolVersion": protocolVersion,
			"capabilities":    map[string]any{"tools": map[string]any{}},
			"serverInfo":      map[string]any{"name": serverName, "version": serverVersion},
		})
	case "ping":
		s.writeResult(req.ID, map[string]any{})
	case "tools/list":
		s.writeResult(req.ID, map[string]any{"tools": toolSpecs()})
	case "tools/call":
		s.handleToolCall(ctx, req)
	default:
		s.writeError(req.ID, errMethodNotFound, fmt.Sprintf("unknown method %q", req.Method))
	}
}

func (s *Server) writeResult(id json.RawMessage, v any) {
	b, err := json.Marshal(v)
	if err != nil {
		s.writeError(id, errInvalidParams, "internal: marshal result: "+err.Error())
		return
	}
	s.write(rpcResponse{JSONRPC: "2.0", ID: id, Result: b})
}

func (s *Server) writeError(id json.RawMessage, code int, msg string) {
	s.write(rpcResponse{JSONRPC: "2.0", ID: id, Error: &rpcError{Code: code, Message: msg}})
}

// write emits one response as a single atomic line.
func (s *Server) write(res rpcResponse) {
	b, err := json.Marshal(res)
	if err != nil {
		return
	}
	s.outMu.Lock()
	defer s.outMu.Unlock()
	_, _ = s.out.Write(append(b, '\n'))
}

// ---------------------------------------------------------------------------
// tools/call
// ---------------------------------------------------------------------------

type toolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type toolResult struct {
	Content []toolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// toolOK renders v as a JSON text payload.
func toolOK(v any) toolResult {
	b, err := json.Marshal(v)
	if err != nil {
		return toolErr("marshal result: %v", err)
	}
	return toolResult{Content: []toolContent{{Type: "text", Text: string(b)}}}
}

func toolErr(format string, args ...any) toolResult {
	return toolResult{
		Content: []toolContent{{Type: "text", Text: fmt.Sprintf(format, args...)}},
		IsError: true,
	}
}

func (s *Server) handleToolCall(ctx context.Context, req rpcRequest) {
	var p struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &p); err != nil || p.Name == "" {
		s.writeError(req.ID, errInvalidParams, `tools/call requires "name" (string) and optional "arguments" (object)`)
		return
	}
	if !knownTool(p.Name) {
		s.writeError(req.ID, errInvalidParams, fmt.Sprintf("unknown tool %q", p.Name))
		return
	}
	if len(p.Arguments) == 0 {
		p.Arguments = []byte(`{}`)
	}
	s.writeResult(req.ID, s.callTool(ctx, p.Name, p.Arguments))
}

func (s *Server) callTool(ctx context.Context, name string, args json.RawMessage) toolResult {
	switch name {
	case "tdlib_search_methods":
		return s.toolSearchMethods(args)
	case "tdlib_describe_method":
		return s.toolDescribeMethod(args)
	case "tdlib_call":
		return s.toolCall(ctx, args)
	case "tdlib_get_updates":
		return s.toolGetUpdates(args)
	case "tdlib_wait_update":
		return s.toolWaitUpdate(ctx, args)
	case "tdlib_auth_status":
		return s.toolAuthStatus()
	case "tdlib_submit_code":
		return s.toolSubmitCode(args)
	}
	return toolErr("unknown tool %q", name) // unreachable: checked by knownTool
}

// ---------------------------------------------------------------------------
// Tool implementations
// ---------------------------------------------------------------------------

type paramBrief struct {
	Name   string `json:"name"`
	GoType string `json:"goType"`
}

type methodSummary struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Params      []paramBrief `json:"params"`
	ReturnType  string       `json:"returnType"`
}

func summarize(m MethodSchema) methodSummary {
	desc := m.Description
	if r := []rune(desc); len(r) > 200 {
		desc = string(r[:200]) + "…"
	}
	ms := methodSummary{Name: m.Name, Description: desc, ReturnType: m.ReturnType}
	for _, p := range m.Params {
		ms.Params = append(ms.Params, paramBrief{Name: p.Name, GoType: p.GoType})
	}
	if ms.Params == nil {
		ms.Params = []paramBrief{}
	}
	return ms
}

func (s *Server) toolSearchMethods(args json.RawMessage) toolResult {
	var a struct {
		Query string `json:"query"`
		Limit int    `json:"limit"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return toolErr("invalid arguments: %v", err)
	}
	if strings.TrimSpace(a.Query) == "" {
		return toolErr(`"query" is required`)
	}
	if a.Limit <= 0 {
		a.Limit = 10
	}
	if a.Limit > 50 {
		a.Limit = 50
	}
	found := s.cfg.Search(a.Query, a.Limit)
	out := make([]methodSummary, 0, len(found))
	for _, m := range found {
		out = append(out, summarize(m))
	}
	return toolOK(out)
}

func (s *Server) toolDescribeMethod(args json.RawMessage) toolResult {
	var a struct {
		Method string `json:"method"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return toolErr("invalid arguments: %v", err)
	}
	m, ok := s.cfg.Lookup(a.Method)
	if !ok {
		return toolErr("unknown method %q; use tdlib_search_methods to find method names", a.Method)
	}
	return toolOK(m)
}

func (s *Server) toolCall(ctx context.Context, args json.RawMessage) toolResult {
	var a struct {
		Method string         `json:"method"`
		Params map[string]any `json:"params"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return toolErr("invalid arguments: %v", err)
	}
	if a.Method == "" {
		return toolErr(`"method" is required`)
	}
	schema, ok := s.cfg.Lookup(a.Method)
	if !ok {
		return toolErr("unknown method %q; use tdlib_search_methods to find method names", a.Method)
	}
	if err := validateParams(schema, a.Params); err != nil {
		return toolErr("invalid params for %s: %v", a.Method, err)
	}
	callCtx, cancel := context.WithTimeout(ctx, s.cfg.CallTimeout)
	defer cancel()
	raw, err := s.cfg.Call(callCtx, a.Method, a.Params)
	if err != nil {
		var callErr *CallError
		if errors.As(err, &callErr) {
			return toolResult{
				Content: []toolContent{{Type: "text", Text: fmt.Sprintf(`{"code":%d,"message":%s}`, callErr.Code, quoteJSON(callErr.Message))}},
				IsError: true,
			}
		}
		return toolErr("%s failed: %v", a.Method, err)
	}
	return toolResult{Content: []toolContent{{Type: "text", Text: string(raw)}}}
}

func quoteJSON(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// validateParams checks raw params against the method schema before any
// TDLib round-trip: unknown parameter names are rejected (listing the valid
// ones) and union parameters supplied as objects must carry an "@type"
// naming a known variant.
func validateParams(schema MethodSchema, params map[string]any) error {
	if len(params) == 0 {
		return nil
	}
	byName := make(map[string]ParamSchema, len(schema.Params))
	valid := make([]string, 0, len(schema.Params))
	for _, p := range schema.Params {
		byName[p.Name] = p
		valid = append(valid, p.Name)
	}
	for name, value := range params {
		p, ok := byName[name]
		if !ok {
			return fmt.Errorf("unknown parameter %q; valid parameters: [%s]", name, strings.Join(valid, ", "))
		}
		if !p.IsUnion {
			continue
		}
		obj, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("parameter %q is a union (%s); pass an object with \"@type\" set to a variant constructor", name, p.GoType)
		}
		variant, _ := obj["@type"].(string)
		if variant == "" {
			return fmt.Errorf("parameter %q is a union (%s); its object must include \"@type\" (one of: %s)", name, p.GoType, strings.Join(p.Variants, ", "))
		}
		if len(p.Variants) > 0 && !containsString(p.Variants, variant) {
			return fmt.Errorf("parameter %q: invalid variant %q; allowed @type values: [%s]", name, variant, strings.Join(p.Variants, ", "))
		}
	}
	return nil
}

func containsString(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

func (s *Server) toolGetUpdates(args json.RawMessage) toolResult {
	var a struct {
		Type  string `json:"type"`
		Limit int    `json:"limit"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return toolErr("invalid arguments: %v", err)
	}
	if a.Limit <= 0 {
		a.Limit = 50
	}
	if a.Limit > s.cfg.UpdateBufferCap {
		a.Limit = s.cfg.UpdateBufferCap
	}
	entries := s.updates.list(a.Type, a.Limit)
	out := make([]json.RawMessage, 0, len(entries))
	for _, e := range entries {
		out = append(out, e.Raw)
	}
	return toolOK(out)
}

func (s *Server) toolWaitUpdate(ctx context.Context, args json.RawMessage) toolResult {
	var a struct {
		Type           string  `json:"type"`
		TimeoutSeconds float64 `json:"timeout_seconds"`
		Contains       string  `json:"contains"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return toolErr("invalid arguments: %v", err)
	}
	if a.TimeoutSeconds <= 0 {
		a.TimeoutSeconds = 30
	}
	if a.TimeoutSeconds > 120 {
		a.TimeoutSeconds = 120
	}
	waitCtx, cancel := context.WithTimeout(ctx, time.Duration(a.TimeoutSeconds*float64(time.Second)))
	defer cancel()
	entry, err := s.updates.wait(waitCtx, a.Type, a.Contains)
	if err != nil {
		filter := a.Type
		if filter == "" {
			filter = "*"
		}
		return toolErr("no update matching type=%s contains=%q within %.0fs", filter, a.Contains, a.TimeoutSeconds)
	}
	return toolResult{Content: []toolContent{{Type: "text", Text: string(entry.Raw)}}}
}

func (s *Server) toolAuthStatus() toolResult {
	state, authorized := s.updates.authStatus()
	return toolOK(map[string]any{
		"state":      state,
		"authorized": authorized,
	})
}

func (s *Server) toolSubmitCode(args json.RawMessage) toolResult {
	var a struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return toolErr("invalid arguments: %v", err)
	}
	if strings.TrimSpace(a.Code) == "" {
		return toolErr(`"code" is required`)
	}
	if s.cfg.SubmitCode == nil {
		return toolErr("code submission is not wired up in this build")
	}
	if err := s.cfg.SubmitCode(a.Code); err != nil {
		return toolErr("%v", err)
	}
	return toolOK(map[string]any{"status": "submitted"})
}
