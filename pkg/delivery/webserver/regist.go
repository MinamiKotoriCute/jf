package webserver

import (
	"io"
	"net/http"

	"github.com/MinamiKotoriCute/jf/pkg/delivery"
	"github.com/rotisserie/eris"
	"github.com/sirupsen/logrus"
)

// f type must be HandleFuncType
func (o *WebServer) RegistGetFunc(baseUrl string, f interface{}) {
	funcInfo, err := delivery.GetHandleFuncInfo(f)
	if err != nil {
		logrus.WithField("error", eris.ToJSON(err, true)).Fatal()
	}

	pattern := baseUrl + "/" + funcInfo.ReqName
	o.serveMux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		data := r.URL.Query().Get("data")
		ctx, err := o.GetHandleContextFunc(r)
		if err != nil {
			logrus.WithField("error", eris.ToJSON(err, true)).Error()
			return
		}

		rspData, err := o.handle(ctx, funcInfo, []byte(data))
		if err != nil {
			logrus.WithContext(ctx).WithField("error", eris.ToJSON(err, true)).Error()
			return
		}

		w.Header().Set("Accept", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		_, err = w.Write(rspData)
		if err != nil {
			logrus.WithContext(ctx).WithField("error", eris.ToJSON(err, true)).Error()
			return
		}
	})
}

// f type must be HandleFuncType
func (o *WebServer) RegistGetFuncs(baseUrl string, f ...interface{}) {
	for _, v := range f {
		o.RegistGetFunc(baseUrl, v)
	}
}

// f type must be HandleFuncType
func (o *WebServer) RegistPostFunc(baseUrl string, f interface{}) {
	funcInfo, err := delivery.GetHandleFuncInfo(f)
	if err != nil {
		logrus.WithField("error", eris.ToJSON(err, true)).Fatal()
	}

	pattern := baseUrl + "/" + funcInfo.ReqName
	o.serveMux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		data, err := io.ReadAll(r.Body)
		if err != nil {
			logrus.WithField("error", eris.ToJSON(err, true)).Error()
			return
		}

		ctx, err := o.GetHandleContextFunc(r)
		if err != nil {
			logrus.WithField("error", eris.ToJSON(err, true)).Error()
			return
		}

		rspData, err := o.handle(ctx, funcInfo, []byte(data))
		if err != nil {
			logrus.WithContext(ctx).WithField("error", eris.ToJSON(err, true)).Error()
			return
		}

		w.Header().Set("Accept", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		_, err = w.Write(rspData)
		if err != nil {
			logrus.WithContext(ctx).WithField("error", eris.ToJSON(err, true)).Error()
			return
		}
	})
}

// f type must be HandleFuncType
func (o *WebServer) RegistPostFuncs(baseUrl string, f ...interface{}) {
	for _, v := range f {
		o.RegistPostFunc(baseUrl, v)
	}
}

func (o *WebServer) RegistFileServer(baseUrl string, dir string) {
	o.serveMux.Handle(baseUrl, http.FileServer(http.Dir(dir)))
}

func (o *WebServer) RegistFileServerStripPrefix(baseUrl string, dir string) {
	o.serveMux.Handle(baseUrl, http.StripPrefix(baseUrl, http.FileServer(http.Dir(dir))))
}
