package tdlib

import "C"
import (
	"encoding/json"
	"fmt"
	"github.com/aliforever/go-tdlib/config"
	"github.com/sirupsen/logrus"
	"sync"
	"unsafe"
)

// #cgo linux CFLAGS: -I/usr/local/include
// #cgo linux LDFLAGS: -Wl,-rpath=/usr/local/lib -ltdjson
// #include <stdlib.h>
// #include "callbacks.h"
import "C"

//export go_td_log_message_callback_ptr
func go_td_log_message_callback_ptr(verbosityLevel C.int, message *C.char) {
	goMessage := C.GoString(message)
	fmt.Printf("Received log message with verbosity level %d: %s\n", int(verbosityLevel), goMessage)
}

type Manager struct {
	handlers *ManagerHandlers

	clientUpdateChannels sync.Map

	options *ManagerOptions
}

// NewManager creates a new Manager instance
func NewManager(handlers *ManagerHandlers, options *ManagerOptions) *Manager {
	if options != nil {
		C.set_log_message_callback(C.int(options.LogVerbosityLevel))

		if options.LogPath != "" {
			cfgBytes, _ := json.Marshal(map[string]interface{}{
				"@type": "setLogStream",
				"log_stream": map[string]interface{}{
					"@type":         "logStreamFile",
					"path":          options.LogPath,
					"max_file_size": 10485760,
				},
			})

			query := C.CString(string(cfgBytes))
			C.td_execute(query)
			C.free(unsafe.Pointer(query))
		}
	}

	manager := &Manager{
		handlers: handlers,
		options:  options,
	}

	go manager.receiveUpdates()

	return manager
}

// NewClient creates a new client instance
func (m *Manager) NewClient(
	apiID int64,
	apiHash string,
	handlers *Handlers,
	cfg *config.Config,
	logger *logrus.Logger,
) *TDLib {

	clientID := m.newClientID()

	updateChannel := make(chan []byte)

	m.clientUpdateChannels.Store(clientID, updateChannel)

	return newClientV2(
		clientID,
		updateChannel,
		apiID,
		apiHash,
		handlers,
		cfg,
		logger,
	)
}

func (m *Manager) newClientID() int {
	result := C.td_create_client_id()

	return int(result)
}

func (m *Manager) receiveNextUpdate(timeout float64) []byte {
	result := C.td_receive(C.double(timeout))
	if result == nil {
		return nil
	}

	return []byte(C.GoString(result))
}

func (m *Manager) receiveUpdates() {
	for {
		updateBytes := m.receiveNextUpdate(10)

		if updateBytes == nil || len(updateBytes) == 0 {
			continue
		}

		if m.handlers != nil && m.handlers.onRawIncomingEvent != nil {
			go m.handlers.onRawIncomingEvent(updateBytes)
		}

		clientID := m.getClientID(updateBytes)
		if clientID != nil {
			if clientChan, ok := m.getClientChannel(*clientID); ok {
				go m.writeClientEvent(clientChan, updateBytes)

				continue
			} else {
				// TODO: Log Received Update For Unknown Client
				fmt.Printf("Received update for unknown client: %d\n", *clientID)
			}
		}

		// TODO: Log General Received Update, Doesn't Belong To Any Client
		fmt.Printf("Received update: %s\n", string(updateBytes))
	}
}

func (m *Manager) getClientChannel(clientID int) (chan []byte, bool) {
	clientValue, ok := m.clientUpdateChannels.Load(clientID)
	if ok {
		clientUpdateChannel, ok := clientValue.(chan []byte)
		if ok {
			return clientUpdateChannel, true
		}
	}

	return nil, false
}

func (m *Manager) writeClientEvent(updateChan chan<- []byte, update []byte) {
	updateChan <- update
}

func (m *Manager) getClientID(update []byte) *int {
	type ClientID struct {
		ID int `json:"@client_id"`
	}

	var clientID ClientID

	err := json.Unmarshal(update, &clientID)
	if err != nil {
		return nil
	}

	return &clientID.ID
}
