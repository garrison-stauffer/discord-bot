package environment

import (
	"fmt"
	"os"
)

func BotSecret() string {
	s := os.Getenv("DISCORD_BOT_SECRET")
	if s == "" {
		panic("required environment variable DISCORD_BOT_SECRET not present")
	}
	fmt.Println(s)
	return s
}

func BindPort() string {
	return os.Getenv("PORT")
}
