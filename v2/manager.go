package tdlib

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
)

// Manager owns the single process-wide receive loop and routes each raw
// TDLib message to the client named by its @client_id field. It never
// spawns a goroutine per update: delivery into a client goes through the
// client's bounded inbox (see Client.enqueue).
//
// The loop is driven by an injected FetchFunc pump, so the routing logic is
// fully testable without cgo. The tdjson-backed wiring (NewClient, and a
// tdjson.Receive pump) lives in manager_tdjson.go behind the `tdjson` build
// tag.
type Manager struct {
	mu      sync.RWMutex
	clients map[int]*Client

	running atomic.Bool

	// OnUnknownClient is invoked when a message arrives for a client ID that
	// is not registered; the message is dropped. May be nil.
	OnUnknownClient func(clientID int)

	// OnOrphanMessage is invoked when a message has no @client_id field (or
	// is malformed); the message is dropped. May be nil.
	OnOrphanMessage func(raw []byte)

	// OnClientInboxDrop is installed on every client registered afterwards
	// and fires when that client's bounded inbox drops its oldest message.
	// May be nil.
	OnClientInboxDrop func(clientID int, dropped int64)
}

// NewManager creates an empty manager.
func NewManager() *Manager {
	return &Manager{clients: make(map[int]*Client)}
}

// Register adds c to the routing table, keyed by c.ID().
func (m *Manager) Register(c *Client) {
	m.mu.Lock()
	c.onInboxDrop = m.OnClientInboxDrop
	m.clients[c.id] = c
	m.mu.Unlock()
}

// Unregister removes a client from the routing table. Messages that arrive
// afterwards for its ID hit OnUnknownClient and are dropped.
func (m *Manager) Unregister(clientID int) {
	m.mu.Lock()
	delete(m.clients, clientID)
	m.mu.Unlock()
}

// Run is the receive loop: pull from fetch, route by @client_id, repeat
// until ctx ends or fetch fails. Only one loop may run per manager.
func (m *Manager) Run(ctx context.Context, fetch FetchFunc) error {
	if !m.running.CompareAndSwap(false, true) {
		return ErrManagerRunning
	}
	defer m.running.Store(false)

	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		raw, err := fetch(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return err
		}
		if len(raw) == 0 {
			continue
		}
		m.route(raw)
	}
}

// route peeks @client_id and hands the raw message to that client's inbox.
// Unknown IDs and client-less messages are dropped through the hooks;
// neither case panics or blocks.
func (m *Manager) route(raw []byte) {
	var head struct {
		ClientID *int `json:"@client_id"`
	}
	if err := json.Unmarshal(raw, &head); err != nil || head.ClientID == nil {
		if m.OnOrphanMessage != nil {
			m.OnOrphanMessage(raw)
		}
		return
	}
	m.mu.RLock()
	c, ok := m.clients[*head.ClientID]
	m.mu.RUnlock()
	if !ok {
		if m.OnUnknownClient != nil {
			m.OnUnknownClient(*head.ClientID)
		}
		return
	}
	c.enqueue(raw)
}
