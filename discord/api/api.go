package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
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

func NewChatReact(cId, mId, react, botSecret string) (*http.Request, error) {
	url := fmt.Sprintf("%s/channels/%s/messages/%s/reactions/%s/@me", "https://discord.com/api", cId, mId, react)
	slog.Debug("creating react request", "url", url)
	req, err := http.NewRequest("PUT", url, nil)
	req.Header.Set("Authorization", "Bot "+botSecret)
	req.Header.Set("User-Agent", "DiscordBot (discord-bot.garrison-stauffer.com, 0.0.1)")
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		return nil, err
	}

	return req, nil
}

func NewDeleteChatReact(cId, mId, react, botSecret string) (*http.Request, error) {
	url := fmt.Sprintf("%s/channels/%s/messages/%s/reactions/%s/@me", "https://discord.com/api", cId, mId, react)
	slog.Debug("creating delete react request", "url", url)
	req, err := http.NewRequest("DELETE", url, nil)
	req.Header.Set("Authorization", "Bot "+botSecret)
	req.Header.Set("User-Agent", "DiscordBot (discord-bot.garrison-stauffer.com, 0.0.1)")
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		return nil, err
	}

	return req, nil
}

func NewMessage(cId, botSecret, message string) (*http.Request, error) {
	url := fmt.Sprintf("%s/channels/%s/messages", "https://discord.com/api", cId)

	slog.Debug("creating message request", "url", url)

	requestBody := map[string]interface{}{
		"content": message,
		"tts":     true,
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bot "+botSecret)
	req.Header.Set("User-Agent", "DiscordBot (discord-bot.garrison-stauffer.com, 0.0.1)")
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func NewAcknowledgeInteraction(iId, iToken, botSecret, message string) (*http.Request, error) {
	url := fmt.Sprintf("%s/interactions/%s/%s/callback", "https://discord.com/api", iId, iToken)
	slog.Debug("creating acknowledge interaction request", "url", url)

	requestBody := map[string]interface{}{
		"type": 4,
		"data": map[string]interface{}{
			"content": message,
		},
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bot "+botSecret)
	req.Header.Set("User-Agent", "DiscordBot (discord-bot.garrison-stauffer.com, 0.0.1)")
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}
