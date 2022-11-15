package tdlib

// #cgo linux CFLAGS: -I/usr/local/include
// #cgo linux LDFLAGS: -Wl,-rpath=/usr/local/lib -ltdjson
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
	responseQueue       map[string]chan []byte
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
		responseQueue: map[string]chan []byte{},
	}
}

func (t *TDLib) ReceiveUpdates() error {
	return t.receiveUpdates()
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

		if t.handlers != nil && t.handlers.rawIncomingEvent != nil {
			go t.handlers.rawIncomingEvent(updateBytes)
		}

		ie, err := incomingevents.GenericFromBytes(updateBytes)
		if err != nil {
			return err
		}

		if ie.RequestID != "" {
			if ch := t.getResponseQueueByRequestID(ie.RequestID); ch != nil {
				go func(responseChan chan []byte, event []byte) {
					responseChan <- event
				}(ch, updateBytes)
				continue
			}
		}

		event, err := incomingevents.FromBytes(updateBytes)
		if err != nil {
			return err
		}

		if t.handlers != nil && t.handlers.incomingEvent != nil {
			go t.handlers.incomingEvent(event)
		}

		if t.handlers.onUpdateConnectionState != nil && event.Type == "updateConnectionState" && event.State != nil {
			go t.handlers.onUpdateConnectionState(event.State.Type)
		}

		if t.handlers.onUpdateAuthorizationState != nil && event.Type == "updateAuthorizationState" && event.AuthorizationState != nil {
			go t.handlers.onUpdateAuthorizationState(event.AuthorizationState.Type)
		}

		t.handlers.eventTypeHandlerLocker.Lock()
		if handler, ok := t.handlers.eventTypeHandlers[event.Type]; ok && handler != nil {
			err := handler.Handle(updateBytes)
			if err != nil {
				return err
			}
		}
	}
}

func (t *TDLib) getResponseQueueByRequestID(requestID string) chan []byte {
	t.responseQueueLocker.Lock()
	defer t.responseQueueLocker.Unlock()

	if ch, ok := t.responseQueue[requestID]; ok {
		return ch
	}

	return nil
}

func sendMap[ResponseType any](t *TDLib, requestType string, data map[string]interface{}) (ResponseType, error) {
	requestID := uuid.NewString()

	eventJS, err := outgoingevents.NewEventJSONFromMap(requestID, requestType, data)
	if err != nil {
		return ResponseType(nil), err
	}

	return _send[ResponseType](t, requestID, eventJS)
}

// TODO: Add timeout
func send[ResponseType any](t *TDLib, data outgoingevents.EventInterface) (ResponseType, error) {
	requestID := uuid.NewString()

	eventJS, err := outgoingevents.NewEventJSON(requestID, data)
	if err != nil {
		return ResponseType(nil), err
	}

	return _send[ResponseType](t, requestID, eventJS)
}

func _send[ResponseType any](t *TDLib, requestID string, str string) (ResponseType, error) {
	ch := make(chan []byte)

	t.responseQueueLocker.Lock()
	t.responseQueue[requestID] = ch
	t.responseQueueLocker.Unlock()

	err := t.fireStringQuery(str)
	if err != nil {
		return ResponseType(nil), err
	}

	resp := <-ch

	t.responseQueueLocker.Lock()
	close(ch)
	delete(t.responseQueue, requestID)
	t.responseQueueLocker.Unlock()

	var errEvent incomingevents.ErrorEvent
	err = json.Unmarshal(resp, &errEvent)
	if err != nil {
		return ResponseType(nil), err
	}

	if errEvent.Type == "error" {
		var str string

		err = json.Unmarshal(errEvent.Message, &str)
		if err != nil {
			return ResponseType(nil), err
		}

		return ResponseType(nil), fmt.Errorf("%d: %s", errEvent.Code, str)
	}

	var respObj ResponseType
	err = json.Unmarshal(resp, &respObj)
	if err != nil {
		return ResponseType(nil), err
	}

	return respObj, nil
}
