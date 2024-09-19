package helper

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

func GetOsEnvString(key string, defaultValue string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	return value
}

func GetOsEnvInt(key string, defaultValue int) int {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	num, err := strconv.Atoi(value)
	if err != nil {
		slog.Warn("GetOsEnvInt failed",
			slog.String("key", key),
			slog.String("value", value),
			slog.Any("err", err))
		return defaultValue
	}

	return num
}

func GetOsEnvInt64(key string, defaultValue int64) int64 {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	num, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		slog.Warn("GetOsEnvInt64 failed",
			slog.String("key", key),
			slog.String("value", value),
			slog.Any("err", err))
		return defaultValue
	}

	return num
}

func GetOsEnvUInt64(key string, defaultValue uint64) uint64 {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	num, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		slog.Warn("GetOsEnvUInt64 failed",
			slog.String("key", key),
			slog.String("value", value),
			slog.Any("err", err))
		return defaultValue
	}

	return num
}

func GetOsEnvBool(key string, defaultValue bool) bool {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	b, err := strconv.ParseBool(value)
	if err != nil {
		slog.Warn("GetOsEnvBool failed",
			slog.String("key", key),
			slog.String("value", value),
			slog.Any("err", err))
		return defaultValue
	}

	return b
}

func GetOsEnvSliceString(key string, defaultValue []string, separator string) []string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	return strings.Split(value, separator)
}

func GetOsEnvMapKeyString(key string, defaultValue map[string]struct{}, separator string) map[string]struct{} {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	m := make(map[string]struct{})
	for _, v := range strings.Split(value, separator) {
		m[v] = struct{}{}
	}

	return m
}

func GetOsEnvTime(key string, defaultValue time.Time, layout string) time.Time {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	t, err := time.Parse(layout, value)
	if err != nil {
		slog.Warn("GetOsEnvTime failed",
			slog.String("key", key),
			slog.String("value", value),
			slog.Any("err", err))
		return defaultValue
	}

	return t
}
