package config

type Key string

const (
	KeyAuthToken      = "YTB_AUTH_TOKEN"
	KeyFfmpegLocation = "YTB_FFMPEG_LOCATION"
)

func init() {
	loadKey(KeyAuthToken, "")
	loadKey(KeyFfmpegLocation, "")
}
