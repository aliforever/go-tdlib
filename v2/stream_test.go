package tdlib

import (
	"context"
	"testing"
	"time"

	"github.com/aliforever/go-tdlib/v2/td"
)

type textHit struct {
	msg  *td.Message
	text *td.MessageText
}

func TestStreamWithContentNarrowsToVariant(t *testing.T) {
	ft := &fakeTransport{}
	c := NewClient(40, ft)
	defer c.Close()

	got := make(chan textHit, 1)
	sub := c.Messages().WithContent[*td.MessageText]().Handle(
		func(ctx context.Context, m *td.Message, content *td.MessageText) {
			got <- textHit{msg: m, text: content}
		})
	defer sub.Close()

	// A photo update must be skipped (content @type mismatch). FIFO delivery
	// makes the ordering assertions deterministic.
	c.enqueue(newMessageUpdate(100, 1, senderUser(5), `{"@type":"messagePhoto"}`))
	c.enqueue(newMessageUpdate(100, 2, senderUser(5), textContent("hello")))

	hit := recvWithin(t, got, 2*time.Second)
	if hit.msg.Id != 2 {
		t.Fatalf("handled message id = %d, want 2 (the photo must be skipped)", hit.msg.Id)
	}
	if hit.text.Text == nil || hit.text.Text.Text != "hello" {
		t.Fatalf("content not decoded as MessageText: %+v", hit.text)
	}
}

func TestStreamInChatFilter(t *testing.T) {
	ft := &fakeTransport{}
	c := NewClient(41, ft)
	defer c.Close()

	got := make(chan textHit, 1)
	sub := c.Messages().InChat(100).WithContent[*td.MessageText]().Handle(
		func(ctx context.Context, m *td.Message, content *td.MessageText) {
			got <- textHit{msg: m, text: content}
		})
	defer sub.Close()

	c.enqueue(newMessageUpdate(200, 1, senderUser(5), textContent("wrong chat")))
	c.enqueue(newMessageUpdate(100, 2, senderUser(5), textContent("right chat")))

	hit := recvWithin(t, got, 2*time.Second)
	if hit.msg.Id != 2 || hit.msg.ChatId != 100 {
		t.Fatalf("InChat passed the wrong message: id=%d chat=%d", hit.msg.Id, hit.msg.ChatId)
	}
}

func TestStreamFromUserFilter(t *testing.T) {
	ft := &fakeTransport{}
	c := NewClient(42, ft)
	defer c.Close()

	got := make(chan textHit, 1)
	sub := c.Messages().FromUser(5).WithContent[*td.MessageText]().Handle(
		func(ctx context.Context, m *td.Message, content *td.MessageText) {
			got <- textHit{msg: m, text: content}
		})
	defer sub.Close()

	c.enqueue(newMessageUpdate(100, 1, senderUser(6), textContent("other user")))
	c.enqueue(newMessageUpdate(100, 2, senderChat(55), textContent("on behalf of a chat")))
	c.enqueue(newMessageUpdate(100, 3, senderUser(5), textContent("from user 5")))

	hit := recvWithin(t, got, 2*time.Second)
	if hit.msg.Id != 3 {
		t.Fatalf("FromUser passed the wrong message: id=%d", hit.msg.Id)
	}
}

func TestStreamImmutability(t *testing.T) {
	ft := &fakeTransport{}
	c := NewClient(43, ft)
	defer c.Close()

	base := c.Messages()
	narrow := base.InChat(100)
	if len(base.filters) != 0 {
		t.Fatal("InChat mutated the base stream")
	}
	if len(narrow.filters) != 1 {
		t.Fatalf("narrowed stream has %d filters, want 1", len(narrow.filters))
	}
}
