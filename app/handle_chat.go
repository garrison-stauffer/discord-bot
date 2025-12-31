package app

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"garrison-stauffer.com/discord-bot/discord/api"
)

type chatMessage struct {
	Id           string       `json:"Id"`
	TextToSpeech bool         `json:"tts"`
	Timestamp    string       `json:"timestamp"`
	Author       author       `json:"author"`
	Attachments  []attachment `json:"attachments"`
	Embeds       []embed      `json:"embeds"`
	GuildId      string       `json:"guild_id"`
	ChannelId    string       `json:"channel_id"`
}

type author struct {
	UserName string `json:"username"`
	Id       string `json:"Id"`
}

type attachment struct {
}

type embed struct {
	Provider *provider `json:"provider"`
	Url      string    `json:"url"`
}

type provider struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

func (a *App) handleNewMessage(msg chatMessage) error {

	isMusicVideo, err := a.isMusicVideo(msg)
	if err != nil {
		slog.Error("error checking if music video", "error", err)
	}

	if isMusicVideo {
		slog.Info("detected music video in message", "author", msg.Author.UserName, "channel_id", msg.ChannelId)

		req, err := api.NewTimeoutRequest(msg.GuildId, msg.Author.Id, a.botSecret)
		if err != nil {
			return err
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("error setting user timeout %v", err)
		} else {
			foo, _ := io.ReadAll(res.Body)
			slog.Debug("timeout response received", "response", string(foo))
		}

		req, err = api.NewChatReact(msg.ChannelId, msg.Id, "🚨", a.botSecret)
		res, err = http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("error setting user timeout %v", err)
		} else {
			foo, _ := io.ReadAll(res.Body)
			slog.Debug("react response received", "response", string(foo))
		}
	}

	return nil
}

func (a *App) handleMessageUpdate(msg chatMessage) error {

	isMusicVideo, err := a.isMusicVideo(msg)
	if err != nil {
		return err
	}

	if isMusicVideo {
		slog.Info("detected music video in updated message", "author", msg.Author.UserName, "channel_id", msg.ChannelId)

		req, err := api.NewTimeoutRequest(msg.GuildId, msg.Author.Id, a.botSecret)
		if err != nil {
			return err
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("error setting user timeout %v", err)
		} else {
			foo, _ := io.ReadAll(res.Body)
			slog.Debug("timeout response received", "response", string(foo))
		}

		req, err = api.NewChatReact(msg.ChannelId, msg.Id, "🚨", a.botSecret)
		res, err = http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("error setting user timeout %v", err)
		} else {
			foo, _ := io.ReadAll(res.Body)
			slog.Debug("react response received", "response", string(foo))
		}

		req, err = api.NewDeleteChatReact(msg.ChannelId, msg.Id, "❤️", a.botSecret)
		res, err = http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("error deleting default react %v", err)
		} else {
			foo, _ := io.ReadAll(res.Body)
			slog.Debug("default react deleted", "response", string(foo))
		}
	}

	return nil
}

func (a *App) isMusicVideo(msg chatMessage) (bool, error) {
	if len(msg.Embeds) > 0 {
		for _, embed := range msg.Embeds {
			if embed.Provider == nil {
				continue
			}

			if strings.ToLower(embed.Provider.Name) == "youtube" {
				slog.Info("found youtube link", "url", embed.Url)

				isMusic, err := a.ytClient.IsMusicVideo(embed.Url)

				if err != nil {
					return false, fmt.Errorf("error calling yt api %w", err)
				}

				return isMusic, nil
			}
		}
	}

	return false, nil
}
