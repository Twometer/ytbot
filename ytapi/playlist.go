package ytapi

import "net/url"

func GetPlaylistItems(playlistId string) ([]MediaItem, error) {
	reqUrl := "https://youtube.com/playlist?list=" + url.QueryEscape(playlistId)
	_, err := Get(reqUrl)
	if err != nil {
		return nil, err
	}

	panic("to do")
}
