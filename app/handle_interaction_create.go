package app

import (
	"context"
	"fmt"
	"garrison-stauffer.com/discord-bot/discord/api"
	"log"
	"net/http"
	"runtime/debug"
)

func (a *App) handleInteractionCreate(msg api.InteractionCreate) error {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("ERROR! Panic caught during interaction handling (%+v):  %v\n", msg, err)
		}
	}()

	var response = "uh oh, something went wrong"
	ctx := context.Background()
	switch msg.Command.Name {
	case "pal_up":
		meta := a.palworld.GetServerInfoIfReady(ctx)
		if meta != nil {
			response = fmt.Sprintf("server is already running, connect at `%s:8211`", *meta.PublicIp)
		} else {
			response = "the server is starting, I will post a message when it is ready"
			go func() {
				defer func() {
					if err := recover(); err != nil {
						fmt.Printf("panic while starting the server: %v\n", err)
						debug.PrintStack()
					}
				}()

				meta, err := a.palworld.StartServer(ctx)
				if err == nil {
					serverReady, _ := api.NewMessage(msg.Channel, a.botSecret, fmt.Sprintf("server is ready, connect to `%s:8211`", *meta.PublicIp))
					http.DefaultClient.Do(serverReady)
				} else {
					serverReady, _ := api.NewMessage(msg.Channel, a.botSecret, fmt.Sprintf("error while starting the server, bother garrison: %v", err))
					http.DefaultClient.Do(serverReady)
				}
			}()
		}
	case "pal_down":
		response = "Okay, I will stop the server."

		go func() {
			defer func() {
				if err := recover(); err != nil {
					fmt.Printf("panic while shutting down the server: %v\n", err)
					debug.PrintStack()
				}
			}()

			err := a.palworld.StopServer(ctx)
			if err != nil {
				serverReady, _ := api.NewMessage(msg.Channel, a.botSecret, fmt.Sprintf("error while shutting down the server: %v", err))
				http.DefaultClient.Do(serverReady)
			} else {
				serverReady, _ := api.NewMessage(msg.Channel, a.botSecret, "server shutdown completed")
				http.DefaultClient.Do(serverReady)
			}
		}()
	case "pal_where":
		meta := a.palworld.GetServerInfoIfReady(ctx)

		if meta == nil {
			response = "server is not running, please use `/pal_up` to start it"
		} else {
			response = fmt.Sprintf("server is already running, connect at `%s:8211`", *meta.PublicIp)
		}

		fmt.Println("I should print the server IP when I can")
	default:
		return fmt.Errorf("unhandled interaction type %s", msg.Command.Name)
	}

	req, err := api.NewAcknowledgeInteraction(msg.Id, msg.ContinuationToken, a.botSecret, response)
	if err != nil {
		return fmt.Errorf("error formatting ack request: %w", err)
	}
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("error while acknowledging the interaction %v\n", err)
		return err
	}

	return nil
}
