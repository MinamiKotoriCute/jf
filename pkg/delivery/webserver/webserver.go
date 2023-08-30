package webserver

import (
	"context"
	"net/http"
	"sync"

	"github.com/golang/glog"
	"github.com/rotisserie/eris"
)

type WebServer struct {
	mutex      sync.Mutex
	httpServer *http.Server
	serveMux   *http.ServeMux
}

func NewWebServer() *WebServer {
	return &WebServer{
		serveMux: http.NewServeMux(),
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
			glog.Fatalf("ListenAndServe fail. err:%v", err)
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

	glog.Infof("webserver stop")
	if err := o.httpServer.Shutdown(ctx); err != nil {
		return eris.Wrapf(err, "shutdown fail")
	}

	glog.Infof("webserver stop end")

	return nil
}
