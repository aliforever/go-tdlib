package tdlib

import "errors"

var (
	// ErrClientClosed is returned by in-flight and new calls after
	// Client.Close.
	ErrClientClosed = errors.New("tdlib: client is closed")

	// ErrUnknownClient indicates a message arrived for a client ID the
	// manager does not know. It is surfaced through Manager.OnUnknownClient,
	// never returned from a call.
	ErrUnknownClient = errors.New("tdlib: message for unknown client")

	// ErrManagerRunning is returned by Manager.Run when a receive loop is
	// already active.
	ErrManagerRunning = errors.New("tdlib: manager receive loop already running")
)
