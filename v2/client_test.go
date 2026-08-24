package tdlib

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aliforever/go-tdlib/v2/td"
)

func TestRoundTripSuccess(t *testing.T) {
	ft := &fakeTransport{}
	c := NewClient(1, ft)
	defer c.Close()

	extra := NewExtra()
	payload := []byte(fmt.Sprintf(`{"@type":"getMe","@extra":%q}`, extra))

	ft.onSend = func(clientID int, p []byte) {
		if clientID != 1 {
			t.Errorf("sent with wrong client ID: %d", clientID)
		}
		resp := fmt.Sprintf(`{"@type":"user","id":42,"@extra":%q}`, extraOf(t, p))
		c.pending.deliver(extraOf(t, p), []byte(resp))
	}

	raw, err := c.roundTrip(context.Background(), payload, extra)
	if err != nil {
		t.Fatalf("roundTrip: %v", err)
	}

	var user struct {
		Type string `json:"@type"`
		ID   int64  `json:"id"`
	}
	if err := json.Unmarshal(raw, &user); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if user.Type != "user" || user.ID != 42 {
		t.Fatalf("unexpected response: %+v", user)
	}
	if c.pending.len() != 0 {
		t.Fatalf("pending registry leaked: %d entries", c.pending.len())
	}
}

func TestRoundTripTDLibError(t *testing.T) {
	ft := &fakeTransport{}
	c := NewClient(1, ft)
	defer c.Close()

	ft.onSend = func(clientID int, p []byte) {
		extra := extraOf(t, p)
		resp := fmt.Sprintf(`{"@type":"error","code":400,"message":"PHONE_NUMBER_INVALID","@extra":%q}`, extra)
		c.pending.deliver(extra, []byte(resp))
	}

	extra := NewExtra()
	payload := []byte(fmt.Sprintf(`{"@type":"setAuthenticationPhoneNumber","@extra":%q}`, extra))

	_, err := c.roundTrip(context.Background(), payload, extra)
	if err == nil {
		t.Fatal("expected error")
	}

	var tdErr *td.Error
	if !errors.As(err, &tdErr) {
		t.Fatalf("errors.As(*td.Error) failed for %T: %v", err, err)
	}
	if tdErr.Code != 400 || tdErr.Message != "PHONE_NUMBER_INVALID" {
		t.Fatalf("error fields not intact: %+v", tdErr)
	}
}

func TestRoundTripContextCancel(t *testing.T) {
	ft := &fakeTransport{}
	c := NewClient(1, ft)
	defer c.Close()

	ctx, cancel := context.WithCancel(context.Background())
	// onSend runs inside roundTrip after registration, so cancelling here
	// deterministically aborts the in-flight wait.
	ft.onSend = func(clientID int, p []byte) { cancel() }

	extra := NewExtra()
	payload := []byte(fmt.Sprintf(`{"@type":"getMe","@extra":%q}`, extra))

	_, err := c.roundTrip(ctx, payload, extra)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if c.pending.len() != 0 {
		t.Fatalf("pending entry leaked after cancel: %d", c.pending.len())
	}

	// The late response (arriving after removal) must be dropped silently.
	c.pending.deliver(extra, []byte(`{"@type":"user"}`))
}

func TestRoundTripClientCloseMidFlight(t *testing.T) {
	ft := &fakeTransport{}
	c := NewClient(2, ft)

	ft.onSend = func(clientID int, p []byte) {
		// Registration already happened at this point; closing fails the
		// in-flight call via failAll and cancels the client context.
		go c.Close()
	}

	extra := NewExtra()
	payload := []byte(fmt.Sprintf(`{"@type":"getMe","@extra":%q}`, extra))

	_, err := c.roundTrip(context.Background(), payload, extra)
	if !errors.Is(err, ErrClientClosed) {
		t.Fatalf("expected ErrClientClosed, got %v", err)
	}
	if c.pending.len() != 0 {
		t.Fatalf("pending entry leaked after close: %d", c.pending.len())
	}
}

func TestRoundTripAfterClose(t *testing.T) {
	ft := &fakeTransport{}
	c := NewClient(3, ft)
	c.Close()

	_, err := c.roundTrip(context.Background(), []byte(`{}`), NewExtra())
	if !errors.Is(err, ErrClientClosed) {
		t.Fatalf("expected ErrClientClosed, got %v", err)
	}
	if ft.sentCount() != 0 {
		t.Fatal("closed client must not send")
	}
}

func TestRoundTripConcurrentCallsFailedByClose(t *testing.T) {
	ft := &fakeTransport{}
	c := NewClient(4, ft)

	const n = 8
	errs := make(chan error, n)
	for i := 0; i < n; i++ {
		go func() {
			extra := NewExtra()
			payload := []byte(fmt.Sprintf(`{"@type":"getMe","@extra":%q}`, extra))
			_, err := c.roundTrip(context.Background(), payload, extra)
			errs <- err
		}()
	}

	// Wait until all calls are in flight, then close.
	deadline := time.Now().Add(2 * time.Second)
	for c.pending.len() < n {
		if time.Now().After(deadline) {
			t.Fatalf("only %d/%d calls in flight", c.pending.len(), n)
		}
		time.Sleep(time.Millisecond)
	}
	c.Close()

	for i := 0; i < n; i++ {
		if err := <-errs; !errors.Is(err, ErrClientClosed) {
			t.Fatalf("call %d: expected ErrClientClosed, got %v", i, err)
		}
	}
	if c.pending.len() != 0 {
		t.Fatalf("pending entries leaked: %d", c.pending.len())
	}
}

func TestRoundTripSendError(t *testing.T) {
	sendErr := errors.New("wire broken")
	ft := &fakeTransport{sendErr: sendErr}
	c := NewClient(5, ft)
	defer c.Close()

	_, err := c.roundTrip(context.Background(), []byte(`{}`), NewExtra())
	if !errors.Is(err, sendErr) {
		t.Fatalf("expected send error, got %v", err)
	}
	if c.pending.len() != 0 {
		t.Fatalf("pending entry leaked after send error: %d", c.pending.len())
	}
}
