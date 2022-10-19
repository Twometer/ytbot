package ytapi

import "net/url"

func GetVideo(id string) (MediaItem, error) {
	reqUrl := "https://www.youtube.com/watch?v=" + url.QueryEscape(id)
	_, err := Get(reqUrl)
	if err != nil {
		return MediaItem{}, err
	}

	return MediaItem{
		Name: "TODO",
		Url:  reqUrl,
	}, nil
}
