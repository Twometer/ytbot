package ytapi

import (
	"errors"
	"net/url"
)

func LoadMediaItems(query string) ([]MediaItem, error) {
	urlVal, err := url.Parse(query)
	if err != nil || len(urlVal.Hostname()) == 0 {
		// not a 'real' url, could be search query
		results, err := Search(query)
		if err != nil {
			return nil, err
		}

		if len(results) == 0 {
			return []MediaItem{}, nil
		} else {
			return []MediaItem{results[0]}, nil
		}
	} else if urlVal.Query().Has("list") {
		items, err := GetPlaylistItems(urlVal.Query().Get("list"))
		if err != nil {
			return nil, err
		} else {
			return items, nil
		}
	} else if urlVal.Query().Has("v") {
		video, err := GetVideo(urlVal.Query().Get("v"))
		if err != nil {
			return nil, err
		}
		return []MediaItem{video}, nil
	} else if urlVal.Hostname() == "youtu.be" {
		video, err := GetVideo(urlVal.Path[1:])
		if err != nil {
			return nil, err
		}
		return []MediaItem{video}, nil
	} else {
		return nil, errors.New("this query is not supported")
	}
}
