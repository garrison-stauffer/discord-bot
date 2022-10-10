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
	Id           string       `json:"id"`
	TextToSpeech bool         `json:"tts"`
	Timestamp    string       `json:"timestamp"`
	Author       author       `json:"author"`
	Attachments  []attachment `json:"attachments"`
	Embeds       []embed      `json:"embeds"`
	GuildId      string       `json:"guild_id"`
	ChannelId    string       `json:"channel_id"`
}

type author struct {
	User string `json:"username"`
	Id   string `json:"id"`
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

func (a *App) handleChat(msg chatMessage) error {
	fmt.Println("uhh, received a chat message?")
	if len(msg.Embeds) > 0 {
		for _, embed := range msg.Embeds {
			if embed.Provider == nil {
				continue
			}

			if strings.ToLower(embed.Provider.Name) == "youtube" {
				log.Printf("received youtube link - %s\n", embed.Url)

				isMusic, err := a.ytClient.IsMusicVideo(embed.Url)

				if err != nil {
					return fmt.Errorf("error calling yt api %w", err)
				}

				if isMusic {
					log.Println("AHHHHHHHHH WE FOUND MUSIC????")

					// need to extract guild + member ids
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

					req, err = api.NewMusicVideoReact(msg.ChannelId, msg.Id, "🚨", a.botSecret)
					res, err = http.DefaultClient.Do(req)
					if err != nil {
						return fmt.Errorf("error setting user timeout %v", err)
					} else {
						foo, _ := io.ReadAll(res.Body)
						fmt.Printf("react received %s\n", string(foo))
					}
				} else {
					log.Println("Huh")
				}
			}
		}
	}

	return nil
}
