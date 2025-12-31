package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"garrison-stauffer.com/discord-bot/app"
	"garrison-stauffer.com/discord-bot/discord"
	"garrison-stauffer.com/discord-bot/discord/gateway"
	"garrison-stauffer.com/discord-bot/environment"
	"garrison-stauffer.com/discord-bot/youtube"
)

func main() {
	// Configure structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	slog.Info("starting discord bot server")
	serverComplete := make(chan error, 1)
	ctx := context.Background()

	var srv *http.Server
	go func() {
		mx := http.NewServeMux()

		port := environment.BindPort()
		slog.Info("http server starting", "port", port)
		srv = &http.Server{
			Addr:    fmt.Sprintf(":%s", port),
			Handler: mx,
		}
		mx.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
			slog.Info("received ping request")
			io.WriteString(w, "pong")
		})
		mx.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
			slog.Info("received healthcheck request")
			w.WriteHeader(200)
		})

		err := srv.ListenAndServe()
		serverComplete <- err
	}()

	slog.Info("initializing discord websocket connection")

	messages := make(chan gateway.Message, 4)
	botClient := discord.NewClient(
		discord.NewConfig(
			"wss://gateway.discord.gg/",
			environment.BotSecret(),
		),
		messages,
	)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	ytClient := youtube.NewClient(
		&http.Client{},
		environment.YtApiKey(),
	)

	err := botClient.Start()
	if err != nil {
		slog.Error("failed to start bot client", "error", err)
		os.Exit(1)
	}

	application := app.New(botClient, ytClient, environment.BotSecret())

	go func() {
		for {
			select {
			case msg := <-messages:
				if err := application.Handle(msg); err != nil {
					slog.Error("error processing message from channel", "error", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	select {
	case <-interrupt:
		slog.Info("received interrupt signal, starting graceful shutdown")
		if err := botClient.Shutdown(); err != nil {
			slog.Error("error shutting down bot client", "error", err)
		}

		if err := srv.Close(); err != nil {
			slog.Error("error shutting down http server", "error", err)
		}

		select {
		case <-serverComplete:
		case <-time.After(time.Second):
			return
		}
		slog.Info("shutdown complete")
	}
}
