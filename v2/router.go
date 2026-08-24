package tdlib

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
)

// defaultMailboxSize bounds each subscriber's queue of undelivered updates.
const defaultMailboxSize = 64

// Backpressure selects what happens when a subscriber's mailbox is full.
type Backpressure int

const (
	// DropNewest drops the incoming update when the mailbox is full. This is
	// the default: the dispatcher never blocks on a slow subscriber.
	DropNewest Backpressure = iota
	// DropOldest evicts the oldest queued update to make room for the new
	// one. The dispatcher never blocks.
	DropOldest
	// Block waits until the mailbox has room, the subscriber is cancelled,
	// or the dispatch context ends. This can stall the owning client's
	// dispatch goroutine (but never the manager receive loop, which is
	// decoupled by the client's bounded inbox).
	Block
)

// DropWarnFunc is called when a subscriber drops updates. dropped is the
// cumulative number of updates this subscriber has dropped so far.
type DropWarnFunc func(dropped int64, tdType string)

// SubscriberOptions configures a subscriber's mailbox.
type SubscriberOptions struct {
	// MailboxSize bounds queued updates; <= 0 uses defaultMailboxSize.
	MailboxSize int
	// Policy selects the backpressure behaviour; zero value is DropNewest.
	Policy Backpressure
	// OnDrop is invoked on drops under DropNewest/DropOldest. May be nil.
	OnDrop DropWarnFunc
}

// HandlerFunc processes one raw update. raw is a private copy owned by the
// subscriber for the duration of the call: it may be retained, but must not
// be mutated.
type HandlerFunc func(ctx context.Context, raw json.RawMessage)

// erasedSubscriber is the type-erased machinery Phase 3's generic On[U]
// builds on: it receives raw JSON with no knowledge of the decoded type.
type erasedSubscriber interface {
	deliver(ctx context.Context, raw json.RawMessage)
	stop()
}

// mailboxItem carries the dispatch context alongside the raw update so the
// handler observes the cancellation state at enqueue time.
type mailboxItem struct {
	ctx context.Context
	raw json.RawMessage
}

// subscriber is a HandlerFunc behind a bounded mailbox drained by its own
// goroutine, so dispatch never executes user handler code inline.
type subscriber struct {
	tdType  string
	handler HandlerFunc
	policy  Backpressure
	onDrop  DropWarnFunc

	mailbox chan mailboxItem
	done    chan struct{}
	stopped sync.Once
	dropped atomic.Int64
}

func newSubscriber(tdType string, h HandlerFunc, opts SubscriberOptions) *subscriber {
	if opts.MailboxSize <= 0 {
		opts.MailboxSize = defaultMailboxSize
	}
	s := &subscriber{
		tdType:  tdType,
		handler: h,
		policy:  opts.Policy,
		onDrop:  opts.OnDrop,
		mailbox: make(chan mailboxItem, opts.MailboxSize),
		done:    make(chan struct{}),
	}
	go s.loop()
	return s
}

func (s *subscriber) loop() {
	for {
		select {
		case <-s.done:
			return
		case item := <-s.mailbox:
			s.handler(item.ctx, item.raw)
		}
	}
}

// deliver enqueues an update according to the backpressure policy. It is
// called from the client's dispatch goroutine via router.dispatch.
func (s *subscriber) deliver(ctx context.Context, raw json.RawMessage) {
	enqueueMailbox(s.mailbox, s.done, s.policy, mailboxItem{ctx: ctx, raw: raw}, s.noteDrop)
}

// enqueueMailbox applies the backpressure policy to one item. It is shared
// by subscriber (raw handler) and chanSubscriber (typed Updates channel);
// noteDrop is invoked on each drop under DropNewest/DropOldest.
func enqueueMailbox(mailbox chan mailboxItem, done chan struct{}, policy Backpressure, item mailboxItem, noteDrop func()) {
	switch policy {
	case DropOldest:
		select {
		case mailbox <- item:
			return
		default:
		}
		select {
		case <-mailbox:
			noteDrop()
		default:
		}
		select {
		case mailbox <- item:
		case <-done:
		}
	case Block:
		select {
		case mailbox <- item:
		case <-done:
		case <-item.ctx.Done():
		}
	default: // DropNewest
		select {
		case mailbox <- item:
		case <-done:
		default:
			noteDrop()
		}
	}
}

func (s *subscriber) noteDrop() {
	n := s.dropped.Add(1)
	if s.onDrop != nil {
		s.onDrop(n, s.tdType)
	}
}

// stop terminates the mailbox goroutine and unblocks deliver. Idempotent.
func (s *subscriber) stop() {
	s.stopped.Do(func() { close(s.done) })
}

// router maps TDLib @type strings to update subscribers.
type router struct {
	mu   sync.RWMutex
	subs map[string][]erasedSubscriber
}

func newRouter() *router {
	return &router{subs: make(map[string][]erasedSubscriber)}
}

// register subscribes s to updates of the given @type and returns a cancel
// function. Cancel is idempotent: it unsubscribes and stops the
// subscriber's mailbox goroutine.
func (r *router) register(tdType string, s erasedSubscriber) (cancel func()) {
	r.mu.Lock()
	r.subs[tdType] = append(r.subs[tdType], s)
	r.mu.Unlock()

	var once sync.Once
	return func() {
		once.Do(func() {
			r.mu.Lock()
			list := r.subs[tdType]
			for i, e := range list {
				if e == s {
					list = append(list[:i:i], list[i+1:]...)
					break
				}
			}
			if len(list) == 0 {
				delete(r.subs, tdType)
			} else {
				r.subs[tdType] = list
			}
			r.mu.Unlock()
			s.stop()
		})
	}
}

// dispatch delivers one update to all subscribers of tdType. The raw buffer
// is copied once so the router never retains the caller's slice; all
// subscribers share that single read-only copy.
func (r *router) dispatch(ctx context.Context, tdType string, raw json.RawMessage) {
	r.mu.RLock()
	list := append([]erasedSubscriber(nil), r.subs[tdType]...)
	r.mu.RUnlock()
	if len(list) == 0 {
		return
	}
	buf := append(json.RawMessage(nil), raw...)
	for _, s := range list {
		s.deliver(ctx, buf)
	}
}

// closeAll stops every subscriber and clears the registry. Called from
// Client.Close.
func (r *router) closeAll() {
	r.mu.Lock()
	subs := r.subs
	r.subs = make(map[string][]erasedSubscriber)
	r.mu.Unlock()
	for _, list := range subs {
		for _, s := range list {
			s.stop()
		}
	}
}
