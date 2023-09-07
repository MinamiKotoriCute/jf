package webserver

import (
	"context"
	"net/http"
)

func DefaultGetContext(r *http.Request) (context.Context, error) {
	return context.Background(), nil
}
