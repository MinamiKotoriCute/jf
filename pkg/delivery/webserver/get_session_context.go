package webserver

import (
	"context"
	"net/http"

	"github.com/MinamiKotoriCute/jf/pkg/delivery"
)

func DefaultGetContext(r *http.Request, funcInfo *delivery.HandleFuncInfo) (context.Context, error) {
	return context.Background(), nil
}
