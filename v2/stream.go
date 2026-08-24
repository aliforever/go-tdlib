package tdlib

import (
	"context"
	"slices"

	"github.com/aliforever/go-tdlib/v2/td"
)

// MessageStream is a narrowing builder for incoming-message subscriptions
// (updateNewMessage). It is immutable: each method returns a new stream
// carrying the accumulated filters, so a base stream is safe to share.
// Consume it via WithContent[T]().Handle(...).
type MessageStream struct {
	c       *Client
	filters []Filter[*td.Message]
}

// Messages starts a stream over all incoming messages.
func (c *Client) Messages() *MessageStream {
	return &MessageStream{c: c}
}

// InChat keeps only messages posted to the given chat.
func (s *MessageStream) InChat(chatID int64) *MessageStream {
	return s.with(func(m *td.Message) bool { return m.ChatId == chatID })
}

// FromUser keeps only messages sent by the given user. Messages sent on
// behalf of a chat (messageSenderChat, e.g. anonymous admins and channel
// posts) never match.
func (s *MessageStream) FromUser(userID int64) *MessageStream {
	return s.with(func(m *td.Message) bool {
		su, ok := m.SenderId.As[*td.MessageSenderUser]()
		return ok && su.UserId == userID
	})
}

func (s *MessageStream) with(f Filter[*td.Message]) *MessageStream {
	return &MessageStream{c: s.c, filters: append(slices.Clone(s.filters), f)}
}

func (s *MessageStream) matches(m *td.Message) bool {
	return passesFilters(m, s.filters)
}

// WithContent narrows the stream to one message content variant: T's @type
// (e.g. "messageText" for *td.MessageText) becomes the discriminator.
func (s *MessageStream) WithContent[T td.MessageContentVariant]() *TypedMessageStream[T] {
	return &TypedMessageStream[T]{stream: s}
}

// TypedMessageStream is a MessageStream narrowed to one content variant.
type TypedMessageStream[T td.MessageContentVariant] struct {
	stream *MessageStream
}

// Handle subscribes fn to messages of this stream whose content is the T
// variant. The content @type is checked before variant extraction and the
// accumulated stream filters, so non-matching messages cost one string
// comparison and no type assertion. The message pointer and the decoded
// content variant are both handed to fn.
func (t *TypedMessageStream[T]) Handle(fn func(context.Context, *td.Message, T)) Subscription {
	var z T
	kind := z.TDType() // constant per variant; nil-receiver safe
	s := t.stream
	return s.c.On[*td.UpdateNewMessage](func(ctx context.Context, u *td.UpdateNewMessage) {
		m := u.Message
		if m == nil || m.Content.Kind() != kind {
			return
		}
		if !s.matches(m) {
			return
		}
		v, ok := m.Content.As[T]()
		if !ok {
			return // Kind matched but the value did not; cannot happen for known variants
		}
		fn(ctx, m, v)
	})
}
