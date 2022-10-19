package ytapi

import (
	"github.com/buger/jsonparser"
	"net/url"
)

func GetPlaylistItems(playlistId string) ([]MediaItem, error) {
	reqUrl := "https://youtube.com/playlist?list=" + url.QueryEscape(playlistId)
	body, err := Get(reqUrl)
	if err != nil {
		return nil, err
	}

	ytData := FindYouTubeData(body)

	results := make([]MediaItem, 0)
	_, err = jsonparser.ArrayEach([]byte(ytData), func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		videoRenderer, dataType, _, _ := jsonparser.Get(value, "playlistVideoRenderer")
		if dataType == jsonparser.Object {
			results = append(results, VideoRendererToMediaItem(videoRenderer, ""))
		}
	}, "contents", "twoColumnBrowseResultsRenderer", "tabs", "[0]", "tabRenderer", "content", "sectionListRenderer", "contents", "[0]", "itemSectionRenderer", "contents", "[0]", "playlistVideoListRenderer", "contents")

	return results, err
}
