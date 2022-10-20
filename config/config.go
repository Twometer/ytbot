package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var configValues map[Key]string

func init() {
	configValues = make(map[Key]string)
}

func loadKey(key Key, defaultValue string) {
	value, exists := os.LookupEnv(string(key))
	if !exists {
		if defaultValue != "" {
			value = defaultValue
		} else {
			log.Fatalf("missing environment variable `%s`", key)
		}
	}
	configValues[key] = value
}

//goland:noinspection GoUnusedExportedFunction
func GetBool(key Key) bool {
	return strings.ToLower(configValues[key]) == "true"
}

//goland:noinspection GoUnusedExportedFunction
func GetString(key Key) string {
	return configValues[key]
}

//goland:noinspection GoUnusedExportedFunction
func GetInt(key Key) int {
	val, err := strconv.Atoi(configValues[key])
	if err != nil {
		log.Fatalf("failed to read int property `%s`: %s", key, err.Error())
	}
	return val
}

//goland:noinspection GoUnusedExportedFunction
func GetMilliseconds(key Key) time.Duration {
	return time.Millisecond * time.Duration(GetInt(key))
}
