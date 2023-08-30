package helper

import (
	"os"
	"strconv"

	"github.com/rotisserie/eris"
)

func GetOsEnvString(key string, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

func GetOsEnvInt(key string, defaultValue int) int {
	if value, ok := os.LookupEnv(key); ok {
		if num, err := strconv.Atoi(value); err == nil {
			return num
		} else {
			panic(eris.Wrapf(err, "key:%s value:%s", key, value))
		}
	}
	return defaultValue
}

func GetOsEnvInt64(key string, defaultValue int64) int64 {
	if value, ok := os.LookupEnv(key); ok {
		if num, err := strconv.ParseInt(value, 10, 64); err == nil {
			return num
		} else {
			panic(eris.Wrapf(err, "key:%s value:%s", key, value))
		}
	}
	return defaultValue
}

func GetOsEnvBool(key string, defaultValue bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		} else {
			panic(eris.Wrapf(err, "key:%s value:%s", key, value))
		}
	}
	return defaultValue
}
