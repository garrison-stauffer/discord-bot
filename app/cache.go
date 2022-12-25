package app

import "garrison-stauffer.com/discord-bot/discord/channels"

type Cache struct {
	Guilds map[string]*Guild
}

func newCache() Cache {
	return Cache{
		Guilds: map[string]*Guild{},
	}
}

type Guild struct {
	Id       string
	Members  map[string]*GuildMember
	Channels map[string]*GuildChannel
}

func newGuild(id string) *Guild {
	return &Guild{
		Id:       id,
		Members:  map[string]*GuildMember{},
		Channels: map[string]*GuildChannel{},
	}
}

type GuildMember struct {
	Id       string
	Nickname string
	Name     string
	IsBot    bool

	IsServerMuted    bool
	IsServerDeafened bool

	IsSelfMuted    bool
	IsSelfDeafened bool
	IsStreaming    bool
	IsVideoCalling bool

	ConnectedTo string // channel id
	_           string
}

type GuildChannel struct {
	Id   string
	Type channels.Type
	Name string
}
