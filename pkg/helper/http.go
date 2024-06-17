package helper

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/MinamiKotoriCute/serr"
)

func GetHttpQueryInt64(r *http.Request, key string, defaultValue int64) (int64, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue, nil
	}

	if i, err := strconv.ParseInt(value, 10, 64); err != nil {
		return 0, serr.Wrap(err)
	} else {
		return i, nil
	}
}

func GetHttpQuerySliceInt64(r *http.Request, key string, defaultValue []int64, splitSeparators string) ([]int64, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue, nil
	}

	result := []int64{}
	for _, s := range strings.Split(value, splitSeparators) {
		if i, err := strconv.ParseInt(s, 10, 64); err != nil {
			return nil, serr.Wrap(err)
		} else {
			result = append(result, i)
		}
	}

	return result, nil
}

func GetHttpQueryString(r *http.Request, key string, defaultValue string) string {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}
	return value
}
