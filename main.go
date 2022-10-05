package main

import (
	"fmt"
	"garrison-stauffer.com/discord-bot/discord/client"
	"garrison-stauffer.com/discord-bot/environment"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	log.Println("Started the server")
	serverComplete := make(chan error, 1)

	var srv *http.Server
	go func() {
		mx := http.NewServeMux()

		srv = &http.Server{
			Addr:    ":8080",
			Handler: mx,
		}
		mx.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
			log.Println("Received message on /ping")
			io.WriteString(w, "Doo doo")
		})
		mx.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
			log.Println("Received message on /healthcheck")
			w.WriteHeader(200)
		})

		err := srv.ListenAndServe()
		serverComplete <- err
	}()

	fmt.Println("Starting websocket")

	botClient := client.NewClient(
		client.NewConfig(
			"wss://gateway.discord.gg/",
			environment.BotSecret(),
		),
	)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	err := botClient.Start()
	if err != nil {
		log.Fatalf("could not start bot client: %v", err)
	}

	select {
	case <-interrupt:
		log.Println("received interrupt, starting to shut down")
		err := botClient.Shutdown()
		if err != nil {
			log.Printf("error shutting down bot client: %v", err)
		}
		srv.Close()

		select {
		case <-serverComplete:
		case <-time.After(time.Second):
			return
		}
		log.Println("shutdown complete")
	}
}
