package config

type Key string

const (
	KeyAuthToken = "YTB_AUTH_TOKEN"
)

func init() {
	loadKey(KeyAuthToken, "")
}
