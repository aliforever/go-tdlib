package tdlib

import (
	"errors"
	"testing"
	"time"
)

func TestPendingRegisterDeliver(t *testing.T) {
	reg := newPendingRegistry()
	call := reg.register("e1")

	reg.deliver("e1", []byte(`{"@type":"ok"}`))

	select {
	case res := <-call.ch:
		if res.err != nil {
			t.Fatalf("unexpected error: %v", res.err)
		}
		if string(res.raw) != `{"@type":"ok"}` {
			t.Fatalf("unexpected raw: %s", res.raw)
		}
	default:
		t.Fatal("expected a delivered result")
	}
}

func TestPendingDeliverCopiesRaw(t *testing.T) {
	reg := newPendingRegistry()
	call := reg.register("e1")

	raw := []byte(`{"@type":"ok"}`)
	reg.deliver("e1", raw)
	raw[2] = 'X' // mutate caller buffer after delivery

	res := <-call.ch
	if string(res.raw) != `{"@type":"ok"}` {
		t.Fatalf("registry retained caller's buffer: %s", res.raw)
	}
}

func TestPendingDeliverUnknownExtraDropped(t *testing.T) {
	reg := newPendingRegistry()
	// Must not panic, must not block.
	reg.deliver("nobody-home", []byte(`{}`))
}

func TestPendingCompleteExactlyOnce(t *testing.T) {
	call := newPendingCall()
	call.complete(callResult{raw: []byte(`first`)})
	call.complete(callResult{raw: []byte(`second`)}) // must be dropped, must not block

	res := <-call.ch
	if string(res.raw) != `first` {
		t.Fatalf("second completion overwrote the first: %s", res.raw)
	}
	select {
	case extra := <-call.ch:
		t.Fatalf("second completion leaked into channel: %v", extra)
	default:
	}
}

func TestPendingFailAll(t *testing.T) {
	reg := newPendingRegistry()
	c1 := reg.register("e1")
	c2 := reg.register("e2")

	sentinel := errors.New("shutting down")
	reg.failAll(sentinel)

	for _, call := range []*pendingCall{c1, c2} {
		select {
		case res := <-call.ch:
			if !errors.Is(res.err, sentinel) {
				t.Fatalf("expected sentinel, got %v", res.err)
			}
		default:
			t.Fatal("expected failed result")
		}
	}
}

func TestPendingRemoveIdempotent(t *testing.T) {
	reg := newPendingRegistry()
	reg.register("e1")
	reg.remove("e1")
	reg.remove("e1") // no panic
	if reg.len() != 0 {
		t.Fatalf("registry not empty: %d", reg.len())
	}
	// Late delivery after removal is dropped.
	reg.deliver("e1", []byte(`{}`))
}

func TestPendingDeliverNeverBlocksDispatcher(t *testing.T) {
	reg := newPendingRegistry()
	call := reg.register("e1")
	call.complete(callResult{raw: []byte(`filled`)})

	done := make(chan struct{})
	go func() {
		defer close(done)
		reg.deliver("e1", []byte(`{}`)) // already completed: must return, not block
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("deliver blocked on a completed call")
	}
}
