package tdlib

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aliforever/go-tdlib/v2/td"
)

// newMessageUpdate builds a raw updateNewMessage envelope. contentJSON is
// the content member's JSON, senderJSON the sender_id member's.
func newMessageUpdate(chatID, msgID int64, senderJSON, contentJSON string) []byte {
	return []byte(fmt.Sprintf(`{"@type":"updateNewMessage","message":{"@type":"message","id":%d,"chat_id":%d,"sender_id":%s,"content":%s}}`,
		msgID, chatID, senderJSON, contentJSON))
}

func textContent(text string) string {
	return fmt.Sprintf(`{"@type":"messageText","text":{"@type":"formattedText","text":%q}}`, text)
}

func senderUser(userID int64) string {
	return fmt.Sprintf(`{"@type":"messageSenderUser","user_id":%d}`, userID)
}

func senderChat(chatID int64) string {
	return fmt.Sprintf(`{"@type":"messageSenderChat","chat_id":%d}`, chatID)
}

func recvWithin[T any](t *testing.T, ch <-chan T, d time.Duration) T {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(d):
		t.Fatal("timed out waiting for delivery")
		var zero T
		return zero
	}
}

func routerSubCount(c *Client) int {
	c.router.mu.RLock()
	defer c.router.mu.RUnlock()
	return len(c.router.subs)
}

func TestFilterCombinators(t *testing.T) {
	isText := Filter[string](func(s string) bool { return s == "text" })
	isShort := Filter[string](func(s string) bool { return len(s) < 10 })

	if !isText.And(isShort)("text") {
		t.Fatal("And: expected true")
	}
	if isText.And(isShort)("photo") {
		t.Fatal("And: expected false")
	}
	if !isText.Or(isShort)("short") {
		t.Fatal("Or: expected true via isShort")
	}
	if !isText.Or(isShort)("text") {
		t.Fatal("Or: expected true via isText")
	}
	if isText.Or(isShort)("a very long string that is not text") {
		t.Fatal("Or: expected false")
	}
	if isText.Not()("text") {
		t.Fatal("Not: expected false")
	}
	if !isText.Not()("photo") {
		t.Fatal("Not: expected true")
	}
}

func TestOnReceivesDecodedUpdate(t *testing.T) {
	ft := &fakeTransport{}
	c := NewClient(30, ft)
	defer c.Close()

	got := make(chan *td.UpdateNewMessage, 1)
	sub := c.On[*td.UpdateNewMessage](func(ctx context.Context, u *td.UpdateNewMessage) {
		got <- u
	})
	defer sub.Close()

	c.enqueue(newMessageUpdate(100, 7, senderUser(5), textContent("hi")))

	u := recvWithin(t, got, 2*time.Second)
	if u.Message == nil || u.Message.Id != 7 || u.Message.ChatId != 100 {
		t.Fatalf("unexpected message: %+v", u.Message)
	}
	su, ok := u.Message.SenderId.As[*td.MessageSenderUser]()
	if !ok || su.UserId != 5 {
		t.Fatalf("sender not decoded as messageSenderUser(5): %v", u.Message.SenderId.Kind())
	}
	if u.Message.Content.Kind() != "messageText" {
		t.Fatalf("content kind = %q", u.Message.Content.Kind())
	}
}

func TestOnFiltersGateDelivery(t *testing.T) {
	ft := &fakeTransport{}
	c := NewClient(31, ft)
	defer c.Close()

	got := make(chan *td.UpdateNewMessage, 2)
	onlyChat100 := Filter[*td.UpdateNewMessage](func(u *td.UpdateNewMessage) bool {
		return u.Message != nil && u.Message.ChatId == 100
	})
	sub := c.On[*td.UpdateNewMessage](func(ctx context.Context, u *td.UpdateNewMessage) {
		got <- u
	}, onlyChat100)
	defer sub.Close()

	// FIFO delivery through the single dispatch goroutine and mailbox makes
	// this deterministic: if the chat-100 message arrives, the chat-200 one
	// was already processed (and filtered out).
	c.enqueue(newMessageUpdate(200, 1, senderUser(5), textContent("nope")))
	c.enqueue(newMessageUpdate(100, 2, senderUser(5), textContent("yes")))

	u := recvWithin(t, got, 2*time.Second)
	if u.Message.Id != 2 {
		t.Fatalf("filter passed the wrong message: id=%d", u.Message.Id)
	}
}

func TestOnCloseStopsDelivery(t *testing.T) {
	ft := &fakeTransport{}
	c := NewClient(32, ft)
	defer c.Close()

	got := make(chan *td.UpdateNewMessage, 1)
	sub := c.On[*td.UpdateNewMessage](func(ctx context.Context, u *td.UpdateNewMessage) {
		got <- u
	})
	sub.Close()
	sub.Close() // idempotent

	if n := routerSubCount(c); n != 0 {
		t.Fatalf("router still has %d subscriptions after Close", n)
	}
	c.enqueue(newMessageUpdate(100, 1, senderUser(5), textContent("hi")))
	select {
	case u := <-got:
		t.Fatalf("delivered after Close: id=%d", u.Message.Id)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestOnMalformedUpdateDropped(t *testing.T) {
	ft := &fakeTransport{}
	c := NewClient(33, ft)
	defer c.Close()

	got := make(chan *td.UpdateNewMessage, 1)
	sub := c.On[*td.UpdateNewMessage](func(ctx context.Context, u *td.UpdateNewMessage) {
		got <- u
	})
	defer sub.Close()

	c.enqueue([]byte(`{"@type":"updateNewMessage","message":"garbage"`))  // malformed JSON
	c.enqueue([]byte(`{"@type":"updateNewMessage","message":"garbage"}`)) // malformed value
	c.enqueue(newMessageUpdate(100, 9, senderUser(5), textContent("ok")))

	u := recvWithin(t, got, 2*time.Second)
	if u.Message.Id != 9 {
		t.Fatalf("expected the valid update, got id=%d", u.Message.Id)
	}
}

func TestUpdatesChannel(t *testing.T) {
	ft := &fakeTransport{}
	c := NewClient(34, ft)
	defer c.Close()

	ctx := context.Background()
	ch, cancel := c.Updates[*td.UpdateNewMessage](ctx, WithMailboxSize(4))

	c.enqueue(newMessageUpdate(100, 1, senderUser(5), textContent("one")))
	c.enqueue(newMessageUpdate(100, 2, senderUser(5), textContent("two")))

	first := recvWithin(t, ch, 2*time.Second)
	if first.Message.Id != 1 {
		t.Fatalf("first update id = %d", first.Message.Id)
	}
	second := recvWithin(t, ch, 2*time.Second)
	if second.Message.Id != 2 {
		t.Fatalf("second update id = %d", second.Message.Id)
	}

	cancel()
	cancel() // idempotent

	// The channel must close once the subscriber drains.
	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("received value after cancel, want closed channel")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("channel not closed after cancel")
	}
	if n := routerSubCount(c); n != 0 {
		t.Fatalf("router still has %d subscriptions after cancel", n)
	}
}

func TestUpdatesContextCancel(t *testing.T) {
	ft := &fakeTransport{}
	c := NewClient(35, ft)
	defer c.Close()

	ctx, cancel := context.WithCancel(context.Background())
	ch, _ := c.Updates[*td.UpdateNewMessage](ctx)

	cancel()

	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("received value after ctx cancel, want closed channel")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("channel not closed after ctx cancel")
	}
	deadline := time.Now().Add(2 * time.Second)
	for routerSubCount(c) != 0 {
		if time.Now().After(deadline) {
			t.Fatal("router entry leaked after ctx cancel")
		}
		time.Sleep(time.Millisecond)
	}
}

func TestNextOneShot(t *testing.T) {
	ft := &fakeTransport{}
	c := NewClient(36, ft)
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	type result struct {
		u   *td.UpdateNewMessage
		err error
	}
	done := make(chan result, 1)
	go func() {
		u, err := c.Next[*td.UpdateNewMessage](ctx, func(u *td.UpdateNewMessage) bool {
			return u.Message != nil && u.Message.ChatId == 100
		})
		done <- result{u, err}
	}()

	// Let Next register before dispatching.
	deadline := time.Now().Add(2 * time.Second)
	for routerSubCount(c) == 0 {
		if time.Now().After(deadline) {
			t.Fatal("Next did not register")
		}
		time.Sleep(time.Millisecond)
	}

	// Non-matching first: Next must keep waiting.
	c.enqueue(newMessageUpdate(200, 1, senderUser(5), textContent("skip")))
	c.enqueue(newMessageUpdate(100, 2, senderUser(5), textContent("hit")))
	// A second matching message must not disturb the one-shot result.
	c.enqueue(newMessageUpdate(100, 3, senderUser(5), textContent("late")))

	r := recvWithin(t, done, 2*time.Second)
	if r.err != nil {
		t.Fatalf("Next: %v", r.err)
	}
	if r.u.Message.Id != 2 {
		t.Fatalf("Next returned id=%d, want the first match (2)", r.u.Message.Id)
	}

	// Next must have unregistered itself.
	deadline = time.Now().Add(2 * time.Second)
	for routerSubCount(c) != 0 {
		if time.Now().After(deadline) {
			t.Fatal("router entry leaked after Next returned")
		}
		time.Sleep(time.Millisecond)
	}
}

func TestNextContextCancel(t *testing.T) {
	ft := &fakeTransport{}
	c := NewClient(37, ft)
	defer c.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.Next[*td.UpdateNewMessage](ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if n := routerSubCount(c); n != 0 {
		t.Fatalf("router entry leaked after cancelled Next: %d", n)
	}
}

func TestNextClientClose(t *testing.T) {
	ft := &fakeTransport{}
	c := NewClient(38, ft)

	errCh := make(chan error, 1)
	go func() {
		_, err := c.Next[*td.UpdateNewMessage](context.Background())
		errCh <- err
	}()

	deadline := time.Now().Add(2 * time.Second)
	for routerSubCount(c) == 0 {
		if time.Now().After(deadline) {
			t.Fatal("Next did not register")
		}
		time.Sleep(time.Millisecond)
	}
	c.Close()

	if err := recvWithin(t, errCh, 2*time.Second); !errors.Is(err, ErrClientClosed) {
		t.Fatalf("expected ErrClientClosed, got %v", err)
	}
}
