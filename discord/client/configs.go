package client

// gatewayConfigs holds state from websocket connections, including sequence ids and reconnect URLs
type gatewayConfig struct {
	reconnectUrl string
	sequence     *int
}

type Config struct {
	gatewayUrl     string
	botSecretToken string
}

func NewConfig(gatewayUrl string, botToken string) Config {
	return Config{
		gatewayUrl:     gatewayUrl,
		botSecretToken: botToken,
	}
}
