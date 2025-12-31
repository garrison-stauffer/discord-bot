package youtube

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
)

type Client interface {
	IsMusicVideo(url string) (bool, error)
}

type clientImpl struct {
	client   *http.Client
	apiToken string
}

type ytResponse struct {
	Items []itemJson `json:"items"`
}

type itemJson struct {
	Snippet snippetJson `json:"snippet"`
}

type snippetJson struct {
	Category string `json:"categoryId"`
}

func NewClient(client *http.Client, apiToken string) Client {
	return &clientImpl{
		client,
		apiToken,
	}
}

func (c *clientImpl) IsMusicVideo(ytUrl string) (bool, error) {
	slog.Debug("checking if URL is music video", "url", ytUrl)
	url, err := url.Parse(ytUrl)
	if err != nil {
		return false, err
	}

	vId, ok := url.Query()["v"]
	if !ok || len(vId) != 1 {
		return false, fmt.Errorf("unable to parse video id from %s", ytUrl)
	}
	slog.Debug("extracted video ID", "video_id", vId[0])

	resp, err := c.client.Get(fmt.Sprintf("https://youtube.googleapis.com/youtube/v3/videos?part=snippet&id=%s&key=%s", vId[0], c.apiToken))
	if err != nil {
		return false, fmt.Errorf("error calling youtube API: %v", err)
	}
	if resp.StatusCode != 200 {
		return false, fmt.Errorf("unexpected yt status %d", resp.StatusCode)
	}

	var response ytResponse
	err = json.NewDecoder(resp.Body).Decode(&response)

	if err != nil {
		return false, fmt.Errorf("error parsing youtube response: %v", err)
	}

	category := response.Items[0].Snippet.Category
	isMusic := category == "10"
	slog.Debug("video category check", "video_id", vId[0], "category", category, "is_music", isMusic)

	return isMusic, nil
}
