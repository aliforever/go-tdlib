package tdlib

import "context"

// Transport is the non-generic seam below the typed API. Generic methods
// cannot be mocked through interfaces, so everything above this seam
// (pending registry, router, dispatch) is testable with a fake Transport.
//
// The production implementation wraps internal/tdjson behind the `tdjson`
// build tag (see manager_tdjson.go); tests supply their own.
type Transport interface {
	// Send enqueues one marshalled request for the given client. payload is
	// a complete TDLib JSON envelope including @type and @extra.
	Send(clientID int, payload []byte) error
}

// FetchFunc pulls the next raw message from the process-wide TDLib receive
// queue. It is the pump driving Manager.Run:
//
//   - return (nil, nil) on timeout to keep polling;
//   - return a non-nil error to stop the loop (a cancelled ctx ends the loop
//     with ctx.Err()).
//
// The production pump wraps tdjson.Receive behind the `tdjson` build tag;
// tests drive Manager.Run with a scripted pump.
type FetchFunc func(ctx context.Context) ([]byte, error)
