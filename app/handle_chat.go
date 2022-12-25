package app

import (
	"fmt"
	"garrison-stauffer.com/discord-bot/discord/api"
	"io"
	"log"
	"net/http"
	"strings"
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
		return err
	}

	if isMusicVideo {
		log.Println("AHHHHHHHHH WE FOUND MUSIC????")

		// need to extract Guild + member ids
		req, err := api.NewTimeoutRequest(msg.GuildId, msg.Author.Id, a.botSecret)
		if err != nil {
			return err
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("error setting user timeout %v", err)
		} else {
			foo, _ := io.ReadAll(res.Body)
			fmt.Printf("timeout received %s\n", string(foo))
		}

		req, err = api.NewChatReact(msg.ChannelId, msg.Id, "🚨", a.botSecret)
		res, err = http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("error setting user timeout %v", err)
		} else {
			foo, _ := io.ReadAll(res.Body)
			fmt.Printf("react received %s\n", string(foo))
		}
	} else {
		// create default react

		log.Println(msg.ChannelId + " is equal to " + resolveVoiceChannelId(msg.GuildId) + " ???")
		if msg.ChannelId == resolveVoiceChannelId(msg.GuildId) {
			// don't react to chat logs
			return nil
		}

		req, err := api.NewChatReact(msg.ChannelId, msg.Id, "❤️", a.botSecret)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("error setting default react %v", err)
		} else {
			foo, _ := io.ReadAll(res.Body)
			fmt.Printf("default react received %s\n", string(foo))
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
		log.Println("AHHHHHHHHH WE FOUND MUSIC????")

		// need to extract Guild + member ids
		req, err := api.NewTimeoutRequest(msg.GuildId, msg.Author.Id, a.botSecret)
		if err != nil {
			return err
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("error setting user timeout %v", err)
		} else {
			foo, _ := io.ReadAll(res.Body)
			fmt.Printf("timeout received %s\n", string(foo))
		}

		req, err = api.NewChatReact(msg.ChannelId, msg.Id, "🚨", a.botSecret)
		res, err = http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("error setting user timeout %v", err)
		} else {
			foo, _ := io.ReadAll(res.Body)
			fmt.Printf("react received %s\n", string(foo))
		}

		req, err = api.NewDeleteChatReact(msg.ChannelId, msg.Id, "❤️", a.botSecret)
		res, err = http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("error deleting default react %v", err)
		} else {
			foo, _ := io.ReadAll(res.Body)
			fmt.Printf("default react received %s\n", string(foo))
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
				log.Printf("received youtube link - %s\n", embed.Url)

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
