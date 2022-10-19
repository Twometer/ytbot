package utils

import "time"

func Timestamp() uint32 {
	return uint32(time.Now().UnixNano() / int64(time.Millisecond))
}
