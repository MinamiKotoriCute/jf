package webserver

import (
	"context"
	"net/http"

	"github.com/MinamiKotoriCute/jf/pkg/delivery"
)

func DefaultGetContext(r *http.Request, funcInfo *delivery.HandleFuncInfo) (context.Context, ContextCloseFuncType, error) {
	return context.Background(), func() {}, nil
}
