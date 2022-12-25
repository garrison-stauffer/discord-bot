package channels

type Type int

const GuildText Type = 0
const DirectMessage Type = 0
const GuildVoice Type = 2
const GroupDirectMessage Type = 3
const GuildCategory Type = 4
const GuildAnnouncement Type = 5
const AnnouncementThread Type = 10
const PublicThread Type = 11
const PrivateThread Type = 12
const GuildStageVoice Type = 13
const GuildDirectory Type = 14
const GuildForum Type = 15

func TypeFrom(t int) Type {
	return Type(t)
}
