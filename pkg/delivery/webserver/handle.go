package webserver

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"

	"github.com/golang/glog"
	"github.com/rotisserie/eris"
	"google.golang.org/protobuf/proto"
)

type HandleFuncType[ReqT proto.Message, RspT proto.Message] func(ctx context.Context, req ReqT) (RspT, error)

type handleFuncInfo struct {
	ReqName string
	NewReq  func() proto.Message
	Call    func(context.Context, proto.Message) (proto.Message, error)
}

// f type must be HandleFuncType
func parseHandleFunc(f interface{}) (*handleFuncInfo, error) {
	fType := reflect.TypeOf(f)
	// if fType.Kind() != reflect.Func {
	// 	return nil, eris.New("f is not func")
	// }

	// if fType.NumIn() != 2 {
	// 	return nil, eris.New("f in num error")
	// }

	// if fType.In(0) != reflect.TypeOf((*context.Context)(nil)).Elem() {
	// 	return nil, eris.New("f in[0] type error")
	// }

	// messageType := reflect.TypeOf((*proto.Message)(nil)).Elem()
	// if !fType.In(1).Implements(messageType) {
	// 	return nil, eris.New("f in[1] type error")
	// }

	// if fType.NumOut() != 2 {
	// 	return nil, eris.New("f out num error")
	// }

	// if fType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
	// 	return nil, eris.New("f out[0] type error")
	// }

	// if !fType.Out(1).Implements(messageType) {
	// 	return nil, eris.New("f out[1] type error")
	// }

	newReq := func() proto.Message {
		return reflect.New(fType.In(1)).Elem().Interface().(proto.Message)
	}
	reqName := newReq().ProtoReflect().Descriptor().FullName()
	call := func(ctx context.Context, req proto.Message) (proto.Message, error) {
		results := reflect.ValueOf(f).Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(req)})
		if !results[1].IsNil() {
			if err := results[1].Interface().(error); err != nil {
				return nil, err
			}
		}

		return results[0].Interface().(proto.Message), nil
	}

	return &handleFuncInfo{
		ReqName: string(reqName),
		NewReq:  newReq,
		Call:    call,
	}, nil
}

// f type must be HandleFuncType
func (o *WebServer) HandleGetFunc(baseUrl string, f interface{}) {
	funcInfo, err := parseHandleFunc(f)
	if err != nil {
		glog.Fatal(eris.ToString(err, true))
	}

	pattern := baseUrl + "/" + funcInfo.ReqName
	o.serveMux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		data := r.URL.Query().Get("data")
		req := funcInfo.NewReq()
		if len(data) != 0 {
			if err := json.Unmarshal([]byte(data), req); err != nil {
				glog.Errorf("proto.Unmarshall fail. err:%v", err)
				return
			}
		}

		ctx := context.Background()
		rsp, err := funcInfo.Call(ctx, req)
		if err != nil {
			glog.Errorf("call service fail. err:%v", eris.ToString(err, true))
			return
		}

		rspData, err := json.Marshal(rsp)
		if err != nil {
			glog.Errorf("proto.Marshal fail. err:%v", err)
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

// check f argument type at compile time
func HandleGet[ReqT proto.Message, RspT proto.Message](o *WebServer, baseUrl string, f HandleFuncType[ReqT, RspT]) {
	o.HandleGetFunc(baseUrl, f)
}

// f type must be HandleFuncType
func (o *WebServer) HandlePostFunc(baseUrl string, f interface{}) {
	funcInfo, err := parseHandleFunc(f)
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

		req := funcInfo.NewReq()
		if len(data) != 0 {
			if err := json.Unmarshal(data, req); err != nil {
				glog.Errorf("proto.Unmarshall fail. err:%v", err)
				return
			}
		}

		ctx := context.Background()
		rsp, err := funcInfo.Call(ctx, req)
		if err != nil {
			glog.Errorf("call service fail. err:%v", eris.ToString(err, true))
			return
		}

		rspData, err := json.Marshal(rsp)
		if err != nil {
			glog.Errorf("proto.Marshal fail. err:%v", err)
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

// check f argument type at compile time
func HandlePost[ReqT proto.Message, RspT proto.Message](o *WebServer, baseUrl string, f HandleFuncType[ReqT, RspT]) {
	o.HandlePostFunc(baseUrl, f)
}
