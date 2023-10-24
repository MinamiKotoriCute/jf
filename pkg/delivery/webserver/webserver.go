package webserver

import (
	"context"
	"net/http"
	"sync"

	"github.com/MinamiKotoriCute/jf/pkg/delivery"
	"github.com/rotisserie/eris"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

type GetHandleContextFuncType func(r *http.Request, funcInfo *delivery.HandleFuncInfo) (context.Context, error)
type CreateInternalErrorRspFuncType func(handleErr error, reqMessageName string) (proto.Message, bool, error)

type WebServer struct {
	mutex                      sync.Mutex
	httpServer                 *http.Server
	ServeMux                   *http.ServeMux
	OnHandleFinishedFunc       delivery.OnHandleFinishedFuncType
	GetHandleContextFunc       GetHandleContextFuncType
	CreateInternalErrorRspFunc CreateInternalErrorRspFuncType
}

func NewWebServer() *WebServer {
	return &WebServer{
		ServeMux:             http.NewServeMux(),
		GetHandleContextFunc: DefaultGetContext,
	}
}

func (o *WebServer) Start(addr string) error {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if o.httpServer != nil {
		return eris.New("already start")
	}

	o.httpServer = &http.Server{
		Addr:    addr,
		Handler: o.ServeMux,
	}

	go func() {
		if err := o.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			logrus.WithField("error", eris.ToJSON(err, true)).Error()
		}
	}()

	return nil
}

func (o *WebServer) Stop(ctx context.Context) error {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if o.httpServer == nil {
		return nil
	}

	if err := o.httpServer.Shutdown(ctx); err != nil {
		return eris.Wrapf(err, "shutdown fail")
	}

	return nil
}
