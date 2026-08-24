package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Fake backend
// ---------------------------------------------------------------------------

var fakeRegistry = map[string]MethodSchema{
	"sendMessage": {
		Name:        "sendMessage",
		Description: "Sends a message. Returns the sent message.",
		Params: []ParamSchema{
			{Name: "chat_id", GoType: "int64", Description: "Target chat."},
			{Name: "message", GoType: "InputMessage", IsUnion: true,
				Variants:    []string{"inputMessageText", "inputMessagePhoto"},
				Description: "The content of the message to be sent."},
		},
		ReturnType: "Message",
	},
	"getMe": {
		Name:        "getMe",
		Description: "Returns the current user.",
		ReturnType:  "User",
	},
}

type fakeCall struct {
	method string
	params map[string]any
}

type fakeBackend struct {
	mu        sync.Mutex
	calls     []fakeCall
	response  json.RawMessage
	callErr   error
	submitErr error
	submitted []string
}

func (f *fakeBackend) lookup(name string) (MethodSchema, bool) {
	m, ok := fakeRegistry[name]
	return m, ok
}

func (f *fakeBackend) search(query string, limit int) []MethodSchema {
	var out []MethodSchema
	for _, m := range fakeRegistry {
		if strings.Contains(strings.ToLower(m.Name), strings.ToLower(query)) ||
			strings.Contains(strings.ToLower(m.Description), strings.ToLower(query)) {
			out = append(out, m)
		}
	}
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}

func (f *fakeBackend) call(_ context.Context, method string, params map[string]any) (json.RawMessage, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, fakeCall{method: method, params: params})
	if f.callErr != nil {
		return nil, f.callErr
	}
	return f.response, nil
}

func (f *fakeBackend) submitCode(code string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.submitErr != nil {
		return f.submitErr
	}
	f.submitted = append(f.submitted, code)
	return nil
}

func (f *fakeBackend) config() Config {
	f.response = json.RawMessage(`{"@type":"message","id":42}`)
	return Config{
		Lookup:     f.lookup,
		Search:     f.search,
		Call:       f.call,
		SubmitCode: f.submitCode,
	}
}

// ---------------------------------------------------------------------------
// In-memory MCP client harness
// ---------------------------------------------------------------------------

// mcpClient talks to a Server over in-memory pipes. A single read loop
// routes responses to waiting callers by id, so concurrent tools/call
// requests (whose responses may arrive out of order) work from tests.
type mcpClient struct {
	t    *testing.T
	in   *io.PipeWriter
	done chan error

	mu      sync.Mutex
	nextID  int
	pending map[string]chan rpcMsg
	dead    chan struct{}
	readErr error
}

func newMCPClient(t *testing.T, cfg Config) (*mcpClient, *Server) {
	t.Helper()
	inR, inW := io.Pipe()
	outR, outW := io.Pipe()
	srv := NewServer(cfg)
	done := make(chan error, 1)
	go func() { done <- srv.Serve(context.Background(), inR, outW) }()
	c := &mcpClient{
		t:       t,
		in:      inW,
		done:    done,
		pending: map[string]chan rpcMsg{},
		dead:    make(chan struct{}),
	}
	go c.readLoop(bufio.NewReaderSize(outR, 1<<20))
	t.Cleanup(func() { _ = inW.Close(); _ = outR.Close(); _ = outW.Close() })
	return c, srv
}

func (c *mcpClient) readLoop(r *bufio.Reader) {
	for {
		line, err := r.ReadBytes('\n')
		if err != nil {
			c.mu.Lock()
			c.readErr = err
			close(c.dead)
			c.mu.Unlock()
			return
		}
		var m rpcMsg
		if err := json.Unmarshal(line, &m); err != nil {
			continue
		}
		key := "null"
		if len(m.ID) > 0 && string(m.ID) != "null" {
			key = string(m.ID)
		}
		c.mu.Lock()
		ch, ok := c.pending[key]
		delete(c.pending, key)
		c.mu.Unlock()
		if ok {
			ch <- m
		}
	}
}

func (c *mcpClient) send(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = c.in.Write(append(b, '\n'))
	return err
}

// exchange registers a response slot for key, sends payload, and awaits the
// routed response.
func (c *mcpClient) exchange(key string, payload any) (rpcMsg, error) {
	ch := make(chan rpcMsg, 1)
	c.mu.Lock()
	c.pending[key] = ch
	c.mu.Unlock()
	if err := c.send(payload); err != nil {
		return rpcMsg{}, err
	}
	select {
	case m := <-ch:
		return m, nil
	case <-c.dead:
		return rpcMsg{}, fmt.Errorf("reader died: %v", c.readErr)
	case <-time.After(10 * time.Second):
		return rpcMsg{}, fmt.Errorf("timed out waiting for response (id %s)", key)
	}
}

type rpcMsg struct {
	ID     json.RawMessage `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  *rpcError       `json:"error"`
}

// requestE sends one JSON-RPC request and awaits its response.
func (c *mcpClient) requestE(method string, params any) (rpcMsg, error) {
	c.mu.Lock()
	c.nextID++
	id := c.nextID
	c.mu.Unlock()
	m := map[string]any{"jsonrpc": "2.0", "id": id, "method": method}
	if params != nil {
		m["params"] = params
	}
	res, err := c.exchange(fmt.Sprintf("%d", id), m)
	if err != nil {
		return rpcMsg{}, err
	}
	if string(res.ID) != fmt.Sprintf("%d", id) {
		return rpcMsg{}, fmt.Errorf("response id = %s, want %d", res.ID, id)
	}
	return res, nil
}

func (c *mcpClient) request(method string, params any) rpcMsg {
	c.t.Helper()
	res, err := c.requestE(method, params)
	if err != nil {
		c.t.Fatalf("request %s: %v", method, err)
	}
	return res
}

// sendRaw writes one unframed line verbatim and awaits the response (error
// responses to malformed frames carry a null id).
func (c *mcpClient) sendRaw(s string) (rpcMsg, error) {
	ch := make(chan rpcMsg, 1)
	c.mu.Lock()
	c.pending["null"] = ch
	c.mu.Unlock()
	if _, err := c.in.Write([]byte(s + "\n")); err != nil {
		return rpcMsg{}, err
	}
	select {
	case m := <-ch:
		return m, nil
	case <-c.dead:
		return rpcMsg{}, fmt.Errorf("reader died: %v", c.readErr)
	case <-time.After(10 * time.Second):
		return rpcMsg{}, fmt.Errorf("timed out waiting for response to raw frame")
	}
}

// callToolE invokes tools/call and unwraps the MCP tool result.
func (c *mcpClient) callToolE(name string, args map[string]any) (text string, isError bool, err error) {
	res, err := c.requestE("tools/call", map[string]any{"name": name, "arguments": args})
	if err != nil {
		return "", false, err
	}
	if res.Error != nil {
		return "", false, fmt.Errorf("rpc error %d: %s", res.Error.Code, res.Error.Message)
	}
	var tr toolResult
	if err := json.Unmarshal(res.Result, &tr); err != nil {
		return "", false, fmt.Errorf("decode tool result: %v", err)
	}
	if len(tr.Content) != 1 || tr.Content[0].Type != "text" {
		return "", false, fmt.Errorf("unexpected tool content: %+v", tr.Content)
	}
	return tr.Content[0].Text, tr.IsError, nil
}

func (c *mcpClient) callTool(name string, args map[string]any) (text string, isError bool) {
	c.t.Helper()
	text, isError, err := c.callToolE(name, args)
	if err != nil {
		c.t.Fatalf("tools/call %s: %v", name, err)
	}
	return text, isError
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestInitializeAndPing(t *testing.T) {
	c, _ := newMCPClient(t, (&fakeBackend{}).config())

	res := c.request("initialize", nil)
	if res.Error != nil {
		t.Fatalf("initialize error: %+v", res.Error)
	}
	var init struct {
		ProtocolVersion string `json:"protocolVersion"`
		ServerInfo      struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"serverInfo"`
		Capabilities map[string]any `json:"capabilities"`
	}
	if err := json.Unmarshal(res.Result, &init); err != nil {
		t.Fatalf("decode initialize result: %v", err)
	}
	if init.ProtocolVersion != "2024-11-05" {
		t.Errorf("protocolVersion = %q", init.ProtocolVersion)
	}
	if init.ServerInfo.Name != "tdlib-mcp" || init.ServerInfo.Version == "" {
		t.Errorf("serverInfo = %+v", init.ServerInfo)
	}
	if _, ok := init.Capabilities["tools"]; !ok {
		t.Errorf("capabilities missing tools: %v", init.Capabilities)
	}

	res = c.request("ping", nil)
	if res.Error != nil {
		t.Fatalf("ping error: %+v", res.Error)
	}
	if string(res.Result) != "{}" {
		t.Errorf("ping result = %s, want {}", res.Result)
	}
}

func TestNotificationGetsNoResponse(t *testing.T) {
	c, _ := newMCPClient(t, (&fakeBackend{}).config())

	// A notification (no id) must never produce a response; the ping right
	// after must be answered with the ping's own id.
	c.send(map[string]any{"jsonrpc": "2.0", "method": "notifications/initialized"})
	res := c.request("ping", nil)
	if res.Error != nil {
		t.Fatalf("ping after notification: %+v", res.Error)
	}
}

func TestUnknownAndMalformed(t *testing.T) {
	c, _ := newMCPClient(t, (&fakeBackend{}).config())

	res := c.request("bogus/method", nil)
	if res.Error == nil || res.Error.Code != errMethodNotFound {
		t.Fatalf("unknown method: want -32601, got %+v", res.Error)
	}

	res, err := c.sendRaw("{not json")
	if err != nil {
		t.Fatalf("malformed frame: %v", err)
	}
	if res.Error == nil || res.Error.Code != errInvalidRequest {
		t.Fatalf("malformed JSON: want -32600, got %+v", res.Error)
	}

	res = c.request("tools/call", map[string]any{"arguments": map[string]any{}}) // no name
	if res.Error == nil || res.Error.Code != errInvalidParams {
		t.Fatalf("tools/call without name: want -32602, got %+v", res.Error)
	}

	res = c.request("tools/call", map[string]any{"name": "no_such_tool"})
	if res.Error == nil || res.Error.Code != errInvalidParams {
		t.Fatalf("unknown tool: want -32602, got %+v", res.Error)
	}
}

func TestToolsList(t *testing.T) {
	c, _ := newMCPClient(t, (&fakeBackend{}).config())

	res := c.request("tools/list", nil)
	if res.Error != nil {
		t.Fatalf("tools/list error: %+v", res.Error)
	}
	var out struct {
		Tools []toolSpec `json:"tools"`
	}
	if err := json.Unmarshal(res.Result, &out); err != nil {
		t.Fatalf("decode tools/list: %v", err)
	}
	want := []string{
		"tdlib_search_methods", "tdlib_describe_method", "tdlib_call",
		"tdlib_get_updates", "tdlib_wait_update", "tdlib_auth_status",
		"tdlib_submit_code",
	}
	if len(out.Tools) != len(want) {
		t.Fatalf("tools/list returned %d tools, want %d", len(out.Tools), len(want))
	}
	for i, name := range want {
		if out.Tools[i].Name != name {
			t.Errorf("tool[%d] = %q, want %q", i, out.Tools[i].Name, name)
		}
		if out.Tools[i].InputSchema["type"] != "object" {
			t.Errorf("tool %q inputSchema.type = %v", name, out.Tools[i].InputSchema["type"])
		}
	}
}

func TestCallValidation(t *testing.T) {
	backend := &fakeBackend{}
	c, _ := newMCPClient(t, backend.config())

	// Unknown method.
	text, isErr := c.callTool("tdlib_call", map[string]any{"method": "noSuchMethod"})
	if !isErr || !strings.Contains(text, "unknown method") {
		t.Errorf("unknown method: isErr=%v text=%q", isErr, text)
	}

	// Unknown parameter; error must list the valid names.
	text, isErr = c.callTool("tdlib_call", map[string]any{
		"method": "sendMessage",
		"params": map[string]any{"chat_id": 1, "bogus_param": 2},
	})
	if !isErr || !strings.Contains(text, "unknown parameter") ||
		!strings.Contains(text, "chat_id") || !strings.Contains(text, "message") {
		t.Errorf("unknown param: isErr=%v text=%q", isErr, text)
	}

	// Top-level "@type" is not a valid param and must be rejected.
	text, isErr = c.callTool("tdlib_call", map[string]any{
		"method": "sendMessage",
		"params": map[string]any{"@type": "sendMessage", "chat_id": 1},
	})
	if !isErr || !strings.Contains(text, "unknown parameter") {
		t.Errorf("envelope @type injection: isErr=%v text=%q", isErr, text)
	}

	// Union param missing "@type".
	text, isErr = c.callTool("tdlib_call", map[string]any{
		"method": "sendMessage",
		"params": map[string]any{"chat_id": 1, "message": map[string]any{"text": "hi"}},
	})
	if !isErr || !strings.Contains(text, `"@type"`) {
		t.Errorf("union missing @type: isErr=%v text=%q", isErr, text)
	}

	// Union param with a variant not in the schema.
	text, isErr = c.callTool("tdlib_call", map[string]any{
		"method": "sendMessage",
		"params": map[string]any{"chat_id": 1, "message": map[string]any{"@type": "inputMessageVideo"}},
	})
	if !isErr || !strings.Contains(text, "inputMessageText") || !strings.Contains(text, "inputMessagePhoto") {
		t.Errorf("bad union variant: isErr=%v text=%q", isErr, text)
	}

	// Validation failures must not reach the backend.
	backend.mu.Lock()
	if len(backend.calls) != 0 {
		t.Errorf("backend called %d times despite validation failures", len(backend.calls))
	}
	backend.mu.Unlock()

	// Valid call: passes validation, params reach the backend untouched
	// (nested "@type" preserved, no envelope "@type" added by the server).
	text, isErr = c.callTool("tdlib_call", map[string]any{
		"method": "sendMessage",
		"params": map[string]any{
			"chat_id": 123,
			"message": map[string]any{"@type": "inputMessageText", "text": map[string]any{"text": "hi"}},
		},
	})
	if isErr {
		t.Fatalf("valid call failed: %q", text)
	}
	if text != `{"@type":"message","id":42}` {
		t.Errorf("call result = %q", text)
	}
	backend.mu.Lock()
	defer backend.mu.Unlock()
	if len(backend.calls) != 1 {
		t.Fatalf("backend calls = %d, want 1", len(backend.calls))
	}
	got := backend.calls[0]
	if got.method != "sendMessage" {
		t.Errorf("backend method = %q", got.method)
	}
	if _, ok := got.params["@type"]; ok {
		t.Errorf("server injected envelope @type into params: %v", got.params)
	}
	msg, ok := got.params["message"].(map[string]any)
	if !ok || msg["@type"] != "inputMessageText" {
		t.Errorf("nested @type not preserved: %v", got.params["message"])
	}
}

func TestCallTDLibError(t *testing.T) {
	backend := &fakeBackend{callErr: &CallError{Code: 400, Message: "CHAT_NOT_FOUND"}}
	c, _ := newMCPClient(t, backend.config())

	text, isErr := c.callTool("tdlib_call", map[string]any{"method": "getMe"})
	if !isErr {
		t.Fatalf("expected isError, got %q", text)
	}
	var e struct {
		Code    int32  `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal([]byte(text), &e); err != nil {
		t.Fatalf("error payload not JSON {code,message}: %q", text)
	}
	if e.Code != 400 || e.Message != "CHAT_NOT_FOUND" {
		t.Errorf("error payload = %+v", e)
	}
}

func TestDescribeAndSearch(t *testing.T) {
	c, _ := newMCPClient(t, (&fakeBackend{}).config())

	text, isErr := c.callTool("tdlib_describe_method", map[string]any{"method": "sendMessage"})
	if isErr {
		t.Fatalf("describe: %q", text)
	}
	var m MethodSchema
	if err := json.Unmarshal([]byte(text), &m); err != nil {
		t.Fatalf("decode schema: %v", err)
	}
	if m.Name != "sendMessage" || m.ReturnType != "Message" || len(m.Params) != 2 {
		t.Errorf("schema = %+v", m)
	}
	if !m.Params[1].IsUnion || len(m.Params[1].Variants) != 2 || m.Params[1].Description == "" {
		t.Errorf("union param = %+v", m.Params[1])
	}

	_, isErr = c.callTool("tdlib_describe_method", map[string]any{"method": "nope"})
	if !isErr {
		t.Error("describe unknown method: expected isError")
	}

	text, isErr = c.callTool("tdlib_search_methods", map[string]any{"query": "message"})
	if isErr {
		t.Fatalf("search: %q", text)
	}
	var summaries []methodSummary
	if err := json.Unmarshal([]byte(text), &summaries); err != nil {
		t.Fatalf("decode search: %v", err)
	}
	if len(summaries) != 1 || summaries[0].Name != "sendMessage" {
		t.Fatalf("search results = %+v", summaries)
	}
	if len(summaries[0].Params) != 2 || summaries[0].Params[0].Name != "chat_id" || summaries[0].Params[0].GoType != "int64" {
		t.Errorf("summary params = %+v", summaries[0].Params)
	}

	_, isErr = c.callTool("tdlib_search_methods", map[string]any{})
	if !isErr {
		t.Error("search without query: expected isError")
	}
}

func TestGetUpdates(t *testing.T) {
	c, srv := newMCPClient(t, (&fakeBackend{}).config())

	srv.HandleUpdate("updateNewMessage", json.RawMessage(`{"@type":"updateNewMessage","message":{"id":1}}`))
	srv.HandleUpdate("updateUserStatus", json.RawMessage(`{"@type":"updateUserStatus","user_id":7}`))
	srv.HandleUpdate("updateNewMessage", json.RawMessage(`{"@type":"updateNewMessage","message":{"id":2}}`))

	text, isErr := c.callTool("tdlib_get_updates", map[string]any{})
	if isErr {
		t.Fatalf("get_updates: %q", text)
	}
	var all []json.RawMessage
	if err := json.Unmarshal([]byte(text), &all); err != nil {
		t.Fatalf("decode updates: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("updates = %d, want 3", len(all))
	}
	if !strings.Contains(string(all[2]), `"id":2`) {
		t.Errorf("expected oldest-first order, last = %s", all[2])
	}

	text, _ = c.callTool("tdlib_get_updates", map[string]any{"type": "updateNewMessage", "limit": 1})
	var filtered []json.RawMessage
	if err := json.Unmarshal([]byte(text), &filtered); err != nil {
		t.Fatalf("decode filtered: %v", err)
	}
	if len(filtered) != 1 || !strings.Contains(string(filtered[0]), `"id":2`) {
		t.Errorf("filtered+limit = %v", filtered)
	}
}

func TestWaitUpdate(t *testing.T) {
	c, srv := newMCPClient(t, (&fakeBackend{}).config())

	// The waiter is registered asynchronously by the serving goroutine;
	// give it a moment before publishing.
	type result struct {
		text  string
		isErr bool
		err   error
	}
	waitResult := make(chan result, 1)
	go func() {
		text, isErr, err := c.callToolE("tdlib_wait_update", map[string]any{
			"type":     "updateNewMessage",
			"contains": `"id":99`,
		})
		waitResult <- result{text, isErr, err}
	}()
	time.Sleep(100 * time.Millisecond)
	srv.HandleUpdate("updateUserStatus", json.RawMessage(`{"@type":"updateUserStatus"}`)) // no match
	srv.HandleUpdate("updateNewMessage", json.RawMessage(`{"@type":"updateNewMessage","message":{"id":1}}`))
	srv.HandleUpdate("updateNewMessage", json.RawMessage(`{"@type":"updateNewMessage","message":{"id":99}}`))

	select {
	case r := <-waitResult:
		if r.err != nil {
			t.Fatalf("wait_update: %v", r.err)
		}
		if r.isErr {
			t.Fatalf("wait_update failed: %q", r.text)
		}
		if !strings.Contains(r.text, `"id":99`) {
			t.Errorf("wait_update returned %q", r.text)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("wait_update did not complete")
	}

	// Timeout path.
	_, isErr := c.callTool("tdlib_wait_update", map[string]any{
		"type":            "updateNever",
		"timeout_seconds": 0.2,
	})
	if !isErr {
		t.Error("wait_update timeout: expected isError")
	}
}

func TestConcurrentWaiters(t *testing.T) {
	c, srv := newMCPClient(t, (&fakeBackend{}).config())

	type result struct {
		text  string
		isErr bool
		err   error
	}
	done := make(chan result, 2)
	go func() {
		text, isErr, err := c.callToolE("tdlib_wait_update", map[string]any{"type": "updateA"})
		done <- result{text, isErr, err}
	}()
	go func() {
		text, isErr, err := c.callToolE("tdlib_wait_update", map[string]any{"type": "updateB"})
		done <- result{text, isErr, err}
	}()
	time.Sleep(100 * time.Millisecond)
	srv.HandleUpdate("updateB", json.RawMessage(`{"@type":"updateB"}`))
	srv.HandleUpdate("updateA", json.RawMessage(`{"@type":"updateA"}`))

	got := map[string]bool{}
	for i := 0; i < 2; i++ {
		select {
		case r := <-done:
			if r.err != nil || r.isErr {
				t.Fatalf("concurrent waiter failed: %v %q", r.err, r.text)
			}
			got[r.text] = true
		case <-time.After(5 * time.Second):
			t.Fatal("concurrent waiters did not complete")
		}
	}
	if !got[`{"@type":"updateA"}`] || !got[`{"@type":"updateB"}`] {
		t.Errorf("concurrent waiter results = %v", got)
	}
}

func TestAuthStatus(t *testing.T) {
	c, srv := newMCPClient(t, (&fakeBackend{}).config())

	text, _ := c.callTool("tdlib_auth_status", map[string]any{})
	var st struct {
		State      string `json:"state"`
		Authorized bool   `json:"authorized"`
	}
	if err := json.Unmarshal([]byte(text), &st); err != nil {
		t.Fatalf("decode auth status: %v", err)
	}
	if st.State != "unknown" || st.Authorized {
		t.Errorf("initial auth status = %+v", st)
	}

	srv.HandleUpdate("updateAuthorizationState", json.RawMessage(
		`{"@type":"updateAuthorizationState","authorization_state":{"@type":"authorizationStateWaitCode"}}`))
	text, _ = c.callTool("tdlib_auth_status", map[string]any{})
	if err := json.Unmarshal([]byte(text), &st); err != nil {
		t.Fatalf("decode auth status: %v", err)
	}
	if st.State != "authorizationStateWaitCode" || st.Authorized {
		t.Errorf("waiting auth status = %+v", st)
	}

	srv.HandleUpdate("updateAuthorizationState", json.RawMessage(
		`{"@type":"updateAuthorizationState","authorization_state":{"@type":"authorizationStateReady"}}`))
	text, _ = c.callTool("tdlib_auth_status", map[string]any{})
	if err := json.Unmarshal([]byte(text), &st); err != nil {
		t.Fatalf("decode auth status: %v", err)
	}
	if st.State != "authorizationStateReady" || !st.Authorized {
		t.Errorf("ready auth status = %+v", st)
	}
}

func TestSubmitCode(t *testing.T) {
	backend := &fakeBackend{}
	c, _ := newMCPClient(t, backend.config())

	text, isErr := c.callTool("tdlib_submit_code", map[string]any{"code": "12345"})
	if isErr {
		t.Fatalf("submit_code: %q", text)
	}
	if !strings.Contains(text, "submitted") {
		t.Errorf("submit_code result = %q", text)
	}
	backend.mu.Lock()
	if len(backend.submitted) != 1 || backend.submitted[0] != "12345" {
		t.Errorf("submitted codes = %v", backend.submitted)
	}
	backend.mu.Unlock()

	backend.mu.Lock()
	backend.submitErr = fmt.Errorf("the authorization flow is not waiting for a code right now")
	backend.mu.Unlock()
	_, isErr = c.callTool("tdlib_submit_code", map[string]any{"code": "12345"})
	if !isErr {
		t.Error("submit_code while not waiting: expected isError")
	}

	_, isErr = c.callTool("tdlib_submit_code", map[string]any{})
	if !isErr {
		t.Error("submit_code without code: expected isError")
	}
}

func TestUpdateBufferRing(t *testing.T) {
	b := newUpdateBuffer(3)
	for i := 0; i < 5; i++ {
		b.add("updateX", json.RawMessage(fmt.Sprintf(`{"@type":"updateX","n":%d}`, i)))
	}
	entries := b.list("", 10)
	if len(entries) != 3 {
		t.Fatalf("ring size = %d, want 3", len(entries))
	}
	// Oldest of the retained (n=2) first, newest (n=4) last.
	if !strings.Contains(string(entries[0].Raw), `"n":2`) ||
		!strings.Contains(string(entries[2].Raw), `"n":4`) {
		t.Errorf("ring order wrong: %v", entries)
	}
}
