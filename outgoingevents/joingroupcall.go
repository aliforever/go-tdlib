package outgoingevents

import (
	"encoding/json"
	"fmt"

	"github.com/aliforever/go-tdlib/entities"
)

type JoinGroupCall struct {
	InputGroupCall entities.InputGroupCall          `json:"input_group_call"`
	JoinParameters entities.GroupCallJoinParameters `json:"join_parameters"`
}

func (JoinGroupCall) Type() string {
	return "joinGroupCall"
}

func (s JoinGroupCall) MarshalJSON() ([]byte, error) {
	type basicJoinGroupCall struct {
		Type           string                           `json:"@type"`
		InputGroupCall entities.InputGroupCall          `json:"input_group_call"`
		JoinParameters entities.GroupCallJoinParameters `json:"join_parameters"`
	}

	b := basicJoinGroupCall{
		Type: s.Type(),
	}

	switch t := s.InputGroupCall.(type) {
	case *entities.InputGroupCallLink:
		type JoinGroupCallLink struct {
			Type string `json:"@type"`
			*entities.InputGroupCallLink
		}

		return json.Marshal(&struct {
			*basicJoinGroupCall
			InputGroupCall *JoinGroupCallLink `json:"input_group_call"`
		}{
			basicJoinGroupCall: &b,
			InputGroupCall: &JoinGroupCallLink{
				Type:               t.Type(),
				InputGroupCallLink: t,
			},
		})
	case *entities.InputGroupCallMessage:
		type JoinGroupCallMessage struct {
			Type string `json:"@type"`
			*entities.InputGroupCallMessage
		}

		return json.Marshal(&struct {
			*basicJoinGroupCall
			InputGroupCall *JoinGroupCallMessage `json:"input_group_call"`
		}{
			basicJoinGroupCall: &b,
			InputGroupCall: &JoinGroupCallMessage{
				Type:                  t.Type(),
				InputGroupCallMessage: t,
			},
		})
	default:
		return nil, fmt.Errorf("invalid type")
	}
}
