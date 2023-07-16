package outgoingevents

import (
	"encoding/json"
	"fmt"
	"github.com/aliforever/go-tdlib/entities"
)

type SendMessageOptions struct {
	DisableNotification bool        `json:"disable_notification"`
	FromBackground      bool        `json:"from_background"`
	SchedulingState     interface{} `json:"scheduling_state"`
}

type SendMessage struct {
	ChatID              int64                        `json:"chat_id"`
	MessageThreadID     int64                        `json:"message_thread_id"`
	ReplyToMessageID    int64                        `json:"reply_to_message_id"`
	Options             *SendMessageOptions          `json:"options,omitempty"`
	ReplyMarkup         entities.ReplyMarkup         `json:"reply_markup,omitempty"`
	InputMessageContent entities.InputMessageContent `json:"input_message_content"`
}

func (s SendMessage) Type() string {
	return "sendMessage"
}

func (s SendMessage) MarshalJSON() ([]byte, error) {
	type basicSendMessage struct {
		ChatID           int64                `json:"chat_id"`
		MessageThreadID  int64                `json:"message_thread_id"`
		ReplyToMessageID int64                `json:"reply_to_message_id"`
		Options          *SendMessageOptions  `json:"options,omitempty"`
		ReplyMarkup      entities.ReplyMarkup `json:"reply_markup,omitempty"`
	}

	b := basicSendMessage{
		ChatID:           s.ChatID,
		MessageThreadID:  s.MessageThreadID,
		ReplyToMessageID: s.ReplyToMessageID,
		Options:          s.Options,
	}

	switch t := s.InputMessageContent.(type) {

	case *entities.InputMessageText:
		type InputMessageText struct {
			Type string `json:"@type"`
			*entities.InputMessageText
		}

		return json.Marshal(&struct {
			*basicSendMessage
			InputMessageContent *InputMessageText `json:"input_message_content"`
		}{
			basicSendMessage: &b,
			InputMessageContent: &InputMessageText{
				Type:             t.Type(),
				InputMessageText: t,
			},
		})
	case *entities.InputMessageVideo:
		type InputMessageVideo struct {
			Type string `json:"@type"`
			*entities.InputMessageVideo
		}

		return json.Marshal(&struct {
			*basicSendMessage
			InputMessageContent *InputMessageVideo `json:"input_message_content"`
		}{
			basicSendMessage: &b,
			InputMessageContent: &InputMessageVideo{
				Type:              t.Type(),
				InputMessageVideo: t,
			},
		})
	default:
		return nil, fmt.Errorf("invalid type")
	}
}
