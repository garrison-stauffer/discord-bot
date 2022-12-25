package api

type GatewayGuildCreate struct {
	GuildId       string        `json:"id"`
	Channels      []Channel     `json:"channels"`
	Members       []GuildMember `json:"members"`
	VoiceStatuses []VoiceState  `json:"voice_states"`
}

type PresenceUpdate struct {
	User    UserRef `json:"user"`
	GuildId string  `json:"guild_id"`
}

type Activity struct {
	Name string `json:"name"`
	Type int    `json:"type"`
}

type VoiceState struct {
	UserId    string `json:"user_id"`
	GuildId   string `json:"guild_id"`
	ChannelId string `json:"channel_id"`

	GuildMember *GuildMember `json:"member"`

	IsServerMuted    bool `json:"mute"`
	IsServerDeafened bool `json:"deaf"`

	IsSelfMuted    bool `json:"self_mute"`
	IsSelfDeafened bool `json:"self_deaf"`
	IsStreaming    bool `json:"self_stream"`
	IsVideoCalling bool `json:"self_video"`
}
