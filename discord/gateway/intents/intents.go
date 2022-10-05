package intents

type Intent int

var Guilds Intent = 1 << 0
var GuildMembers Intent = 1 << 1
var GuildBans Intent = 1 << 2
var GuildEmojiAndStickers Intent = 1 << 3
var GuildIntegrations Intent = 1 << 4
var GuildWebhooks Intent = 1 << 5
var GuildInvites Intent = 1 << 6
var VoiceStatus Intent = 1 << 7
var GuildPresence Intent = 1 << 8
var GuildMessages Intent = 1 << 9
var GuildMessageReactions Intent = 1 << 10
var GuildMessageTyping Intent = 1 << 11
var DirectMessage Intent = 1 << 12
var DirectMessageReactions Intent = 1 << 13
var DirectMessageTyping Intent = 1 << 14
var MessageContent Intent = 1 << 15
var GuildScheduledEvents Intent = 1 << 16
var AutoModConfiguration Intent = 1 << 20
var AutoModExecution Intent = 1 << 21

func BuildIntentPermissions(intents ...Intent) Intent {
	aggregator := 0
	for _, intent := range intents {
		aggregator |= int(intent)
	}
	return Intent(aggregator)
}

func (i Intent) Int() int {
	return int(i)
}
