package ytapi

import (
	"github.com/buger/jsonparser"
	"net/url"
)

func Search(query string) ([]MediaItem, error) {
	reqUrl := "https://www.youtube.com/results?search_query=" + url.QueryEscape(query)
	body, err := Get(reqUrl)
	if err != nil {
		return nil, err
	}

	ytData := FindYouTubeData(body)

	results := make([]MediaItem, 0)

	_, err = jsonparser.ArrayEach([]byte(ytData), func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		videoRenderer, dataType, _, _ := jsonparser.Get(value, "videoRenderer")
		if dataType == jsonparser.Object {
			results = append(results, VideoRendererToMediaItem(videoRenderer, ""))
		}
	}, "contents", "twoColumnSearchResultsRenderer", "primaryContents", "sectionListRenderer", "contents", "[0]", "itemSectionRenderer", "contents")

	return results, err
}
