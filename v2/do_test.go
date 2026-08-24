package tdlib

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/aliforever/go-tdlib/v2/td"
)

// decodeEnvelope extracts @type and @extra from a marshalled request
// payload plus the full field map.
func decodeEnvelope(t *testing.T, payload []byte) (string, string, map[string]any) {
	t.Helper()
	var head struct {
		Type  string `json:"@type"`
		Extra string `json:"@extra"`
	}
	if err := json.Unmarshal(payload, &head); err != nil {
		t.Fatalf("decode envelope head: %v", err)
	}
	var full map[string]any
	if err := json.Unmarshal(payload, &full); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	return head.Type, head.Extra, full
}

func TestDoHappyPath(t *testing.T) {
	ft := &fakeTransport{}
	c := NewClient(10, ft)
	defer c.Close()

	ft.onSend = func(clientID int, p []byte) {
		tdType, extra, _ := decodeEnvelope(t, p)
		if tdType != "getMe" {
			t.Errorf("envelope @type = %q, want getMe", tdType)
		}
		if extra == "" {
			t.Error("envelope missing @extra")
		}
		// First keys must be @type/@extra (single-marshal injection).
		if string(p[:9]) != `{"@type":` {
			t.Errorf("payload does not start with @type: %s", p)
		}
		resp := `{"@type":"user","id":42,"first_name":"Ada","@extra":` + `"` + extra + `"}`
		c.pending.deliver(extra, []byte(resp))
	}

	user, err := c.Do(context.Background(), td.GetMe{})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	// R = *td.User: new(R) is **User; this asserts json.Unmarshal allocated.
	if user == nil {
		t.Fatal("Do returned nil *td.User")
	}
	if user.Id != 42 || user.FirstName != "Ada" {
		t.Fatalf("unexpected user: %+v", user)
	}
	if c.pending.len() != 0 {
		t.Fatalf("pending leaked: %d", c.pending.len())
	}
}

func TestDoRequestFieldsMarshalled(t *testing.T) {
	ft := &fakeTransport{}
	c := NewClient(11, ft)
	defer c.Close()

	ft.onSend = func(clientID int, p []byte) {
		tdType, extra, full := decodeEnvelope(t, p)
		if tdType != "getMessage" {
			t.Errorf("envelope @type = %q, want getMessage", tdType)
		}
		if full["chat_id"].(float64) != 100 || full["message_id"].(float64) != 7 {
			t.Errorf("request fields missing/wrong: %v", full)
		}
		c.pending.deliver(extra, []byte(`{"@type":"message","id":7,"chat_id":100,"@extra":"`+extra+`"}`))
	}

	msg, err := c.Do(context.Background(), td.GetMessage{ChatId: 100, MessageId: 7})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if msg.Id != 7 || msg.ChatId != 100 {
		t.Fatalf("unexpected message: %+v", msg)
	}
}

func TestDoUnionContainerResult(t *testing.T) {
	// getMessageFileType returns td.MessageFileType, a union container
	// delivered by value (R is not a pointer).
	ft := &fakeTransport{}
	c := NewClient(12, ft)
	defer c.Close()

	ft.onSend = func(clientID int, p []byte) {
		_, extra, _ := decodeEnvelope(t, p)
		c.pending.deliver(extra, []byte(`{"@type":"messageFileTypeUnknown","@extra":"`+extra+`"}`))
	}

	mt, err := c.Do(context.Background(), td.GetMessageFileType{MessageFileHead: "x"})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if mt.Kind() != "messageFileTypeUnknown" || !mt.Known() {
		t.Fatalf("unexpected union container: kind=%q known=%v", mt.Kind(), mt.Known())
	}
	if _, ok := mt.As[*td.MessageFileTypeUnknown](); !ok {
		t.Fatal("As[*td.MessageFileTypeUnknown] failed")
	}
}

func TestDoTDLibError(t *testing.T) {
	ft := &fakeTransport{}
	c := NewClient(13, ft)
	defer c.Close()

	ft.onSend = func(clientID int, p []byte) {
		_, extra, _ := decodeEnvelope(t, p)
		c.pending.deliver(extra, []byte(`{"@type":"error","code":400,"message":"CHAT_NOT_FOUND","@extra":"`+extra+`"}`))
	}

	_, err := c.Do(context.Background(), td.GetChat{ChatId: 1})
	var tdErr *td.Error
	if !errors.As(err, &tdErr) {
		t.Fatalf("expected *td.Error, got %T: %v", err, err)
	}
	if tdErr.Code != 400 || tdErr.Message != "CHAT_NOT_FOUND" {
		t.Fatalf("error fields not intact: %+v", tdErr)
	}
}

func TestDoContextCancel(t *testing.T) {
	ft := &fakeTransport{}
	c := NewClient(14, ft)
	defer c.Close()

	ctx, cancel := context.WithCancel(context.Background())
	ft.onSend = func(clientID int, p []byte) { cancel() }

	_, err := c.Do(ctx, td.GetMe{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if c.pending.len() != 0 {
		t.Fatalf("pending leaked after cancel: %d", c.pending.len())
	}
}

func TestDoRaw(t *testing.T) {
	ft := &fakeTransport{}
	c := NewClient(15, ft)
	defer c.Close()

	ft.onSend = func(clientID int, p []byte) {
		tdType, extra, full := decodeEnvelope(t, p)
		if tdType != "setOption" {
			t.Errorf("envelope @type = %q, want setOption", tdType)
		}
		if full["name"] != "test" {
			t.Errorf("fields not injected: %v", full)
		}
		c.pending.deliver(extra, []byte(`{"@type":"ok","@extra":"`+extra+`"}`))
	}

	raw, err := c.DoRaw(context.Background(), "setOption", map[string]any{
		"name":  "test",
		"value": map[string]any{"@type": "optionValueBoolean", "value": true},
	})
	if err != nil {
		t.Fatalf("DoRaw: %v", err)
	}
	var head struct {
		Type string `json:"@type"`
	}
	if err := json.Unmarshal(raw, &head); err != nil || head.Type != "ok" {
		t.Fatalf("unexpected raw response: %s (%v)", raw, err)
	}

	// nil fields must still produce a valid envelope.
	ft.onSend = func(clientID int, p []byte) {
		tdType, extra, full := decodeEnvelope(t, p)
		if tdType != "getMe" || len(full) != 2 { // @type + @extra only
			t.Errorf("nil-fields envelope wrong: %s", p)
		}
		c.pending.deliver(extra, []byte(`{"@type":"user","id":1,"@extra":"`+extra+`"}`))
	}
	if _, err := c.DoRaw(context.Background(), "getMe", nil); err != nil {
		t.Fatalf("DoRaw nil fields: %v", err)
	}

	// Non-object fields are rejected.
	if _, err := c.DoRaw(context.Background(), "getMe", 42); err == nil {
		t.Fatal("DoRaw accepted a scalar fields payload")
	}
}

func TestDoAsyncAwait(t *testing.T) {
	ft := &fakeTransport{}
	c := NewClient(16, ft)
	defer c.Close()

	ft.onSend = func(clientID int, p []byte) {
		_, extra, _ := decodeEnvelope(t, p)
		c.pending.deliver(extra, []byte(`{"@type":"user","id":99,"@extra":"`+extra+`"}`))
	}

	f := c.DoAsync(context.Background(), td.GetMe{})
	user, err := f.Await(context.Background())
	if err != nil {
		t.Fatalf("Await: %v", err)
	}
	if user == nil || user.Id != 99 {
		t.Fatalf("unexpected user: %+v", user)
	}

	// The result is memoized: a second Await returns the same value.
	again, err := f.Await(context.Background())
	if err != nil || again != user {
		t.Fatalf("second Await differs: %v, %v", again, err)
	}
}

func TestDoAsyncAwaitContextCancel(t *testing.T) {
	ft := &fakeTransport{}
	c := NewClient(17, ft)
	defer c.Close()

	f := c.DoAsync(context.Background(), td.GetMe{}) // never answered
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, err := f.Await(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
}

func TestFutureMapChain(t *testing.T) {
	ft := &fakeTransport{}
	c := NewClient(18, ft)
	defer c.Close()

	ft.onSend = func(clientID int, p []byte) {
		_, extra, _ := decodeEnvelope(t, p)
		c.pending.deliver(extra, []byte(`{"@type":"user","id":7,"first_name":"Ada","@extra":"`+extra+`"}`))
	}

	f := c.DoAsync(context.Background(), td.GetMe{})
	name := f.Map(func(u *td.User) (string, error) { return u.FirstName, nil })
	length := name.Map(func(s string) (int, error) { return len(s), nil })

	n, err := length.Await(context.Background())
	if err != nil {
		t.Fatalf("Await: %v", err)
	}
	if n != 3 {
		t.Fatalf("mapped chain result = %d, want 3", n)
	}
}

func TestFutureMapErrorPropagation(t *testing.T) {
	ft := &fakeTransport{}
	c := NewClient(19, ft)
	defer c.Close()

	// Case 1: the request fails; fn must not run and the error propagates.
	ft.onSend = func(clientID int, p []byte) {
		_, extra, _ := decodeEnvelope(t, p)
		c.pending.deliver(extra, []byte(`{"@type":"error","code":500,"message":"boom","@extra":"`+extra+`"}`))
	}
	ran := false
	f := c.DoAsync(context.Background(), td.GetMe{}).Map(func(u *td.User) (int, error) {
		ran = true
		return 0, nil
	})
	_, err := f.Await(context.Background())
	var tdErr *td.Error
	if !errors.As(err, &tdErr) || tdErr.Code != 500 {
		t.Fatalf("expected *td.Error 500, got %v", err)
	}
	if ran {
		t.Fatal("Map fn ran despite a failed parent future")
	}

	// Case 2: the mapping function itself fails.
	ft.onSend = func(clientID int, p []byte) {
		_, extra, _ := decodeEnvelope(t, p)
		c.pending.deliver(extra, []byte(`{"@type":"user","id":1,"@extra":"`+extra+`"}`))
	}
	mapErr := errors.New("map failed")
	g := c.DoAsync(context.Background(), td.GetMe{}).Map(func(u *td.User) (int, error) {
		return 0, mapErr
	})
	if _, err := g.Await(context.Background()); !errors.Is(err, mapErr) {
		t.Fatalf("expected map error, got %v", err)
	}
}

// TestFacadeGetMe exercises one generated facade method end to end; the
// other 1008 compile in facade_gen.go by the same shape.
func TestFacadeGetMe(t *testing.T) {
	ft := &fakeTransport{}
	c := NewClient(20, ft)
	defer c.Close()

	ft.onSend = func(clientID int, p []byte) {
		tdType, extra, _ := decodeEnvelope(t, p)
		if tdType != "getMe" {
			t.Errorf("facade sent @type %q, want getMe", tdType)
		}
		c.pending.deliver(extra, []byte(`{"@type":"user","id":555,"@extra":"`+extra+`"}`))
	}

	user, err := c.GetMe(context.Background())
	if err != nil {
		t.Fatalf("GetMe facade: %v", err)
	}
	if user.Id != 555 {
		t.Fatalf("unexpected user id: %d", user.Id)
	}
}
