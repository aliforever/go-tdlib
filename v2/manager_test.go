package tdlib

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"
)

// scriptedPump returns a FetchFunc fed through queue.
func scriptedPump(queue chan []byte) FetchFunc {
	return func(ctx context.Context) ([]byte, error) {
		select {
		case b := <-queue:
			return b, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(5 * time.Millisecond):
			return nil, nil // poll timeout
		}
	}
}

func runManager(t *testing.T, m *Manager, fetch FetchFunc) (context.Context, context.CancelFunc) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- m.Run(ctx, fetch) }()
	t.Cleanup(func() {
		cancel()
		if err := <-done; !errors.Is(err, context.Canceled) {
			t.Errorf("manager Run returned %v, want context.Canceled", err)
		}
	})
	return ctx, cancel
}

func TestManagerRoutesByClientID(t *testing.T) {
	m := NewManager()
	queue := make(chan []byte, 8)
	runManager(t, m, scriptedPump(queue))

	c1 := NewClient(1, &fakeTransport{})
	c2 := NewClient(2, &fakeTransport{})
	defer c1.Close()
	defer c2.Close()
	m.Register(c1)
	m.Register(c2)

	got1 := make(chan json.RawMessage, 1)
	got2 := make(chan json.RawMessage, 1)
	defer c1.router.register("updateUser", collectSubscriber("updateUser", SubscriberOptions{}, got1))()
	defer c2.router.register("updateUser", collectSubscriber("updateUser", SubscriberOptions{}, got2))()

	queue <- []byte(`{"@type":"updateUser","@client_id":2}`)

	select {
	case <-got2:
	case <-time.After(time.Second):
		t.Fatal("client 2 did not receive its update")
	}
	select {
	case raw := <-got1:
		t.Fatalf("client 1 received an update for client 2: %s", raw)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestManagerUnknownClientDropped(t *testing.T) {
	m := NewManager()
	queue := make(chan []byte, 8)
	runManager(t, m, scriptedPump(queue))

	unknown := make(chan int, 1)
	m.OnUnknownClient = func(id int) { unknown <- id }

	queue <- []byte(`{"@type":"updateUser","@client_id":999}`)

	select {
	case id := <-unknown:
		if id != 999 {
			t.Fatalf("wrong unknown client ID: %d", id)
		}
	case <-time.After(time.Second):
		t.Fatal("unknown client message was not reported")
	}
}

func TestManagerOrphanMessageDropped(t *testing.T) {
	m := NewManager()
	queue := make(chan []byte, 8)
	runManager(t, m, scriptedPump(queue))

	orphans := make(chan []byte, 2)
	m.OnOrphanMessage = func(raw []byte) { orphans <- raw }

	queue <- []byte(`{"@type":"updateUser"}`) // no @client_id
	queue <- []byte(`not json at all`)       // malformed

	for i := 0; i < 2; i++ {
		select {
		case <-orphans:
		case <-time.After(time.Second):
			t.Fatalf("orphan %d not reported", i)
		}
	}
}

func TestManagerInboxFullDropsOldest(t *testing.T) {
	m := NewManager()

	c := NewClient(7, &fakeTransport{})
	defer c.Close()

	dropped := make(chan int64, 1)
	m.OnClientInboxDrop = func(clientID int, n int64) {
		if clientID != 7 {
			t.Errorf("drop reported for wrong client: %d", clientID)
		}
		select {
		case dropped <- n:
		default:
		}
	}
	m.Register(c)

	// Stall the dispatch goroutine itself: a Block-policy subscriber with a
	// blocked handler, a full mailbox, and one in-flight deliver.
	h, started, release := blockedHandler()
	defer close(release)
	defer c.router.register("updateY", newSubscriber("updateY", h, SubscriberOptions{MailboxSize: 1, Policy: Block}))()
	m.route([]byte(`{"@type":"updateY","@client_id":7,"i":-2}`)) // handler blocks on it
	<-started
	m.route([]byte(`{"@type":"updateY","@client_id":7,"i":-1}`)) // fills mailbox
	m.route([]byte(`{"@type":"updateY","@client_id":7,"i":0}`))  // dispatch goroutine now blocked in deliver

	// Overflow the inbox while dispatch is stalled.
	for i := 1; i <= defaultInboxSize+10; i++ {
		m.route([]byte(fmt.Sprintf(`{"@type":"updateZ","@client_id":7,"i":%d}`, i)))
	}

	select {
	case n := <-dropped:
		if n == 0 {
			t.Fatal("drop counter never advanced")
		}
	case <-time.After(time.Second):
		t.Fatal("full inbox did not trigger the drop-oldest warning")
	}
}

func TestManagerRejectsSecondRun(t *testing.T) {
	m := NewManager()
	queue := make(chan []byte)
	runManager(t, m, scriptedPump(queue))

	// Wait until the first loop has actually claimed the manager; otherwise
	// the second Run could win the CompareAndSwap race.
	deadline := time.Now().Add(2 * time.Second)
	for !m.running.Load() {
		if time.Now().After(deadline) {
			t.Fatal("first receive loop did not start")
		}
		time.Sleep(time.Millisecond)
	}

	if err := m.Run(context.Background(), scriptedPump(queue)); !errors.Is(err, ErrManagerRunning) {
		t.Fatalf("expected ErrManagerRunning, got %v", err)
	}
}

// TestRoundTripThroughManager is the end-to-end path: roundTrip registers a
// pending call, the fake transport scripts TDLib's response onto the
// manager's pump, and the manager routes it back by @client_id and @extra.
func TestRoundTripThroughManager(t *testing.T) {
	m := NewManager()
	queue := make(chan []byte, 8)
	runManager(t, m, scriptedPump(queue))

	ft := &fakeTransport{}
	c := NewClient(11, ft)
	defer c.Close()
	m.Register(c)

	ft.onSend = func(clientID int, p []byte) {
		resp := fmt.Sprintf(`{"@type":"user","id":7,"@extra":%q,"@client_id":%d}`, extraOf(t, p), clientID)
		queue <- []byte(resp)
	}

	extra := NewExtra()
	payload := []byte(fmt.Sprintf(`{"@type":"getMe","@extra":%q}`, extra))

	raw, err := c.roundTrip(context.Background(), payload, extra)
	if err != nil {
		t.Fatalf("roundTrip: %v", err)
	}
	var user struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(raw, &user); err != nil || user.ID != 7 {
		t.Fatalf("unexpected response %s (err %v)", raw, err)
	}
	if c.pending.len() != 0 {
		t.Fatalf("pending leaked: %d", c.pending.len())
	}
}

func TestManagerUnregisterStopsRouting(t *testing.T) {
	m := NewManager()
	queue := make(chan []byte, 8)
	runManager(t, m, scriptedPump(queue))

	c := NewClient(21, &fakeTransport{})
	defer c.Close()
	m.Register(c)

	got := make(chan json.RawMessage, 1)
	defer c.router.register("updateUser", collectSubscriber("updateUser", SubscriberOptions{}, got))()

	m.Unregister(21)

	unknown := make(chan int, 1)
	m.OnUnknownClient = func(id int) { unknown <- id }

	queue <- []byte(`{"@type":"updateUser","@client_id":21}`)

	select {
	case <-unknown:
	case <-time.After(time.Second):
		t.Fatal("message for unregistered client was not reported as unknown")
	}
	select {
	case raw := <-got:
		t.Fatalf("unregistered client still received updates: %s", raw)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestEnqueueToClosedClientDoesNotPanic(t *testing.T) {
	c := NewClient(31, &fakeTransport{})
	c.Close()
	c.enqueue([]byte(`{"@type":"updateUser"}`)) // must be silently discarded
}
