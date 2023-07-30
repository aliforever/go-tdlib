package entities

type GroupCall struct {
	ID                           int64           `json:"id"`
	Title                        string          `json:"title"`
	ScheduledStartDate           int64           `json:"scheduled_start_date"`
	EnabledStartNotification     bool            `json:"enabled_start_notification"`
	IsActive                     bool            `json:"is_active"`
	IsJoined                     bool            `json:"is_joined"`
	NeedRejoin                   bool            `json:"need_rejoin"`
	CanBeManaged                 bool            `json:"can_be_managed"`
	ParticipantCount             int64           `json:"participant_count"`
	LoadedAllParticipants        bool            `json:"loaded_all_participants"`
	RecentSpeakers               []RecentSpeaker `json:"recent_speakers"`
	IsMyVideoEnabled             bool            `json:"is_my_video_enabled"`
	CanEnableVideo               bool            `json:"can_enable_video"`
	MuteNewParticipants          bool            `json:"mute_new_participants"`
	CanToggleMuteNewParticipants bool            `json:"can_toggle_mute_new_participants"`
	RecordDuration               int64           `json:"record_duration"`
	IsVideoRecorded              bool            `json:"is_video_recorded"`
	Duration                     int64           `json:"duration"`
}
