package tdlib

import (
	"encoding/json"
	"sync"
)

// callResult is the single outcome of a pending call: either the raw
// response bytes or a failure (e.g. ErrClientClosed from failAll).
type callResult struct {
	raw json.RawMessage
	err error
}

// pendingCall is a one-shot rendezvous between the goroutine blocked in
// Client.roundTrip and the dispatch goroutine delivering the response.
type pendingCall struct {
	ch   chan callResult // buffered, cap 1: completion never blocks
	once sync.Once
}

func newPendingCall() *pendingCall {
	return &pendingCall{ch: make(chan callResult, 1)}
}

// complete delivers the result exactly once; later completions (a late
// response racing Close, a duplicate delivery) are silently dropped.
// Because ch is buffered, complete never blocks the dispatcher.
func (p *pendingCall) complete(r callResult) {
	p.once.Do(func() { p.ch <- r })
}

// pendingRegistry correlates in-flight requests by their @extra key. It
// replaces v1's responseQueue: entries are removed by the caller via
// `defer remove(extra)` on every exit path, so timeouts and cancellations
// cannot leak channels.
type pendingRegistry struct {
	mu    sync.Mutex
	calls map[string]*pendingCall
}

func newPendingRegistry() *pendingRegistry {
	return &pendingRegistry{calls: make(map[string]*pendingCall)}
}

// register installs a new pending call under extra. If extra is already
// registered (caller bug — extras must be unique), the existing entry is
// replaced.
func (r *pendingRegistry) register(extra string) *pendingCall {
	call := newPendingCall()
	r.mu.Lock()
	r.calls[extra] = call
	r.mu.Unlock()
	return call
}

// deliver hands a response to the waiting caller. raw is copied so the
// registry never retains the dispatcher's buffer. An absent entry (response
// arriving after a timeout/cancel removed it) is dropped without panic.
func (r *pendingRegistry) deliver(extra string, raw []byte) {
	r.mu.Lock()
	call, ok := r.calls[extra]
	r.mu.Unlock()
	if !ok {
		return
	}
	buf := append(json.RawMessage(nil), raw...)
	call.complete(callResult{raw: buf})
}

// failAll completes every registered call with err. Entries stay in the
// map; each caller's deferred remove cleans them up.
func (r *pendingRegistry) failAll(err error) {
	r.mu.Lock()
	calls := make([]*pendingCall, 0, len(r.calls))
	for _, call := range r.calls {
		calls = append(calls, call)
	}
	r.mu.Unlock()
	for _, call := range calls {
		call.complete(callResult{err: err})
	}
}

// remove deletes the entry for extra. It is idempotent.
func (r *pendingRegistry) remove(extra string) {
	r.mu.Lock()
	delete(r.calls, extra)
	r.mu.Unlock()
}

// len reports the number of in-flight calls. Used by tests.
func (r *pendingRegistry) len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.calls)
}
