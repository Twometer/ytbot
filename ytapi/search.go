package ytapi

import "net/url"

func Search(query string) ([]MediaItem, error) {
	reqUrl := "https://www.youtube.com/results?search_query=" + url.QueryEscape(query)
	_, err := Get(reqUrl)
	if err != nil {
		return nil, err
	}

	panic("to do")
}
