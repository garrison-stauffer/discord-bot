package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func NewTimeoutRequest(gId, mId string, botSecret string) (*http.Request, error) {
	requestBody := map[string]interface{}{
		"communication_disabled_until": time.
			Now().
			Add(45 * time.Second).
			Format(time.RFC3339),
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/guilds/%s/members/%s", "https://discord.com/api", gId, mId)
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bot "+botSecret)
	req.Header.Set("User-Agent", "DiscordBot (discord-bot.garrison-stauffer.com, 0.0.1)")
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		return nil, err
	}

	return req, nil
}

func NewMusicVideoReact(cId, mId, react, botSecret string) (*http.Request, error) {
	url := fmt.Sprintf("%s/channels/%s/messages/%s/reactions/%s/@me", "https://discord.com/api", cId, mId, react)
	fmt.Println("attempting to send repost react to", url)
	req, err := http.NewRequest("PUT", url, nil)
	req.Header.Set("Authorization", "Bot "+botSecret)
	req.Header.Set("User-Agent", "DiscordBot (discord-bot.garrison-stauffer.com, 0.0.1)")
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		return nil, err
	}

	return req, nil
}
