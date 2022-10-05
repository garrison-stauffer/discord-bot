package gateway

type Message struct {
	OpCode     int                    `json:"op"`
	Event      map[string]interface{} `json:"d"`
	SequenceId *int                   `json:"s"`
	Type       *string                `json:"t"`
}

var OpDispatch = 0
var OpHeartbeat = 1
var OpIdentify = 2
var OpPresenceUpdate = 3
var OpVoiceStateUpdate = 4
var OpResume = 6
var OpReconnect = 7
var OpRequestGuildMembers = 8
var OpInvalidSession = 9
var OpHello = 10
var OpHeartbeatAck = 11
