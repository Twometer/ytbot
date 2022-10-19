package ytapi

import (
	"github.com/buger/jsonparser"
	"net/url"
)

func GetVideo(id string) (MediaItem, error) {
	reqUrl := "https://www.youtube.com/watch?v=" + url.QueryEscape(id)
	body, err := Get(reqUrl)
	if err != nil {
		return MediaItem{}, err
	}

	ytData := FindYouTubeData(body)

	renderer, _, _, err := jsonparser.Get([]byte(ytData), "contents", "twoColumnWatchNextResults", "results", "results", "contents", "[0]", "videoPrimaryInfoRenderer")
	if err != nil {
		return MediaItem{}, err
	}
	item := VideoRendererToMediaItem(renderer, id)

	return item, nil
}
