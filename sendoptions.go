package tdlib

import "time"

type SendOptions struct {
	timeout *time.Duration
}

func NewSendOptions(timeout time.Duration) *SendOptions {
	return &SendOptions{timeout: &timeout}
}
