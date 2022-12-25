package api

type GuildMember struct {
	Avatar           *string `json:"avatar"`
	IsServerMuted    bool    `json:"mute"`
	IsServerDeafened bool    `json:"deaf"`
	Nickname         string  `json:"nick"`
	User             User    `json:"user"`
}
