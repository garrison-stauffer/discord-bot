package main

import (
	"fmt"
	"garrison-stauffer.com/discord-bot/internal/client"
	"io"
	"log"
	"net/http"
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
	client.StartWebsocket()
	srv.Close()
	select {
	case <-serverComplete:
	case <-time.After(time.Second):
		return
	}
	fmt.Println("Websocket terminated, shutting down")
}
