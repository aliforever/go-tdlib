package entities

type GroupCall struct {
	ID                           int64           `json:"id"`                               // Group call identifier
	Title                        string          `json:"title"`                            // Group call title; for video chats only
	InviteLink                   string          `json:"invite_link"`                      // Invite link for the group call
	ScheduledStartDate           int64           `json:"scheduled_start_date"`             // Expected start time (Unix timestamp)
	EnabledStartNotification     bool            `json:"enabled_start_notification"`       // True if user will get a notification on start
	IsActive                     bool            `json:"is_active"`                        // True if call is active
	IsVideoChat                  bool            `json:"is_video_chat"`                    // True if the call is bound to a chat
	IsRtmpStream                 bool            `json:"is_rtmp_stream"`                   // True if the call is an RTMP stream
	IsJoined                     bool            `json:"is_joined"`                        // True if the call is joined
	NeedRejoin                   bool            `json:"need_rejoin"`                      // True if user was kicked and needs rejoin
	IsOwned                      bool            `json:"is_owned"`                         // True if user is the owner
	CanBeManaged                 bool            `json:"can_be_managed"`                   // True if the current user can manage the call
	ParticipantCount             int64           `json:"participant_count"`                // Number of participants
	HasHiddenListeners           bool            `json:"has_hidden_listeners"`             // True if muted users aren't returned
	LoadedAllParticipants        bool            `json:"loaded_all_participants"`          // True if all participants are loaded
	RecentSpeakers               []RecentSpeaker `json:"recent_speakers"`                  // Up to 3 recent speakers
	IsMyVideoEnabled             bool            `json:"is_my_video_enabled"`              // True if current user's video is enabled
	IsMyVideoPaused              bool            `json:"is_my_video_paused"`               // True if current user's video is paused
	CanEnableVideo               bool            `json:"can_enable_video"`                 // True if user can broadcast video/screen
	MuteNewParticipants          bool            `json:"mute_new_participants"`            // True if only admins can unmute new participants
	CanToggleMuteNewParticipants bool            `json:"can_toggle_mute_new_participants"` // True if current user can toggle mute setting
	RecordDuration               int64           `json:"record_duration"`                  // Ongoing recording duration (seconds)
	IsVideoRecorded              bool            `json:"is_video_recorded"`                // True if video file is being recorded
	Duration                     int64           `json:"duration"`                         // Call duration (ended calls only)
}
