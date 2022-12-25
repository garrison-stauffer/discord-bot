package app

import (
	"encoding/json"
	"garrison-stauffer.com/discord-bot/discord/api"
	"garrison-stauffer.com/discord-bot/discord/channels"
	"log"
)

func (a *App) handleGuildCreate(msg api.GatewayGuildCreate) error {
	guild, ok := a.cache.Guilds[msg.GuildId]
	if !ok {
		guild = newGuild(msg.GuildId)
		a.cache.Guilds[msg.GuildId] = guild
	}

	for _, c := range msg.Channels {
		_, ok := guild.Channels[c.Id]
		if ok {
			// already exists, assume nothin gin the channel changed
			continue
		}

		guild.Channels[c.Id] = &GuildChannel{
			Id:   c.Id,
			Type: channels.TypeFrom(c.Type),
			Name: c.Name,
		}
	}

	for _, u := range msg.Members {
		cachedUser, ok := guild.Members[u.User.Id]
		if ok {
			// member already cached, let's just get the updated state if so.
			// this might never be run
			cachedUser.IsServerMuted = u.IsServerMuted
			cachedUser.IsServerDeafened = u.IsServerDeafened
			continue
		}

		guild.Members[u.User.Id] = &GuildMember{
			Id:       u.User.Id,
			Nickname: u.Nickname,
			Name:     u.User.Name,
			IsBot:    u.User.IsBot,

			IsServerMuted:    u.IsServerMuted,
			IsServerDeafened: u.IsServerDeafened,
		}
	}

	for _, v := range msg.VoiceStatuses {
		user, ok := guild.Members[v.UserId]

		if !ok {
			panic("Uh oh, I've never seen this user before? " + v.UserId)
		}

		user.IsSelfMuted = v.IsSelfMuted
		user.IsSelfDeafened = v.IsSelfDeafened
		user.IsStreaming = v.IsStreaming
		user.IsVideoCalling = v.IsVideoCalling

		user.ConnectedTo = v.ChannelId
	}

	json, _ := json.Marshal(a.cache)
	log.Println("fuuuuuck you garrison" + string(json))

	return nil
}
