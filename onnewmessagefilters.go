package tdlib

import (
	"github.com/aliforever/go-tdlib/entities"
	"github.com/aliforever/go-tdlib/incomingevents"
)

type newMessageHandler struct {
	filters *OnNewMessageFilters
	handler func(message *incomingevents.UpdateNewMessage)
}

// checkFilters checks if the message matches the filters
func (h *newMessageHandler) checkFilters(message *incomingevents.UpdateNewMessage) bool {
	if h.filters == nil {
		return true
	}

	return h.filters.FilterMatchesMessage(message)
}

// newOnNewMessageHandler creates a new handler for UpdateNewMessage event
func newOnNewMessageHandler(
	handler func(message *incomingevents.UpdateNewMessage),
	filters *OnNewMessageFilters,
) *newMessageHandler {

	return &newMessageHandler{filters: filters, handler: handler}
}

type OnNewMessageFilters struct {
	ChatId     *int64
	SenderType *entities.MessageSenderType
	IsOutgoing *bool
	Custom     func(message *incomingevents.UpdateNewMessage) bool
}

func NewMessageFilters() *OnNewMessageFilters {
	return &OnNewMessageFilters{}
}

func (f *OnNewMessageFilters) SetChatId(chatId int64) *OnNewMessageFilters {
	f.ChatId = &chatId

	return f
}

func (f *OnNewMessageFilters) SetSenderTypeUser() *OnNewMessageFilters {
	user := entities.MessageSenderTypeUser

	f.SenderType = &user

	return f
}

func (f *OnNewMessageFilters) SetSenderTypeChat() *OnNewMessageFilters {
	chat := entities.MessageSenderTypeChat

	f.SenderType = &chat

	return f
}

func (f *OnNewMessageFilters) SetIsOutgoingTrue() *OnNewMessageFilters {
	isOutgoing := true

	f.IsOutgoing = &isOutgoing

	return f
}

func (f *OnNewMessageFilters) SetIsOutgoingFalse() *OnNewMessageFilters {
	isOutgoing := false

	f.IsOutgoing = &isOutgoing

	return f
}

func (f *OnNewMessageFilters) SetCustomFilter(
	fn func(message *incomingevents.UpdateNewMessage) bool,
) *OnNewMessageFilters {

	f.Custom = fn

	return f
}

// FilterMatchesMessage checks if the message matches the filters
func (f *OnNewMessageFilters) FilterMatchesMessage(message *incomingevents.UpdateNewMessage) bool {
	if f.ChatId != nil && message.Message.ChatId != *f.ChatId {
		return false
	}

	if f.SenderType != nil && message.Message.SenderId.Type != *f.SenderType {
		return false
	}

	if f.IsOutgoing != nil && message.Message.IsOutgoing != *f.IsOutgoing {
		return false
	}

	return true
}
