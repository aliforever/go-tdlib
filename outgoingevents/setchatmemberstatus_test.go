package outgoingevents

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aliforever/go-tdlib/entities"
)

func TestSetChatMemberStatus_MarshalJSON(t *testing.T) {
	type fields struct {
		ChatID   int64
		MemberID int64
		Status   entities.ChatMemberStatus
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name: "Success 1",
			fields: fields{
				ChatID:   123,
				MemberID: 123,
				Status: &entities.ChatMemberStatusAdministrator{
					CustomTitle: "",
					CanBeEdited: false,
					Rights: &entities.ChatAdministratorRights{
						CanManageChat:       true,
						CanChangeInfo:       true,
						CanPostMessages:     true,
						CanEditMessages:     true,
						CanDeleteMessages:   true,
						CanInviteUsers:      true,
						CanRestrictMembers:  true,
						CanPinMessages:      true,
						CanManageTopics:     true,
						CanPromoteMembers:   true,
						CanManageVideoChats: true,
						CanPostStories:      true,
						CanEditStories:      true,
						CanDeleteStories:    true,
						IsAnonymous:         true,
					},
				},
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := SetChatMemberStatus{
				ChatID:   tt.fields.ChatID,
				MemberID: tt.fields.MemberID,
				Status:   tt.fields.Status,
			}
			got, err := s.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			fmt.Println(string(got))

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarshalJSON() got = %v, want %v", got, tt.want)
			}
		})
	}
}
