package tdlib

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"
)

// collectSubscriber returns a subscriber whose handler appends every
// received raw message into ch, plus the channel.
func collectSubscriber(tdType string, opts SubscriberOptions, ch chan json.RawMessage) *subscriber {
	return newSubscriber(tdType, func(ctx context.Context, raw json.RawMessage) {
		ch <- append(json.RawMessage(nil), raw...)
	}, opts)
}

func TestRouterRegisterDispatchCancel(t *testing.T) {
	r := newRouter()
	got := make(chan json.RawMessage, 4)

	s := collectSubscriber("updateUser", SubscriberOptions{}, got)
	cancel := r.register("updateUser", s)

	r.dispatch(context.Background(), "updateUser", json.RawMessage(`{"@type":"updateUser"}`))
	select {
	case raw := <-got:
		if string(raw) != `{"@type":"updateUser"}` {
			t.Fatalf("unexpected raw: %s", raw)
		}
	case <-time.After(time.Second):
		t.Fatal("subscriber did not receive update")
	}

	// Other types are not delivered.
	r.dispatch(context.Background(), "updateChat", json.RawMessage(`{"@type":"updateChat"}`))
	select {
	case raw := <-got:
		t.Fatalf("delivered wrong type: %s", raw)
	case <-time.After(50 * time.Millisecond):
	}

	cancel()
	cancel() // idempotent
	r.dispatch(context.Background(), "updateUser", json.RawMessage(`{"@type":"updateUser"}`))
	select {
	case raw := <-got:
		t.Fatalf("delivered after cancel: %s", raw)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestRouterDispatchDoesNotRetainCallerBuffer(t *testing.T) {
	r := newRouter()
	got := make(chan json.RawMessage, 1)
	s := collectSubscriber("updateX", SubscriberOptions{}, got)
	defer r.register("updateX", s)()

	raw := []byte(`{"@type":"updateX","v":1}`)
	r.dispatch(context.Background(), "updateX", raw)
	for i := range raw {
		raw[i] = ' '
	}

	select {
	case delivered := <-got:
		if string(delivered) != `{"@type":"updateX","v":1}` {
			t.Fatalf("router retained caller's buffer: %s", delivered)
		}
	case <-time.After(time.Second):
		t.Fatal("no delivery")
	}
}

// blockedHandler returns a handler that signals when it starts processing
// and then blocks until release is closed, letting tests fill the mailbox.
func blockedHandler() (HandlerFunc, <-chan struct{}, chan<- struct{}) {
	started := make(chan struct{})
	release := make(chan struct{})
	var once atomic.Int32
	return func(ctx context.Context, raw json.RawMessage) {
		if once.CompareAndSwap(0, 1) {
			close(started)
		}
		<-release
	}, started, release
}

func TestRouterDropNewest(t *testing.T) {
	r := newRouter()
	var droppedTotal atomic.Int64
	var drops atomic.Int64

	h, started, release := blockedHandler()
	s := newSubscriber("updateX", h, SubscriberOptions{
		MailboxSize: 1,
		Policy:      DropNewest, // also the zero-value default
		OnDrop: func(dropped int64, tdType string) {
			drops.Add(1)
			droppedTotal.Store(dropped)
			if tdType != "updateX" {
				t.Errorf("wrong tdType in warning: %q", tdType)
			}
		},
	})
	defer r.register("updateX", s)()
	defer close(release)

	raw := json.RawMessage(`{"@type":"updateX"}`)
	r.dispatch(context.Background(), "updateX", raw) // taken by handler, blocks
	<-started
	r.dispatch(context.Background(), "updateX", raw) // fills mailbox
	r.dispatch(context.Background(), "updateX", raw) // dropped
	r.dispatch(context.Background(), "updateX", raw) // dropped

	// The dispatcher returned without blocking (we'd deadlock otherwise:
	// the handler is still blocked on release).
	if got := drops.Load(); got != 2 {
		t.Fatalf("expected 2 drops, got %d", got)
	}
	if got := droppedTotal.Load(); got != 2 {
		t.Fatalf("expected cumulative drop count 2 in warning, got %d", got)
	}
}

func TestRouterDropOldest(t *testing.T) {
	r := newRouter()
	handled := make(chan json.RawMessage, 4)

	var once atomic.Int32
	started := make(chan struct{})
	release := make(chan struct{})
	h := func(ctx context.Context, raw json.RawMessage) {
		if once.CompareAndSwap(0, 1) {
			close(started)
			<-release
		}
		handled <- append(json.RawMessage(nil), raw...)
	}

	var drops atomic.Int64
	s := newSubscriber("updateX", h, SubscriberOptions{
		MailboxSize: 1,
		Policy:      DropOldest,
		OnDrop:      func(dropped int64, tdType string) { drops.Add(1) },
	})
	defer r.register("updateX", s)()

	r.dispatch(context.Background(), "updateX", json.RawMessage(`{"v":"first"}`)) // handler blocks on it
	<-started
	r.dispatch(context.Background(), "updateX", json.RawMessage(`{"v":"second"}`)) // mailbox
	r.dispatch(context.Background(), "updateX", json.RawMessage(`{"v":"third"}`))  // evicts "second"

	close(release)

	first := <-handled
	if string(first) != `{"v":"first"}` {
		t.Fatalf("unexpected first handled: %s", first)
	}
	second := <-handled
	if string(second) != `{"v":"third"}` {
		t.Fatalf("DropOldest kept the wrong item: %s", second)
	}
	if got := drops.Load(); got != 1 {
		t.Fatalf("expected 1 drop, got %d", got)
	}
}

func TestRouterBlockPolicy(t *testing.T) {
	r := newRouter()
	h, started, release := blockedHandler()
	s := newSubscriber("updateX", h, SubscriberOptions{MailboxSize: 1, Policy: Block})
	defer r.register("updateX", s)()

	raw := json.RawMessage(`{"@type":"updateX"}`)
	r.dispatch(context.Background(), "updateX", raw) // handler blocks on it
	<-started
	r.dispatch(context.Background(), "updateX", raw) // mailbox full now

	blocked := make(chan struct{})
	go func() {
		defer close(blocked)
		r.dispatch(context.Background(), "updateX", raw) // must block
	}()

	select {
	case <-blocked:
		t.Fatal("Block policy dispatch returned while mailbox was full")
	case <-time.After(50 * time.Millisecond):
	}

	close(release) // drain the mailbox
	select {
	case <-blocked:
	case <-time.After(time.Second):
		t.Fatal("Block policy dispatch did not proceed after space freed")
	}
}

func TestRouterBlockPolicyUnblockedByCancel(t *testing.T) {
	r := newRouter()
	h, started, _ := blockedHandler() // release never closed
	s := newSubscriber("updateX", h, SubscriberOptions{MailboxSize: 1, Policy: Block})
	cancel := r.register("updateX", s)

	raw := json.RawMessage(`{"@type":"updateX"}`)
	r.dispatch(context.Background(), "updateX", raw)
	<-started
	r.dispatch(context.Background(), "updateX", raw) // mailbox full

	blocked := make(chan struct{})
	go func() {
		defer close(blocked)
		r.dispatch(context.Background(), "updateX", raw)
	}()

	time.Sleep(20 * time.Millisecond)
	cancel() // must unblock the stuck deliver

	select {
	case <-blocked:
	case <-time.After(time.Second):
		t.Fatal("cancel did not unblock a Block-policy dispatch")
	}
}

func TestRouterCloseAll(t *testing.T) {
	r := newRouter()
	got1 := make(chan json.RawMessage, 1)
	got2 := make(chan json.RawMessage, 1)
	r.register("updateA", collectSubscriber("updateA", SubscriberOptions{}, got1))
	r.register("updateB", collectSubscriber("updateB", SubscriberOptions{}, got2))

	r.closeAll()

	r.dispatch(context.Background(), "updateA", json.RawMessage(`{}`))
	r.dispatch(context.Background(), "updateB", json.RawMessage(`{}`))
	select {
	case <-got1:
		t.Fatal("delivered after closeAll")
	case <-got2:
		t.Fatal("delivered after closeAll")
	case <-time.After(50 * time.Millisecond):
	}
}
