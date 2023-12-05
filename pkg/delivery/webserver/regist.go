package webserver

import (
	"errors"
	"io"
	"net/http"

	"github.com/MinamiKotoriCute/jf/pkg/delivery"
	"github.com/MinamiKotoriCute/serr"
	"github.com/sirupsen/logrus"
)

// f type must be HandleFuncType
func (o *WebServer) RegistFunc(baseUrl string, httpMethod HttpMethod, middlewares []MiddlewareFuncType, f interface{}) {
	funcInfo, err := delivery.GetHandleFuncInfo(f)
	if err != nil {
		logrus.WithField("error", serr.ToJSON(err, true)).Fatal()
	}

	pattern := baseUrl + "/" + funcInfo.ReqName
	o.ServeMux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		if !httpMethod.Match(r.Method) {
			http.Error(w, "method not support", http.StatusBadRequest)
			return
		}

		data := ""
		if r.Method == "GET" {
			data = r.URL.Query().Get("data")
		} else if r.Method == "POST" {
			defer r.Body.Close()
			tempData, err := io.ReadAll(r.Body)
			if err != nil {
				logrus.WithField("error", serr.ToJSON(err, true)).Error()
				http.Error(w, "error", http.StatusBadRequest)
				return
			}

			data = string(tempData)

			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", httpMethod.ToString())
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		} else {
			http.Error(w, "method not support", http.StatusBadRequest)
			return
		}

		for _, v := range middlewares {
			isContinus := v(w, r, funcInfo)
			if !isContinus {
				return
			}
		}

		if r.Method == "OPTIONS" {
			return
		}

		ctx, ctxCloseFunc, err := o.GetHandleContextFunc(r, funcInfo)
		if err != nil {
			var httpErr *HttpError
			if errors.As(err, &httpErr) {
				logrus.WithField("error", serr.ToJSON(err, true)).Info("GetHandleContextFunc fail")
				http.Error(w, httpErr.Message, httpErr.StatusCode)
				return
			}

			logrus.WithField("error", serr.ToJSON(err, true)).Error()
			http.Error(w, "error", http.StatusBadRequest)
			return
		}
		defer ctxCloseFunc()

		rspData, err := o.handle(ctx, funcInfo, []byte(data))
		if err != nil {
			logrus.WithContext(ctx).WithField("error", serr.ToJSON(err, true)).Error()
			http.Error(w, "error", http.StatusBadRequest)
			return
		}

		w.Header().Set("Accept", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		_, err = w.Write(rspData)
		if err != nil {
			logrus.WithContext(ctx).WithField("error", serr.ToJSON(err, true)).Error()
			http.Error(w, "error", http.StatusBadRequest)
			return
		}
	})
}

// f type must be HandleFuncType
func (o *WebServer) RegistFuncs(baseUrl string, httpMethod HttpMethod, middlewares []MiddlewareFuncType, f ...interface{}) {
	for _, v := range f {
		o.RegistFunc(baseUrl, httpMethod, middlewares, v)
	}
}
