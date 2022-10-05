package client

import (
	"encoding/json"
	"fmt"
	"garrison-stauffer.com/discord-bot/discord/gateway"
	"garrison-stauffer.com/discord-bot/discord/gateway/intents"
	"log"
	"runtime/debug"
	"sync"
	"time"
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
}

func NewClient(config Config) Client {
	return &clientImpl{
		config: config,
		gatewayConfig: &gatewayConfig{
			reconnectUrl: config.gatewayUrl,
			sequence:     nil,
		},
		state:            ClientAwaitingStart,
		shutdown:         make(chan struct{}),
		receivedMessages: make(chan gateway.Message),
		outgoingMessages: make(chan gateway.Message),
	}
}

func (c *clientImpl) Start() error {
	err := c.locked(func() error {
		if c.state != ClientAwaitingStart {
			return fmt.Errorf("client already started")
		}
		log.Println("starting client")

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
	log.Println("attempting to acquire lock, dumping stacktrace")
	debug.PrintStack()

	c.mux.Lock()
	defer func() {
		c.mux.Unlock()
		log.Println("freed the lock, dumping stacktrace")
		debug.PrintStack()
	}()

	log.Println("acquired lock, dumping stacktrace")
	debug.PrintStack()

	return body()
}

func (c *clientImpl) Shutdown() error {
	log.Printf("Currently shutting down the client, current state is %v\n", c.state)
	return c.locked(func() error {
		if c.state == ClientClosed {
			return fmt.Errorf("cannot close client that's already started closing (current state %v)", c.state)
		}
		log.Println("shutting down")
		c.state = ClientClosed
		_ = c.session.Stop()
		close(c.receivedMessages)
		close(c.outgoingMessages)
		log.Println("finished shutting down")

		return nil
	})
}

func (c *clientImpl) readMessages() {
	for {
		done := false
		select {
		case msg, ok := <-c.receivedMessages:
			if !ok {
				log.Println("channel is closed")
				done = true
			}
			err := c.handle(msg)
			if err != nil {
				data, _ := json.Marshal(msg)
				log.Printf("error handling message %s", string(data))
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
			log.Printf("cannot reconnect when client state is %v\n", c.state)
		}

		log.Println("reconnecting client")

		sess := newSession(c, c.reconnect)
		err := sess.Start()

		if err != nil {
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
		log.Println("finished reconnecting")

		return nil
	})

	if err != nil {
		log.Printf("error reconnecting: %v\n", err)
	} else {
		log.Println("new client connected")
	}
}

func (c *clientImpl) Send(msg gateway.Message) error {
	c.outgoingMessages <- msg
	return nil
}

func (c *clientImpl) handle(msg gateway.Message) error {
	if msg.SequenceId != nil {
		c.gatewayConfig.sequence = msg.SequenceId
	}
	switch msg.OpCode {
	case gateway.OpHello:
		// I need to parse the Hello message from this? for now just read the json map
		requestedIntervalMs, ok := msg.Event["heartbeat_interval"].(float64)
		if !ok {
			return fmt.Errorf("could not get the heart_beat interval from %v", msg)
		}
		go func() {
			delay := int64(requestedIntervalMs * .87)
			ticker := time.NewTicker(time.Duration(delay) * time.Millisecond)
			log.Printf("Kicking off heartbeater with interval: %d\n", delay)

			// send initial heartbeat
			msg := gateway.NewHeartbeat(c.gatewayConfig.sequence)
			err := c.Send(*msg)
			if err != nil {
				log.Printf("error writing initial heartbeat %v\n", err)
				return
			}

			// send the Identify call
			botIntents := intents.BuildIntentPermissions(intents.VoiceStatus, intents.GuildMessageReactions, intents.GuildPresence, intents.GuildMessages, intents.MessageContent)
			msg = gateway.NewIdentify(botIntents, c.config.botSecretToken)
			err = c.Send(*msg)
			if err != nil {
				log.Printf("error writing identify message %v\n", err)
				return
			}

			for {
				select {
				case <-ticker.C:
					fmt.Println(c.gatewayConfig.sequence)
					msg := gateway.NewHeartbeat(c.gatewayConfig.sequence)
					log.Printf("sending heartbeat")
					_ = c.Send(*msg)
				}
			}
		}()
		return nil
	case gateway.OpHeartbeatAck:
		log.Println("received heartbeat ack")
		return nil
	case gateway.OpDispatch:
		log.Println("received dispatch message")

		switch *msg.Type {
		case "READY":
			reconnectUrl, ok := msg.Event["resume_gateway_url"].(string)
			if ok {
				c.gatewayConfig.reconnectUrl = reconnectUrl
			}
			break
		default:
			bytes, _ := json.Marshal(msg)
			log.Printf("unhandled dispatch type %s: %s", *msg.Type, string(bytes))
		}
		// TODO: switch based on the message type
		return nil
	default:
		return fmt.Errorf("unhandled opcode received (%d)", msg.OpCode)
	}
}
