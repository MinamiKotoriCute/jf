package helper

import (
	"net/http"
	"strconv"

	"github.com/rotisserie/eris"
)

func GetHttpQueryInt64(r *http.Request, key string, defaultValue int64) (int64, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue, nil
	}

	if i, err := strconv.ParseInt(value, 10, 64); err != nil {
		return 0, eris.Wrap(err, "")
	} else {
		return i, nil
	}
}
