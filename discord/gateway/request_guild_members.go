package gateway

func NewRequestGuildMembers(sequenceId *int, guildId string) *Message {
	return &Message{
		OpCode:     OpRequestGuildMembers,
		SequenceId: sequenceId,
		Event: &map[string]interface{}{
			"guild_id":  guildId,
			"query":     "",
			"limit":     0,
			"presences": true,
		},
	}
}
