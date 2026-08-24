package tdlib

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"

	"github.com/aliforever/go-tdlib/v2/td"
)

// Filter gates delivery of a decoded update: the update reaches the handler
// only if every filter returns true. Filters run after JSON decode, before
// the handler, on the subscriber's own goroutine.
type Filter[U any] func(U) bool

// And returns a filter requiring both f and g.
func (f Filter[U]) And(g Filter[U]) Filter[U] {
	return func(u U) bool { return f(u) && g(u) }
}

// Or returns a filter requiring f or g.
func (f Filter[U]) Or(g Filter[U]) Filter[U] {
	return func(u U) bool { return f(u) || g(u) }
}

// Not returns a filter negating f.
func (f Filter[U]) Not() Filter[U] {
	return func(u U) bool { return !f(u) }
}

// Subscription is an active update registration. Close stops delivery and
// releases the subscriber's goroutine; it is idempotent.
type Subscription interface {
	Close()
}

type subscription struct{ cancel func() }

func (s subscription) Close() { s.cancel() }

// SubscriberOption configures the mailbox behind an Updates channel; see
// SubscriberOptions for the semantics of each knob.
type SubscriberOption func(*SubscriberOptions)

// WithMailboxSize bounds queued undelivered updates (<= 0 keeps the default).
func WithMailboxSize(n int) SubscriberOption {
	return func(o *SubscriberOptions) { o.MailboxSize = n }
}

// WithBackpressure selects what happens when the mailbox is full; the zero
// value is DropNewest.
func WithBackpressure(p Backpressure) SubscriberOption {
	return func(o *SubscriberOptions) { o.Policy = p }
}

// WithDropWarning installs a hook invoked when updates are dropped under
// DropNewest/DropOldest.
func WithDropWarning(f DropWarnFunc) SubscriberOption {
	return func(o *SubscriberOptions) { o.OnDrop = f }
}

func applySubscriberOptions(opts []SubscriberOption) SubscriberOptions {
	var cfg SubscriberOptions
	for _, o := range opts {
		if o != nil {
			o(&cfg)
		}
	}
	return cfg
}

// updateTDType resolves the @type discriminator of an update variant from
// its type alone. Generated TDType methods return a constant without
// dereferencing the receiver, so the nil pointer is safe.
func updateTDType[U td.UpdateVariant]() string {
	var z U
	return z.TDType()
}

// decodeUpdate decodes one raw update into U. U is always a pointer type
// (the sealed variant set is pointer-implemented), so new(U) is a pointer
// to pointer and json.Unmarshal allocates.
func decodeUpdate[U td.UpdateVariant](raw json.RawMessage) (U, error) {
	u := new(U)
	if err := json.Unmarshal(raw, u); err != nil {
		var zero U
		return zero, err
	}
	return *u, nil
}

func passesFilters[U any](v U, filters []Filter[U]) bool {
	for _, f := range filters {
		if f != nil && !f(v) {
			return false
		}
	}
	return true
}

// On subscribes handler to updates of variant U — the subscription key is
// U's @type, so On[*td.UpdateNewMessage] receives exactly new-message
// updates. The handler runs on the subscriber's own goroutine behind a
// bounded mailbox (default size, DropNewest), never on the client's
// dispatch goroutine. Malformed updates are dropped, not delivered.
func (c *Client) On[U td.UpdateVariant](handler func(context.Context, U), filters ...Filter[U]) Subscription {
	tdType := updateTDType[U]()
	h := func(ctx context.Context, raw json.RawMessage) {
		v, err := decodeUpdate[U](raw)
		if err != nil {
			return
		}
		if !passesFilters(v, filters) {
			return
		}
		handler(ctx, v)
	}
	cancel := c.router.register(tdType, newSubscriber(tdType, h, SubscriberOptions{}))
	return subscription{cancel: cancel}
}

// Updates is the channel form of On: decoded updates of variant U are
// delivered on the returned channel until cancel is called, ctx ends, or
// the client is closed. Cancel unregisters the subscriber and closes the
// channel once in-flight delivery drains; the channel is closed by the
// subscriber's own goroutine, so close never races a send.
func (c *Client) Updates[U td.UpdateVariant](ctx context.Context, opts ...SubscriberOption) (<-chan U, func()) {
	tdType := updateTDType[U]()
	s := newChanSubscriber[U](tdType, applySubscriberOptions(opts))
	cancel := c.router.register(tdType, s)
	s.start(ctx)

	// Unregister when the caller's context ends so a forgotten Updates does
	// not leak its router entry until client Close.
	go func() {
		select {
		case <-ctx.Done():
			cancel()
		case <-c.ctx.Done():
			cancel() // idempotent; Close already stopped the subscriber
		case <-s.done:
		}
	}()

	return s.out, cancel
}

// Next waits for the first update of variant U passing all filters, then
// unregisters. It returns ctx.Err() on cancellation and ErrClientClosed if
// the client closes first.
func (c *Client) Next[U td.UpdateVariant](ctx context.Context, filters ...Filter[U]) (U, error) {
	tdType := updateTDType[U]()
	match := make(chan U, 1)
	h := func(_ context.Context, raw json.RawMessage) {
		v, err := decodeUpdate[U](raw)
		if err != nil || !passesFilters(v, filters) {
			return
		}
		select {
		case match <- v:
		default: // a previous delivery already claimed the one-shot slot
		}
	}
	cancel := c.router.register(tdType, newSubscriber(tdType, h, SubscriberOptions{}))
	defer cancel()

	select {
	case <-ctx.Done():
		var zero U
		return zero, ctx.Err()
	case <-c.ctx.Done():
		var zero U
		return zero, ErrClientClosed
	case v := <-match:
		return v, nil
	}
}

// ---------------------------------------------------------------------------
// chanSubscriber: the erasedSubscriber behind Updates
// ---------------------------------------------------------------------------

// chanSubscriber mirrors subscriber's bounded-mailbox mechanics but decodes
// into U and forwards values to an exported channel. The loop goroutine is
// the only sender on out and closes it on exit, so cancel can never race a
// send against a close — unlike subscriber, where the handler runs user
// code that must not be trusted with channel ownership.
type chanSubscriber[U td.UpdateVariant] struct {
	tdType string
	policy Backpressure
	onDrop DropWarnFunc
	ctx    context.Context // Updates caller ctx; ends delivery early

	mailbox chan mailboxItem
	out     chan U
	done    chan struct{}
	stopped sync.Once
	dropped atomic.Int64
}

func newChanSubscriber[U td.UpdateVariant](tdType string, opts SubscriberOptions) *chanSubscriber[U] {
	if opts.MailboxSize <= 0 {
		opts.MailboxSize = defaultMailboxSize
	}
	s := &chanSubscriber[U]{
		tdType:  tdType,
		policy:  opts.Policy,
		onDrop:  opts.OnDrop,
		mailbox: make(chan mailboxItem, opts.MailboxSize),
		out:     make(chan U),
		done:    make(chan struct{}),
	}
	return s
}

// start binds the caller's ctx and launches the loop. It is separate from
// newChanSubscriber so Updates can register before the goroutine starts.
func (s *chanSubscriber[U]) start(ctx context.Context) {
	s.ctx = ctx
	go s.loop()
}

// deliver is called from the client's dispatch goroutine via
// router.dispatch; it only enqueues, per the backpressure policy.
func (s *chanSubscriber[U]) deliver(ctx context.Context, raw json.RawMessage) {
	enqueueMailbox(s.mailbox, s.done, s.policy, mailboxItem{ctx: ctx, raw: raw}, s.noteDrop)
}

func (s *chanSubscriber[U]) loop() {
	defer close(s.out)
	for {
		select {
		case <-s.done:
			return
		case item := <-s.mailbox:
			v, err := decodeUpdate[U](item.raw)
			if err != nil {
				continue
			}
			select {
			case s.out <- v:
			case <-s.done:
				return
			case <-s.ctx.Done():
				return
			}
		}
	}
}

func (s *chanSubscriber[U]) noteDrop() {
	n := s.dropped.Add(1)
	if s.onDrop != nil {
		s.onDrop(n, s.tdType)
	}
}

// stop terminates the loop goroutine (which then closes out) and unblocks
// deliver. Idempotent.
func (s *chanSubscriber[U]) stop() {
	s.stopped.Do(func() { close(s.done) })
}
