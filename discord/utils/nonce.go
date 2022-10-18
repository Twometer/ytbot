package utils

import (
	"math/rand"
	"strconv"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func NewNonce() string {
	return strconv.FormatUint(rand.Uint64(), 10)
}
