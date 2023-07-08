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
	ChatId      *int64
	SenderType  *entities.MessageSenderType
	IsOutgoing  *bool
	IsText      *bool
	IsAudio     *bool
	IsAnimation *bool
	IsPhoto     *bool
	IsDocument  *bool
	IsVideo     *bool
	IsVideoNote *bool
	IsVoiceNote *bool
	Custom      func(message *incomingevents.UpdateNewMessage) bool
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

func (f *OnNewMessageFilters) SetIsTextTrue() *OnNewMessageFilters {
	isText := true

	f.IsText = &isText

	return f
}

func (f *OnNewMessageFilters) SetIsTextFalse() *OnNewMessageFilters {
	isText := false

	f.IsText = &isText

	return f
}

func (f *OnNewMessageFilters) SetIsAudioTrue() *OnNewMessageFilters {
	isAudio := true

	f.IsAudio = &isAudio

	return f
}

func (f *OnNewMessageFilters) SetIsAudioFalse() *OnNewMessageFilters {
	isAudio := false

	f.IsAudio = &isAudio

	return f
}

func (f *OnNewMessageFilters) SetIsAnimationTrue() *OnNewMessageFilters {
	isAnimation := true

	f.IsAnimation = &isAnimation

	return f
}

func (f *OnNewMessageFilters) SetIsAnimationFalse() *OnNewMessageFilters {
	isAnimation := false

	f.IsAnimation = &isAnimation

	return f
}

func (f *OnNewMessageFilters) SetIsPhotoTrue() *OnNewMessageFilters {
	isPhoto := true

	f.IsPhoto = &isPhoto

	return f
}

func (f *OnNewMessageFilters) SetIsPhotoFalse() *OnNewMessageFilters {
	isPhoto := false

	f.IsPhoto = &isPhoto

	return f
}

func (f *OnNewMessageFilters) SetIsDocumentTrue() *OnNewMessageFilters {
	isDocument := true

	f.IsDocument = &isDocument

	return f
}

func (f *OnNewMessageFilters) SetIsDocumentFalse() *OnNewMessageFilters {
	isDocument := false

	f.IsDocument = &isDocument

	return f
}

func (f *OnNewMessageFilters) SetIsVideoTrue() *OnNewMessageFilters {
	isVideo := true

	f.IsVideo = &isVideo

	return f
}

func (f *OnNewMessageFilters) SetIsVideoFalse() *OnNewMessageFilters {
	isVideo := false

	f.IsVideo = &isVideo

	return f
}

func (f *OnNewMessageFilters) SetIsVideoNoteTrue() *OnNewMessageFilters {
	isVideoNote := true

	f.IsVideoNote = &isVideoNote

	return f
}

func (f *OnNewMessageFilters) SetIsVideoNoteFalse() *OnNewMessageFilters {
	isVideoNote := false

	f.IsVideoNote = &isVideoNote

	return f
}

func (f *OnNewMessageFilters) SetIsVoiceNoteTrue() *OnNewMessageFilters {
	isVoiceNote := true

	f.IsVoiceNote = &isVoiceNote

	return f
}

func (f *OnNewMessageFilters) SetIsVoiceNoteFalse() *OnNewMessageFilters {
	isVoiceNote := false

	f.IsVoiceNote = &isVoiceNote

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

	if f.IsText != nil && message.Message.Content.Type != entities.ContentTypeText {
		return false
	}

	if f.IsAnimation != nil && message.Message.Content.Type != entities.ContentTypeAnimation {
		return false
	}

	if f.IsPhoto != nil && message.Message.Content.Type != entities.ContentTypePhoto {
		return false
	}

	if f.IsDocument != nil && message.Message.Content.Type != entities.ContentTypeDocument {
		return false
	}

	if f.IsVideo != nil && message.Message.Content.Type != entities.ContentTypeVideo {
		return false
	}

	if f.IsVideoNote != nil && message.Message.Content.Type != entities.ContentTypeVideoNote {
		return false
	}

	if f.IsVoiceNote != nil && message.Message.Content.Type != entities.ContentTypeVoiceNote {
		return false
	}

	if f.IsAudio != nil && message.Message.Content.Type != entities.ContentTypeAudio {
		return false
	}

	if f.Custom != nil && !f.Custom(message) {
		return false
	}

	return true
}
