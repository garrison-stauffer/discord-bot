package app

import (
	"encoding/json"
	"fmt"
	"log"

	"garrison-stauffer.com/discord-bot/discord"
	"garrison-stauffer.com/discord-bot/discord/api"
	"garrison-stauffer.com/discord-bot/discord/gateway"
	"garrison-stauffer.com/discord-bot/youtube"
)

type App struct {
	discordClient discord.Client
	ytClient      youtube.Client
	botSecret     string
	cache         Cache
}

func New(discordClient discord.Client, ytClient youtube.Client, botSecret string) *App {
	return &App{
		discordClient: discordClient,
		ytClient:      ytClient,
		botSecret:     botSecret,
		cache:         newCache(),
	}
}

func (a *App) Handle(msg gateway.Message) error {
	log.Println("received dispatch message")
	switch msg.OpCode {
	case gateway.OpDispatch:
		switch *msg.Type {
		case "READY":
			log.Println("received ready event")
			return nil // no business logic for READY event
		case "MESSAGE_CREATE":
			log.Printf("received uhh %s event\n", *msg.Type)
			dataBytes, _ := json.Marshal(msg.Event)
			var message chatMessage
			err := json.Unmarshal(dataBytes, &message)
			if err != nil {
				return err
			}

			return a.handleNewMessage(message)
		case "MESSAGE_UPDATE":
			log.Printf("received uhh %s event\n", *msg.Type)
			dataBytes, _ := json.Marshal(msg.Event)
			var message chatMessage
			err := json.Unmarshal(dataBytes, &message)
			if err != nil {
				return err
			}

			return a.handleNewMessage(message)
		case "VOICE_STATE_UPDATE":
			log.Println("received voce_state_update event")

			dataBytes, _ := json.Marshal(msg.Event)
			var message api.VoiceState
			err := json.Unmarshal(dataBytes, &message)
			if err != nil {
				return fmt.Errorf("error parsing voice state update %v", err)
			}

			return a.handleVoiceUpdate(message)
		case "PRESENCE_UPDATE":
			log.Println("received present_update event")
			return nil // TODO?
		case "GUILD_CREATE":
			log.Println("bootstrapping server from GUILD_CREATE event")

			dataBytes, _ := json.Marshal(msg.Event)
			var message api.GatewayGuildCreate
			err := json.Unmarshal(dataBytes, &message)
			if err != nil {
				return err
			}

			return a.handleGuildCreate(message)
		default:
			bytes, _ := json.Marshal(msg)
			log.Printf("unhandled dispatch type %s: %s", *msg.Type, string(bytes))
		}
	default:
		bytes, _ := json.Marshal(msg)
		log.Printf("unhandled opcode %s with body %s", msg.OpCode, string(bytes))
	}

	return nil
}
