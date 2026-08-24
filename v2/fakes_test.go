package tdlib

import (
	"encoding/json"
	"sync"
)

// fakeTransport records sent payloads and lets a test script reactions to
// each send (e.g. deliver a scripted response into the client's registry).
type fakeTransport struct {
	mu      sync.Mutex
	sent    []sentPayload
	onSend  func(clientID int, payload []byte)
	sendErr error
}

type sentPayload struct {
	clientID int
	payload  []byte
}

func (f *fakeTransport) Send(clientID int, payload []byte) error {
	f.mu.Lock()
	f.sent = append(f.sent, sentPayload{clientID: clientID, payload: append([]byte(nil), payload...)})
	onSend := f.onSend
	sendErr := f.sendErr
	f.mu.Unlock()
	if onSend != nil {
		onSend(clientID, payload)
	}
	return sendErr
}

func (f *fakeTransport) sentCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.sent)
}

// extraOf extracts the @extra key from a marshalled request envelope.
func extraOf(t interface{ Fatalf(string, ...any) }, payload []byte) string {
	var head struct {
		Extra string `json:"@extra"`
	}
	if err := json.Unmarshal(payload, &head); err != nil {
		t.Fatalf("extract @extra: %v", err)
	}
	return head.Extra
}
