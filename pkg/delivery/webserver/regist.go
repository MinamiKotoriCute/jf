package webserver

import (
	"io"
	"net/http"

	"github.com/MinamiKotoriCute/jf/pkg/delivery"
	"github.com/rotisserie/eris"
	"github.com/sirupsen/logrus"
)

// f type must be HandleFuncType
func (o *WebServer) RegistFunc(baseUrl string, httpMethod HttpMethod, f interface{}) {
	funcInfo, err := delivery.GetHandleFuncInfo(f)
	if err != nil {
		logrus.WithField("error", eris.ToJSON(err, true)).Fatal()
	}

	pattern := baseUrl + "/" + funcInfo.ReqName
	o.ServeMux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		if !httpMethod.Match(r.Method) {
			return
		}

		data := ""
		if r.Method == "GET" {
			data = r.URL.Query().Get("data")
		} else if r.Method == "POST" {
			defer r.Body.Close()
			tempData, err := io.ReadAll(r.Body)
			if err != nil {
				logrus.WithField("error", eris.ToJSON(err, true)).Error()
				return
			}

			data = string(tempData)
		} else if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", httpMethod.ToString())
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			return
		} else {
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
func (o *WebServer) RegistFuncs(baseUrl string, httpMethod HttpMethod, f ...interface{}) {
	for _, v := range f {
		o.RegistFunc(baseUrl, httpMethod, v)
	}
}
