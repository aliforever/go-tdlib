//go:build cgo && tdjson

package tdlib

import (
	"context"

	"github.com/aliforever/go-tdlib/v2/internal/tdjson"
)

// tdjsonTransport adapts the tdjson cgo edge to the Transport seam.
// td_send cannot fail synchronously, so Send always returns nil.
type tdjsonTransport struct{}

func (tdjsonTransport) Send(clientID int, payload []byte) error {
	tdjson.Send(clientID, payload)
	return nil
}

// NewClient allocates a TDLib client ID via libtdjson, creates the client
// and registers it for routing.
func (m *Manager) NewClient() *Client {
	c := NewClient(tdjson.CreateClientID(), tdjsonTransport{})
	m.Register(c)

	// TDLib does not emit the initial updateAuthorizationState until the
	// client's first request; kick it so authorization subscribers see the
	// state machine start. The response carries no @extra and is harmlessly
	// dispatched as an unmatched update.
	tdjson.Send(c.ID(), []byte(`{"@type":"getAuthorizationState"}`))

	return c
}

// RunTDJSON runs the manager receive loop on the real tdjson.Receive pump.
// The 1s receive timeout bounds how often the loop re-checks ctx.
func (m *Manager) RunTDJSON(ctx context.Context) error {
	return m.Run(ctx, func(ctx context.Context) ([]byte, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		return tdjson.Receive(1.0), nil
	})
}

// Execute runs a synchronous TDLib request (td_execute). Only log stream
// configuration and a few other requests may be executed this way.
func Execute(payload []byte) []byte {
	return tdjson.Execute(payload)
}

// SetLogVerbosity registers the TDLib log callback at the given maximum
// verbosity level. Effective only once per process.
func SetLogVerbosity(level int) {
	tdjson.SetLogVerbosity(level)
}

// SetLogHandler installs the handler for TDLib log messages.
func SetLogHandler(h func(verbosityLevel int, message string)) {
	tdjson.SetLogHandler(h)
}
