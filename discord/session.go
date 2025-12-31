package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"garrison-stauffer.com/discord-bot/discord/gateway"
	"github.com/gorilla/websocket"
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
	slog.Debug("sending message", "message", string(bytes))
	return s.connection.WriteMessage(1, bytes)
}

func (s *sessionImpl) State() SessionState {
	return s.state
}

func (s *sessionImpl) connectToGateway() (*websocket.Conn, error) {
	slog.Info("connecting to gateway", "url", s.client.gatewayConfig.reconnectUrl)
	conn, _, err := websocket.DefaultDialer.Dial(s.client.gatewayConfig.reconnectUrl, nil)
	return conn, err
}

func (s *sessionImpl) startHandlingMessages() {
	socketClosed := make(chan struct{})
	slog.Info("starting message handler")

	go func() {
		defer close(socketClosed)
		for {
			msgType, msgJson, err := s.connection.ReadMessage()
			if err != nil {
				slog.Error("error reading message, restarting connection", "error", err)
				_ = s.connection.Close()
				return
			}

			slog.Debug("received raw message", "type", msgType, "body", string(msgJson))
			decoder := json.NewDecoder(bytes.NewReader(msgJson))
			var msg gateway.Message
			err = decoder.Decode(&msg)
			if err != nil {
				slog.Error("error decoding message json", "error", err, "json", string(msgJson))

				var msgV2 gateway.WeirdMessage
				decoder = json.NewDecoder(bytes.NewReader(msgJson))
				err = decoder.Decode(&msgV2)
				if err != nil {
					slog.Error("error decoding weird message json", "error", err, "json", string(msgJson))
					continue
				}

				msg = gateway.Message{
					OpCode:     msgV2.OpCode,
					Type:       msgV2.Type,
					SequenceId: msgV2.SequenceId,
					Event:      &map[string]interface{}{},
				}
			}

			s.client.receivedMessages <- msg
		}
	}()

	for {
		select {
		case msg := <-s.client.outgoingMessages:
			slog.Debug("sending outgoing message", "opcode", msg.OpCode)
			err := s.SendMessage(msg)
			if err != nil {
				slog.Error("error sending message, requeuing", "error", err)
				_ = s.client.Send(msg)
			}
		case <-socketClosed:
			slog.Info("session disconnected, reconnecting in 5s")
			s.state = SessionDisconnected
			time.Sleep(5 * time.Second)
			s.onDisconnect()
			return
		case <-s.shutdown:
			s.state = SessionDisconnected

			slog.Info("received shutdown signal")
			err := s.connection.Close()
			if err != nil {
				slog.Error("error closing connection, marking as closed anyway", "error", err)
			}
			select {
			case <-socketClosed:
				slog.Info("socket properly closed")
			case <-time.After(time.Second):
				slog.Warn("socket close timeout after 1s")
			}
			return
		}
	}

}
