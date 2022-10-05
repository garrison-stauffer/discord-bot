package gateway

func NewHeartbeat(sequence *int) *Message {
	return &Message{
		OpCode:     OpHeartbeat,
		SequenceId: sequence,
	}
}
