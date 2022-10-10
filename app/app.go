package app

import (
	"encoding/json"
	"garrison-stauffer.com/discord-bot/discord"
	"garrison-stauffer.com/discord-bot/discord/gateway"
	"garrison-stauffer.com/discord-bot/youtube"
	"log"
)

type App struct {
	discordClient discord.Client
	ytClient      youtube.Client
	botSecret     string
}

func New(discordClient discord.Client, ytClient youtube.Client, botSecret string) *App {
	return &App{
		discordClient: discordClient,
		ytClient:      ytClient,
		botSecret:     botSecret,
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
			log.Println("Fuck me")
			log.Printf("received uhh %s event\n", *msg.Type)
			dataBytes, _ := json.Marshal(msg.Event)
			var message chatMessage
			err := json.Unmarshal(dataBytes, &message)
			if err != nil {
				return err
			}

			return a.handleChat(message)
		case "MESSAGE_UPDATE":
			log.Printf("received uhh %s event\n", *msg.Type)
			dataBytes, _ := json.Marshal(msg.Event)
			var message chatMessage
			err := json.Unmarshal(dataBytes, &message)
			if err != nil {
				return err
			}

			return a.handleChat(message)
		case "VOICE_STATE_UPDATE":
			log.Println("received voce_state_update event")
			return nil // TODO
		case "PRESENCE_UPDATE":
			log.Println("received present_update event")
			return nil // TODO?
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
