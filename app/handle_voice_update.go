package app

import (
	"fmt"
	"log/slog"
	"net/http"

	"garrison-stauffer.com/discord-bot/discord/api"
)

func (a *App) handleVoiceUpdate(msg api.VoiceState) error {
	guild, ok := a.cache.Guilds[msg.GuildId]
	if !ok {
		return fmt.Errorf("guild was not cached, could not handle voice update: guild_id=%s", msg.GuildId)
	}

	user, ok := guild.Members[msg.UserId]
	if !ok {
		return fmt.Errorf("member with id %s was not loaded, could not handle voice update", msg.UserId)
	}

	err := a.resolveChangeAndSendMessage(msg, guild, user)

	user.ConnectedTo = msg.ChannelId

	user.IsServerMuted = msg.IsServerMuted
	user.IsSelfDeafened = msg.IsSelfDeafened
	user.IsStreaming = msg.IsStreaming
	user.IsSelfDeafened = msg.IsSelfDeafened
	user.IsSelfMuted = msg.IsSelfMuted
	user.IsVideoCalling = msg.IsVideoCalling

	return err
}

func (a *App) resolveChangeAndSendMessage(msg api.VoiceState, guild *Guild, user *GuildMember) error {
	newChannel := msg.ChannelId
	oldChannel := user.ConnectedTo

	username := user.Name
	if user.Nickname != "" {
		username = user.Nickname
	}

	var message string

	if newChannel != "" {
		newChannelInfo, ok := guild.Channels[newChannel]
		if !ok {
			return fmt.Errorf("could not identify channel id %s when handling voice update", newChannel)
		}

		if oldChannel == "" {
			// user joined a channel
			message = fmt.Sprintf("%s connected to %s", username, newChannelInfo.Name)
		} else if newChannel == oldChannel {
			// user updated their state
			if msg.IsSelfDeafened != user.IsSelfDeafened {
				action := resolveAction(msg.IsSelfDeafened, "deafened", "un-deafened")
				message = fmt.Sprintf("%s %s themselves in %s", username, action, newChannelInfo.Name)
			} else if msg.IsSelfMuted != user.IsSelfMuted {
				action := resolveAction(msg.IsSelfMuted, "muted", "un-muted")
				message = fmt.Sprintf("%s %s themselves in %s", username, action, newChannelInfo.Name)
			} else if msg.IsStreaming != user.IsStreaming {
				action := resolveAction(msg.IsStreaming, "started streaming", "stopped streaming")
				message = fmt.Sprintf("%s %s in %s", username, action, newChannelInfo.Name)
			} else if msg.IsVideoCalling != user.IsVideoCalling {
				action := resolveAction(msg.IsVideoCalling, "started video chatting", "stopped video chatting")
				funny := resolveAction(msg.IsVideoCalling, "someone make them stop - please", "thank god")
				message = fmt.Sprintf("%s %s in %s, %s", username, action, newChannelInfo.Name, funny)
			} else if msg.IsServerMuted != user.IsServerMuted {
				message = fmt.Sprintf("%s was %s by an admin", username, resolveAction(msg.IsServerMuted, "muted", "un-muted"))
			} else if msg.IsServerDeafened != user.IsServerDeafened {
				message = fmt.Sprintf("%s was %s by an admin", username, resolveAction(msg.IsServerDeafened, "deafened", "un-deafened"))
			} else {
				message = fmt.Sprintf("%s did something in %s, but we're not really sure what yet", username, newChannelInfo.Name)
			}
		} else {
			oldChannelInfo, ok := guild.Channels[oldChannel]
			if !ok {
				return fmt.Errorf("could not identify the old channel id %s when handling voice update", oldChannel)
			}

			message = fmt.Sprintf("%s moved to %s from %s", username, newChannelInfo.Name, oldChannelInfo.Name)
		}
	} else if oldChannel != "" {
		oldChannelInfo, ok := guild.Channels[oldChannel]
		if !ok {
			return fmt.Errorf("could not identify the old channel id %s when handling voice update", oldChannel)
		}

		message = fmt.Sprintf("%s disconnected from %s", username, oldChannelInfo.Name)
	}

	sendChannel := voiceLogsChannelIdForGuild(msg.GuildId)
	if sendChannel == "" {
		return fmt.Errorf("could not resolve channel to send to for this voice update")
	}

	req, _ := api.NewMessage(sendChannel, a.botSecret, message)
	_, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("error sending voice update message", "error", err)
		return err
	}

	slog.Info("sent voice update message", "guild_id", msg.GuildId, "user", username, "message", message)
	return nil
}

const mumble = "1023753903994572820"
const testServer = "663183295818760213"

func voiceLogsChannelIdForGuild(guildId string) string {
	switch guildId {
	case mumble:
		return "1056472933314334800"
	case testServer:
		return "1056460813390598195"
	default:
		return ""
	}
}

func resolveAction(predicate bool, ifTrue, ifFalse string) string {
	if predicate {
		return ifTrue
	} else {
		return ifFalse
	}
}
