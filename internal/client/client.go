package client

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"os"
	"os/signal"
	"time"
)

func StartWebsocket() {
	c, _, err := websocket.DefaultDialer.Dial("wss://gateway.discord.gg", nil)
	if err != nil {
		log.Fatal(fmt.Errorf("error connecting to discord gateway: %v", err))
	}

	defer c.Close()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	socketClosed := make(chan struct{})

	go func() {
		defer close(socketClosed)
		for {
			msgType, msg, err := c.ReadMessage()
			if err != nil {
				log.Printf("error while reading a message %v\n", err)
				return
			}
			log.Printf("Received message of type %d with body %s", msgType, string(msg))
		}
	}()

	for {
		select {
		case <-socketClosed:
			return
		case <-interrupt:
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-socketClosed:
				log.Println("socket properly closed")
			case <-time.After(time.Second):
				log.Println("exiting without socket being closed, took longer than 1s")
			}
			return
		}
	}

}
