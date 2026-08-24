//go:build cgo && tdjson

package tdjson

// #cgo linux CFLAGS: -I/usr/local/include
// #cgo linux LDFLAGS: -Wl,-rpath=/usr/local/lib -ltdjson
// #include <stdlib.h>
// #include "callbacks.h"
import "C"

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

// CreateClientID allocates a new TDLib client identifier.
func CreateClientID() int {
	return int(C.td_create_client_id())
}

// Send enqueues payload for the given client. td_send is fire-and-forget and
// cannot fail synchronously, so there is no error return.
func Send(clientID int, payload []byte) {
	cstr := C.CString(string(payload))
	C.td_send(C.int(clientID), cstr)
	C.free(unsafe.Pointer(cstr))
}

// Execute runs a request synchronously via td_execute. Only a handful of
// requests (log stream configuration, etc.) may be executed this way.
// Returns nil when TDLib has no synchronous answer.
func Execute(payload []byte) []byte {
	cstr := C.CString(string(payload))
	defer C.free(unsafe.Pointer(cstr))
	res := C.td_execute(cstr)
	if res == nil {
		return nil
	}
	return []byte(C.GoString(res))
}

// Receive waits up to timeoutSeconds for the next message from any client
// and returns its raw JSON. It returns nil on timeout; the returned buffer
// is a fresh copy owned by the caller.
func Receive(timeoutSeconds float64) []byte {
	res := C.td_receive(C.double(timeoutSeconds))
	if res == nil {
		return nil
	}
	return []byte(C.GoString(res))
}

var (
	logCallbackSet atomic.Bool

	logHandlerMu sync.RWMutex
	logHandler   func(verbosityLevel int, message string)
)

// SetLogHandler installs the handler invoked from TDLib's log callback. It
// may be changed at any time; a nil handler silences log output.
func SetLogHandler(h func(verbosityLevel int, message string)) {
	logHandlerMu.Lock()
	logHandler = h
	logHandlerMu.Unlock()
}

//export go_td_log_message_callback_ptr
func go_td_log_message_callback_ptr(verbosityLevel C.int, message *C.char) {
	logHandlerMu.RLock()
	h := logHandler
	logHandlerMu.RUnlock()
	if h != nil {
		h(int(verbosityLevel), C.GoString(message))
	}
}

// SetLogVerbosity registers the Go log callback with TDLib at the given
// maximum verbosity level. The callback can only be registered once per
// process; subsequent calls are silently ignored.
func SetLogVerbosity(level int) {
	if logCallbackSet.CompareAndSwap(false, true) {
		C.set_log_message_callback(C.int(level))
	}
}
