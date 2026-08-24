package main

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"
)

// updateEntry is one buffered TDLib update.
type updateEntry struct {
	Type       string
	Raw        json.RawMessage
	ReceivedAt time.Time
}

// matches reports whether the entry satisfies a waiter's filters.
func (e updateEntry) matches(typ, contains string) bool {
	if typ != "" && e.Type != typ {
		return false
	}
	if contains != "" && !strings.Contains(string(e.Raw), contains) {
		return false
	}
	return true
}

type updateWaiter struct {
	typ      string
	contains string
	ch       chan updateEntry // buffered(1); exactly one send, then removed
}

// updateBuffer is a fixed-capacity ring of recent updates plus a set of
// active waiters, fed by the client's single OnRaw hook. All methods are
// non-blocking for the caller except wait, which parks on a channel.
type updateBuffer struct {
	mu     sync.Mutex
	ring   []updateEntry
	next   int // write position
	count  int
	waiter []*updateWaiter

	authState string // last seen authorization_state.@type
}

func newUpdateBuffer(cap int) *updateBuffer {
	return &updateBuffer{ring: make([]updateEntry, cap)}
}

// add stores the update (copying raw) and releases any matching waiters.
// It runs on the client's dispatch goroutine and must not block.
func (b *updateBuffer) add(tdType string, raw json.RawMessage) {
	e := updateEntry{
		Type:       tdType,
		Raw:        append(json.RawMessage(nil), raw...),
		ReceivedAt: time.Now(),
	}
	b.mu.Lock()
	defer b.mu.Unlock()

	b.ring[b.next] = e
	b.next = (b.next + 1) % len(b.ring)
	if b.count < len(b.ring) {
		b.count++
	}

	if tdType == "updateAuthorizationState" {
		var u struct {
			State struct {
				Type string `json:"@type"`
			} `json:"authorization_state"`
		}
		if json.Unmarshal(e.Raw, &u) == nil && u.State.Type != "" {
			b.authState = u.State.Type
		}
	}

	keep := b.waiter[:0]
	for _, w := range b.waiter {
		if e.matches(w.typ, w.contains) {
			w.ch <- e // buffered, capacity 1: cannot block
			continue
		}
		keep = append(keep, w)
	}
	b.waiter = keep
}

// list returns up to limit most recent entries (oldest first) whose @type
// equals typeFilter (empty matches all).
func (b *updateBuffer) list(typeFilter string, limit int) []updateEntry {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]updateEntry, 0, min(limit, b.count))
	for i := 0; i < b.count; i++ {
		e := b.ring[(b.next-b.count+i+len(b.ring))%len(b.ring)]
		if typeFilter == "" || e.Type == typeFilter {
			out = append(out, e)
		}
	}
	if len(out) > limit {
		out = out[len(out)-limit:]
	}
	return out
}

// wait parks until an update matching the filters arrives or ctx is done.
// Concurrent waiters (same or different filters) are supported.
func (b *updateBuffer) wait(ctx context.Context, typeFilter, contains string) (updateEntry, error) {
	w := &updateWaiter{typ: typeFilter, contains: contains, ch: make(chan updateEntry, 1)}
	b.mu.Lock()
	b.waiter = append(b.waiter, w)
	b.mu.Unlock()

	select {
	case e := <-w.ch:
		return e, nil
	case <-ctx.Done():
		b.mu.Lock()
		for i, x := range b.waiter {
			if x == w {
				b.waiter = append(b.waiter[:i], b.waiter[i+1:]...)
				break
			}
		}
		b.mu.Unlock()
		return updateEntry{}, ctx.Err()
	}
}

// authStatus reports the last seen authorization state kind ("unknown" if
// none yet) and whether it is authorizationStateReady.
func (b *updateBuffer) authStatus() (state string, authorized bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.authState == "" {
		return "unknown", false
	}
	return b.authState, b.authState == "authorizationStateReady"
}
