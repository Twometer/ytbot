package ytapi

import (
	"github.com/buger/jsonparser"
	"strings"
)

func FindJsonData(data string, left string, right string) string {
	out := data[strings.Index(data, left+"{")+len(left):]
	out = out[:strings.Index(out, right)]
	return out
}

func FindYouTubeData(data string) string {
	return FindJsonData(data, "var ytInitialData = ", ";</script>")
}

func VideoRendererToMediaItem(videoRenderer []byte, fallbackId string) MediaItem {
	id, _ := jsonparser.GetString(videoRenderer, "videoId")
	name, _ := jsonparser.GetString(videoRenderer, "title", "runs", "[0]", "text")

	if len(id) == 0 {
		id = fallbackId
	}

	return MediaItem{
		Name: name,
		Url:  "https://youtube.com/watch?v=" + id,
	}
}
