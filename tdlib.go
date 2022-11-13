package tdlib

// #cgo linux CFLAGS: -I/usr/local/include
// #cgo linux LDFLAGS: -Wl,-rpath=/usr/local/lib -L/usr/local/lib -ltdjson
// #include <stdlib.h>
// #include <td/telegram/td_json_client.h>
// #include <td/telegram/td_log.h>
import "C"

import (
	"encoding/json"
	"fmt"
	"github.com/aliforever/go-tdlib/config"
	"github.com/aliforever/go-tdlib/incomingevents"
	"github.com/aliforever/go-tdlib/outgoingevents"
	"github.com/google/uuid"
	"math/rand"
	"sync"
	"time"
	"unsafe"
)

type TDLib struct {
	apiID   int64
	apiHash string

	client unsafe.Pointer

	cfg config.Config

	handlers *Handlers

	responseQueueLocker sync.Mutex
	responseQueue       map[string]chan incomingevents.Event
}

func NewClient(apiID int64, apiHash string, handlers *Handlers, cfg *config.Config) *TDLib {
	rand.Seed(time.Now().Unix())

	// TODO: Add more fields to config
	if cfg == nil {
		cfg = &config.Config{
			UseFileDatabase:     true,
			UseChatInfoDatabase: true,
			UseMessageDatabase:  true,
		}
	}

	if cfg.SystemVersion == "" {
		cfg.SystemVersion = "1.0.0"
	}

	if cfg.SystemLanguageCode == "" {
		cfg.SystemLanguageCode = "en"
	}

	if cfg.DeviceModel == "" {
		cfg.DeviceModel = "Server"
	}

	if cfg.ApplicationVersion == "" {
		cfg.ApplicationVersion = "1.0.0"
	}

	if cfg.LogPath != "" {
		cfgBytes, _ := json.Marshal(map[string]interface{}{
			"@type": "setLogStream",
			"log_stream": map[string]interface{}{
				"@type":         "logStreamFile",
				"path":          cfg.LogPath,
				"max_file_size": 10485760,
			},
		})

		query := C.CString(string(cfgBytes))
		C.td_json_client_execute(nil, query)
		C.free(unsafe.Pointer(query))
	}

	return &TDLib{
		apiID:         apiID,
		apiHash:       apiHash,
		client:        C.td_json_client_create(),
		cfg:           *cfg,
		handlers:      handlers,
		responseQueue: map[string]chan incomingevents.Event{},
	}
}

func (t *TDLib) ReceiveUpdates() error {
	return t.receiveUpdates()
}

// TODO: Add send timeout
func (t *TDLib) send(data outgoingevents.EventInterface) (incomingevents.Event, error) {
	ch := make(chan incomingevents.Event)

	requestID := uuid.NewString()

	t.responseQueueLocker.Lock()
	t.responseQueue[requestID] = ch
	t.responseQueueLocker.Unlock()

	eventJS, err := outgoingevents.NewEventJSON(requestID, data)
	if err != nil {
		return incomingevents.Event{}, err
	}

	err = t.fireStringQuery(eventJS)
	if err != nil {
		return incomingevents.Event{}, err
	}

	resp := <-ch

	t.responseQueueLocker.Lock()
	close(ch)
	delete(t.responseQueue, requestID)
	t.responseQueueLocker.Unlock()

	if resp.Type == "error" {
		return incomingevents.Event{}, fmt.Errorf("%d: %s", resp.Code, resp.Message)
	}

	return resp, nil
}

func (t *TDLib) fireStringQuery(data string) error {
	query := C.CString(data)

	defer C.free(unsafe.Pointer(query))

	C.td_json_client_send(t.client, query)

	return nil
}

func (t *TDLib) receiveNextUpdate(timeout int64) []byte {
	return []byte(C.GoString(C.td_json_client_receive(t.client, C.double(timeout))))
}

func (t *TDLib) receiveUpdates() error {
	for {
		updateBytes := t.receiveNextUpdate(10)

		if updateBytes == nil || len(updateBytes) == 0 {
			continue
		}

		if t.handlers != nil && t.handlers.RawIncomingEvent != nil {
			go t.handlers.RawIncomingEvent(updateBytes)
		}

		ie, err := incomingevents.FromBytes(updateBytes)
		if err != nil {
			return err
		}

		if ie.RequestID != "" {
			if ch := t.getResponseQueueByRequestID(ie.RequestID); ch != nil {
				go func(responseChan chan incomingevents.Event, event incomingevents.Event) {
					responseChan <- event
				}(ch, ie)
				continue
			}
		}

		if t.handlers != nil && t.handlers.IncomingEvent != nil {
			go t.handlers.IncomingEvent(ie)
		}

		if t.handlers.OnUpdateConnectionState != nil && ie.Type == "updateConnectionState" && ie.State != nil {
			go t.handlers.OnUpdateConnectionState(ie.State.Type)
		}

		if t.handlers.OnUpdateAuthorizationState != nil && ie.Type == "updateAuthorizationState" && ie.AuthorizationState != nil {
			go t.handlers.OnUpdateAuthorizationState(ie.AuthorizationState.Type)
		}
	}
}

func (t *TDLib) getResponseQueueByRequestID(requestID string) chan incomingevents.Event {
	t.responseQueueLocker.Lock()
	defer t.responseQueueLocker.Unlock()

	if ch, ok := t.responseQueue[requestID]; ok {
		return ch
	}

	return nil
}
