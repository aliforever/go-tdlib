package tdlib

// #cgo linux CFLAGS: -I/usr/local/include
// #cgo linux LDFLAGS: -Wl,-rpath=/usr/local/lib -ltdjson
// #include <stdlib.h>
// #include "callbacks.h"
import "C"

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aliforever/go-tdlib/config"
	"github.com/aliforever/go-tdlib/incomingevents"
	"github.com/aliforever/go-tdlib/outgoingevents"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"math/rand"
	"sync"
	"time"
)

type TDLib struct {
	clientID int

	updatesChan chan []byte

	cfg config.Config

	apiID   int64
	apiHash string

	handlers *Handlers

	responseQueueLocker sync.Mutex
	responseQueue       map[string]chan []byte

	logger *logrus.Logger
}

func newClientV2(
	clientID int,
	updateChannel chan []byte,
	apiID int64,
	apiHash string,
	handlers *Handlers,
	cfg *config.Config,
	logger *logrus.Logger,
) *TDLib {
	rand.Seed(time.Now().Unix())

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

	return &TDLib{
		clientID:      clientID,
		updatesChan:   updateChannel,
		apiID:         apiID,
		apiHash:       apiHash,
		handlers:      handlers,
		cfg:           *cfg,
		responseQueue: map[string]chan []byte{},
		logger:        logger,
	}
}

func (t *TDLib) ReceiveUpdates() error {
	return t.receiveUpdates()
}

func (t *TDLib) fireStringQuery(data string) error {
	C.td_send(C.int(t.clientID), C.CString(data))

	return nil
}

func (t *TDLib) receiveUpdates() error {
	// This is to make sure that the client is ready to receive updates
	go t.GetAuthorizationState()

	for updateBytes := range t.updatesChan {
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

		if ie.Type == "error" && ie.RequestID == "" {
			if t.handlers != nil && t.handlers.onError != nil {
				errEvent, err := incomingevents.ErrorFromBytes(updateBytes)
				if err != nil {
					return err
				}
				go t.handlers.onError(errEvent)
			}
			continue
		}

		if ie.RequestID != "" {
			go t.informResponse(ie.RequestID, updateBytes)
			continue
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

		if t.handlers.authorizationHandler != nil && isAuthorizationEvent(event) {
			go t.handlers.authorizationHandler.Process(t, event.AuthorizationState.Type)
		}

		if t.handlers.onUpdateAuthorizationState != nil && event.Type == "updateAuthorizationState" && event.AuthorizationState != nil {
			go t.handlers.onUpdateAuthorizationState(event.AuthorizationState.Type)
		}

		go t.handleEventType(event.Type, updateBytes)
	}

	return nil
}

func (t *TDLib) handleEventType(eventType string, data []byte) {
	if handler := t.getEventTypeHandler(eventType); handler != nil {
		err := handler.Handle(data)
		if err != nil {
			// TODO: Add Logs
			return
		}
	}

	if eventType == "updateNewMessage" {
		for i := range t.handlers.onNewMessageHandlers {
			handler := t.handlers.onNewMessageHandlers[i]

			if err := NewEventHandlerWithFilter(handler.handler, handler.checkFilters).Handle(data); err != nil {
				t.logger.Error(err)
			}
		}
	}

	if eventType == "updateMessageEdited" {
		for i := range t.handlers.onMessageEditedHandlers {
			handler := t.handlers.onMessageEditedHandlers[i]

			if err := NewEventHandler[incomingevents.UpdateMessageEdited](handler).Handle(data); err != nil {
				t.logger.Error(err)
			}
		}
	}

	if eventType == "updateMessageContent" {
		for i := range t.handlers.onMessageContentHandlers {
			handler := t.handlers.onMessageContentHandlers[i]

			if err := NewEventHandler[incomingevents.UpdateMessageContent](handler).Handle(data); err != nil {
				t.logger.Error(err)
			}
		}
	}
}

func (t *TDLib) getEventTypeHandler(eventType string) event {
	t.handlers.eventTypeHandlerLocker.Lock()
	defer t.handlers.eventTypeHandlerLocker.Unlock()

	if handler, ok := t.handlers.eventTypeHandlers[eventType]; ok && handler != nil {
		return handler
	}

	return nil
}

func (t *TDLib) getResponseQueueByRequestID(requestID string) chan []byte {
	t.responseQueueLocker.Lock()
	defer t.responseQueueLocker.Unlock()

	if ch, ok := t.responseQueue[requestID]; ok {
		return ch
	}

	return nil
}

func (t *TDLib) newResponseChannel(requestID string) chan []byte {
	ch := make(chan []byte)

	t.responseQueueLocker.Lock()
	defer t.responseQueueLocker.Unlock()

	t.responseQueue[requestID] = ch

	return ch
}

func (t *TDLib) removeResponseChannel(requestID string, ch chan []byte) {
	t.responseQueueLocker.Lock()
	defer t.responseQueueLocker.Unlock()

	close(ch)
	delete(t.responseQueue, requestID)
}

func (t *TDLib) informResponse(requestID string, data []byte) {
	ch := t.getResponseQueueByRequestID(requestID)
	if ch != nil {
		ch <- data
	}
}

func sendMap[ResponseType any](
	t *TDLib,
	requestType string,
	data map[string]interface{},
	sendOptions ...*SendOptions,
) (*ResponseType, error) {

	requestID := uuid.NewString()

	eventJS, err := outgoingevents.NewEventJSONFromMap(requestID, requestType, data)
	if err != nil {
		return nil, err
	}

	return _send[ResponseType](t, requestID, eventJS, sendOptions...)
}

func send[ResponseType any](
	t *TDLib,
	data outgoingevents.EventInterface,
	sendOptions ...*SendOptions,
) (*ResponseType, error) {

	requestID := uuid.NewString()

	eventJS, err := outgoingevents.NewEventJSON(requestID, data)
	if err != nil {
		return nil, err
	}

	return _send[ResponseType](t, requestID, eventJS, sendOptions...)
}

func _send[ResponseType any](
	t *TDLib,
	requestID string,
	str string,
	sendOptions ...*SendOptions,
) (*ResponseType, error) {
	ch := t.newResponseChannel(requestID)

	if t.logger != nil {
		t.logger.Debug(str)
	}

	err := t.fireStringQuery(str)
	if err != nil {
		return nil, err
	}

	var resp []byte

	if len(sendOptions) == 0 || sendOptions[0].timeout == nil {
		resp = <-ch
	} else {
		ticker := time.NewTicker(*sendOptions[0].timeout)

		select {
		case <-ticker.C:
			return nil, errors.New("request_timed_out")
		case resp = <-ch:
			break
		}
	}

	t.removeResponseChannel(requestID, ch)

	var errEvent incomingevents.ErrorEvent
	err = json.Unmarshal(resp, &errEvent)
	if err != nil {
		return nil, err
	}

	if errEvent.Type == "error" {
		var str string

		err = json.Unmarshal(errEvent.Message, &str)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("%d: %s", errEvent.Code, str)
	}

	var respObj *ResponseType
	err = json.Unmarshal(resp, &respObj)
	if err != nil {
		return nil, err
	}

	return respObj, nil
}
