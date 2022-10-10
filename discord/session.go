package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"garrison-stauffer.com/discord-bot/discord/gateway"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

type session interface {
	Start() error
	Stop() error
	SendMessage(msg gateway.Message) error
	State() SessionState
}

type sessionImpl struct {
	state        SessionState
	connection   *websocket.Conn
	shutdown     chan struct{}
	client       *clientImpl
	onDisconnect func()
}

func newSession(client *clientImpl, onDisconnect func()) session {
	return &sessionImpl{
		state:        SessionInitializing,
		connection:   nil, // initialized in Start()
		shutdown:     make(chan struct{}, 10),
		client:       client,
		onDisconnect: onDisconnect,
	}
}

func (s *sessionImpl) Start() error {
	conn, err := s.connectToGateway()
	if err != nil {
		return err
	}

	s.connection = conn
	s.state = SessionConnected

	go s.startHandlingMessages()

	return nil
}

func (s *sessionImpl) Stop() error {
	s.state = SessionDisconnecting
	s.shutdown <- struct{}{}

	// should I listen for a shut down signal from the reader?
	return nil
}

func (s *sessionImpl) SendMessage(msg gateway.Message) error {
	if s.state == SessionDisconnecting || s.state == SessionDisconnected {
		return fmt.Errorf("cannot send message while session is disconnecting")
	}
	bytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("error creating json message %v: %v", err, msg)
	}
	log.Printf("sending message? %s", string(bytes))
	return s.connection.WriteMessage(1, bytes)
}

func (s *sessionImpl) State() SessionState {
	return s.state
}

func (s *sessionImpl) connectToGateway() (*websocket.Conn, error) {
	conn, _, err := websocket.DefaultDialer.Dial(s.client.gatewayConfig.reconnectUrl, nil)
	return conn, err
}

func (s *sessionImpl) startHandlingMessages() {
	socketClosed := make(chan struct{})
	log.Println("starting to handle messages")

	go func() {
		defer close(socketClosed)
		for { // iterate as long as messages come in
			msgType, msgJson, err := s.connection.ReadMessage()
			if err != nil {
				// exit loop so that we close socketClosed, and can go forward with a Reconnect
				log.Printf("error while reading a message - going to restart connection: %v. \n", err)
				return
			}

			log.Printf("Received message of type %d with body %s", msgType, string(msgJson))
			decoder := json.NewDecoder(bytes.NewReader(msgJson))
			var msg gateway.Message
			err = decoder.Decode(&msg)
			if err != nil {
				log.Printf("Error decoding msg json %s: %v", string(msgJson), err)
				continue
			}

			s.client.receivedMessages <- msg
		}
	}()

	for {
		select {
		case msg := <-s.client.outgoingMessages:
			log.Printf("sending message %v", msg)
			err := s.SendMessage(msg)
			if err != nil {
				log.Printf("error sending message off channel, requeuing %v", err)
				_ = s.client.Send(msg)
			}
		case <-socketClosed:
			fmt.Println("received message from session being disconnected")
			s.state = SessionDisconnected
			s.onDisconnect()
			return
		case <-s.shutdown:
			s.state = SessionDisconnected

			fmt.Println("received message from shutdown channel")
			err := s.connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Printf("error closing connection, going to mark client as closed anyways: %v\n", err)
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
