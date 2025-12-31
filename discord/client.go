package discord

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"
	"time"

	"garrison-stauffer.com/discord-bot/discord/gateway"
	"garrison-stauffer.com/discord-bot/discord/gateway/intents"
)

// Client is a wrapper around the
type Client interface {
	Start() error
	Send(message gateway.Message) error // allow users to send messages, e.g. if they had a Cron, we can re-use this for heartbeats + connections
	Shutdown() error
}

type clientImpl struct {
	config           Config
	gatewayConfig    *gatewayConfig
	mux              sync.Mutex
	state            ClientState
	shutdown         chan struct{}
	session          session
	receivedMessages chan gateway.Message
	outgoingMessages chan gateway.Message
	messages         chan<- gateway.Message
	heartbeater      Heartbeater
}

func NewClient(config Config, messages chan<- gateway.Message) Client {
	c := &clientImpl{
		config: config,
		gatewayConfig: &gatewayConfig{
			reconnectUrl: config.gatewayUrl,
			sequence:     nil,
		},
		state:            ClientAwaitingStart,
		shutdown:         make(chan struct{}),
		receivedMessages: make(chan gateway.Message),
		outgoingMessages: make(chan gateway.Message),
		messages:         messages,
	}

	c.heartbeater = NewHeartbeater(c)

	return c
}

func (c *clientImpl) Start() error {
	err := c.locked(func() error {
		if c.state != ClientAwaitingStart {
			return fmt.Errorf("client already started")
		}
		slog.Info("starting discord client")

		c.session = newSession(c, c.reconnect)
		err := c.session.Start()
		if err != nil {
			return err
		}

		c.state = ClientConnected
		go c.readMessages()

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *clientImpl) locked(body func() error) error {
	slog.Debug("attempting to acquire client lock")
	debug.PrintStack()

	c.mux.Lock()
	defer func() {
		c.mux.Unlock()
		slog.Debug("released client lock")
		debug.PrintStack()
	}()

	slog.Debug("acquired client lock")
	debug.PrintStack()

	return body()
}

func (c *clientImpl) Shutdown() error {
	slog.Info("shutting down discord client", "state", c.state)
	return c.locked(func() error {
		if c.state == ClientClosed {
			return fmt.Errorf("cannot close client that's already started closing (current state %v)", c.state)
		}
		slog.Info("executing shutdown")
		c.state = ClientClosed
		_ = c.session.Stop()
		close(c.receivedMessages)
		close(c.outgoingMessages)
		slog.Info("shutdown complete")

		return nil
	})
}

func (c *clientImpl) readMessages() {
	for {
		done := false
		select {
		case msg, ok := <-c.receivedMessages:
			if !ok {
				slog.Info("message channel closed")
				done = true
			}
			slog.Debug("received message", "opcode", msg.OpCode)
			err := c.handle(msg)
			if err != nil {
				data, _ := json.Marshal(msg)
				slog.Error("error handling message", "error", err, "message", string(data))
			}
		}

		if done {
			break
		}
	}

	c.shutdown <- struct{}{}
}

func (c *clientImpl) reconnect() {
	err := c.locked(func() error {
		if c.state != ClientDisconnected {
			slog.Warn("cannot reconnect in current state", "state", c.state)
		}

		slog.Info("reconnecting discord client")

		// hack: reset sequence + reconnect URL to skip reconnects
		c.gatewayConfig.sequence = nil
		c.gatewayConfig.reconnectUrl = "wss://gateway.discord.gg/"

		sess := newSession(c, c.reconnect)
		err := sess.Start()

		if err != nil {
			sess.Stop()
			go func() {
				select {
				case <-time.After(5 * time.Second):
					c.reconnect()
				}
			}()
			return fmt.Errorf("scheduled new reconnect in 5 seconds")
		}

		c.session = sess
		c.state = ClientConnected
		slog.Info("reconnection successful")

		return nil
	})

	if err != nil {
		slog.Error("error reconnecting", "error", err)
	} else {
		slog.Info("new client connected")
	}
}

func (c *clientImpl) Send(msg gateway.Message) error {
	c.outgoingMessages <- msg
	return nil
}

func (c *clientImpl) handle(msg gateway.Message) error {
	if msg.Type != nil {
		slog.Debug("handling message", "opcode", msg.OpCode, "type", *msg.Type)
	} else {
		slog.Debug("handling message", "opcode", msg.OpCode)
	}

	if msg.SequenceId != nil {
		c.gatewayConfig.sequence = msg.SequenceId
	}
	switch msg.OpCode {
	case gateway.OpHello:
		requestedIntervalMs, ok := (*msg.Event)["heartbeat_interval"].(float64)
		if !ok {
			return fmt.Errorf("could not get the heart_beat interval from %v", msg)
		}
		go func() {
			// send initial heartbeat
			msg := gateway.NewHeartbeat(c.gatewayConfig.sequence)
			err := c.Send(*msg)
			if err != nil {
				slog.Error("error writing initial heartbeat", "error", err)
				return
			}

			// start off heartbeater in the background that can be stopped on disconnect
			delay := int64(requestedIntervalMs * .87)
			slog.Info("starting heartbeater", "interval_ms", delay)
			go c.heartbeater.Start(time.Duration(delay) * time.Millisecond)

			botIntents := intents.BuildIntentPermissions(intents.VoiceStatus, intents.Guilds, intents.GuildMembers, intents.GuildMessageReactions, intents.GuildPresence, intents.GuildMessages, intents.MessageContent)
			msg = gateway.NewIdentify(botIntents, c.config.botSecretToken)
			err = c.Send(*msg)
			if err != nil {
				slog.Error("error writing identify message", "error", err)
				return
			}
		}()
		return nil
	case gateway.OpHeartbeatAck:
		return nil
	case gateway.OpReconnect:
		slog.Info("received reconnect opcode, stopping current session")
		_ = c.session.Stop()
		time.Sleep(5 * time.Second)
		c.reconnect()
		return nil
	case gateway.OpInvalidSession:
		slog.Info("received invalid session, stopping current session")
		_ = c.session.Stop()
		time.Sleep(5 * time.Second)
		c.reconnect()
		return nil
	case gateway.OpDispatch:
		switch *msg.Type {
		case "READY":
			reconnectUrl, ok := (*msg.Event)["resume_gateway_url"].(string)
			sessionId, ok := (*msg.Event)["session_id"].(string)
			if ok {
				c.gatewayConfig.reconnectUrl = reconnectUrl
				c.gatewayConfig.sessionId = sessionId
			}
			break
		}
	}

	c.messages <- msg
	return nil
}
