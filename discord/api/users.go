package api

type User struct {
	Id     string `json:"id"`
	Name   string `json:"username"`
	Avatar string `json:"avatar"`
	IsBot  bool   `json:"bot"`
}

type UserRef struct {
	Id string `json:"id"`
}
