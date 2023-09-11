package webserver

import (
	"context"
	"net/http"
	"sync"

	"github.com/MinamiKotoriCute/jf/pkg/delivery"
	"github.com/rotisserie/eris"
	"github.com/sirupsen/logrus"
)

type GetHandleContextFuncType func(r *http.Request) (context.Context, error)

type WebServer struct {
	mutex                sync.Mutex
	httpServer           *http.Server
	serveMux             *http.ServeMux
	OnHandleFinished     delivery.OnHandleFinishedFuncType
	GetHandleContextFunc GetHandleContextFuncType
}

func NewWebServer() *WebServer {
	return &WebServer{
		serveMux:             http.NewServeMux(),
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
		Handler: o.serveMux,
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
