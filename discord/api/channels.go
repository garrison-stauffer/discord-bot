package api

type Channel struct {
	Flags int    `json:"flags"`
	Id    string `json:"id"`
	Type  int    `json:"type"`
	Name  string `json:"name"`
}
