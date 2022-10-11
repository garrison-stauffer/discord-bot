package main

import (
	"fmt"
	"garrison-stauffer.com/discord-bot/app"
	"garrison-stauffer.com/discord-bot/discord"
	"garrison-stauffer.com/discord-bot/discord/gateway"
	"garrison-stauffer.com/discord-bot/environment"
	"garrison-stauffer.com/discord-bot/youtube"
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

		fmt.Println("port used is ", environment.BindPort())
		srv = &http.Server{
			Addr:    fmt.Sprintf(":%s", environment.BindPort()),
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

	messages := make(chan gateway.Message, 4)
	botClient := discord.NewClient(
		discord.NewConfig(
			//"wss://gateway.discord.gg/",
			"wss://gateway-us-east1-c.discord.gg",
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

	isVid, err := ytClient.IsMusicVideo("https://www.youtube.com/watch?v=ldN9fNhZcsQ")
	if err != nil {
		panic(err)
	}
	println(isVid)

	err = botClient.Start()

	app := app.New(botClient, ytClient, environment.BotSecret())

	go func() {
		for {
			select {
			case msg := <-messages:
				err := app.Handle(msg)
				if err != nil {
					log.Printf("error processing message from channel %v", err)
				}
			}
		}
	}()

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
