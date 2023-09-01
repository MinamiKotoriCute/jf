package webserver

import (
	"context"
	"io"
	"net/http"

	"github.com/MinamiKotoriCute/jf/pkg/delivery"
	"github.com/golang/glog"
	"github.com/rotisserie/eris"
	"google.golang.org/protobuf/encoding/protojson"
)

func handle(funcInfo *delivery.HandleFuncInfo, data []byte) ([]byte, error) {
	req := funcInfo.NewReq()
	if len(data) != 0 {
		if err := protojson.Unmarshal(data, req); err != nil {
			return nil, eris.Wrap(err, "")
		}
	}

	ctx := context.Background()
	rsp, err := funcInfo.Call(ctx, req)
	if err != nil {
		return nil, eris.Wrap(err, "")
	}

	m := &protojson.MarshalOptions{
		UseProtoNames: true,
	}
	rspData, err := m.Marshal(rsp)
	if err != nil {
		return nil, eris.Wrap(err, "")
	}

	return rspData, nil
}

// f type must be HandleFuncType
func (o *WebServer) RegistGetFunc(baseUrl string, f interface{}) {
	funcInfo, err := delivery.GetHandleFuncInfo(f)
	if err != nil {
		glog.Fatal(eris.ToString(err, true))
	}

	pattern := baseUrl + "/" + funcInfo.ReqName
	o.serveMux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		data := r.URL.Query().Get("data")

		rspData, err := handle(funcInfo, []byte(data))
		if err != nil {
			glog.Errorf("handle fail. err:%v", eris.ToString(err, true))
			return
		}

		w.Header().Set("Accept", "application/json")
		_, err = w.Write(rspData)
		if err != nil {
			glog.Errorf("w.Write fail. err:%v", err)
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
		glog.Fatal(eris.ToString(err, true))
	}

	pattern := baseUrl + "/" + funcInfo.ReqName
	o.serveMux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		data, err := io.ReadAll(r.Body)
		if err != nil {
			glog.Errorf("io.ReadAl fail. err:%v", err)
			return
		}

		rspData, err := handle(funcInfo, []byte(data))
		if err != nil {
			glog.Errorf("handle fail. err:%v", eris.ToString(err, true))
			return
		}

		w.Header().Set("Accept", "application/json")
		_, err = w.Write(rspData)
		if err != nil {
			glog.Errorf("w.Write fail. err:%v", err)
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
