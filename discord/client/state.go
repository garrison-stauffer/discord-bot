package client

type ClientState string
type SessionState string

var (
	ClientAwaitingStart ClientState = "Initializing"
	ClientDisconnected  ClientState = "Connecting"
	ClientConnected     ClientState = "Connected"
	ClientClosed        ClientState = "Closed"

	SessionInitializing  SessionState = "Initializing"
	SessionConnected     SessionState = "Connected"
	SessionDisconnecting SessionState = "Disconnecting"
	SessionDisconnected  SessionState = "Disconnected"
)
