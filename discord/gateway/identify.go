package gateway

import "garrison-stauffer.com/discord-bot/discord/gateway/intents"

func NewIdentify(permissions intents.Intent, botSecret string) *Message {
	return &Message{
		OpCode:     OpIdentify,
		SequenceId: nil,
		Event: map[string]interface{}{
			"token": botSecret,
			"properties": map[string]interface{}{
				"os":      "linux",
				"browser": "github.com/garrison-stauffer/discord-bot",
				"device":  "github.com/garrison-stauffer/discord-bot",
			},
			"intents": permissions.Int(),
		},
	}
}
