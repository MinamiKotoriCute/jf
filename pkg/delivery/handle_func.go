package delivery

import (
	"bytes"
	"context"
	"reflect"
	"runtime/debug"

	"github.com/DataDog/gostackparse"
	"github.com/MinamiKotoriCute/serr"
	"google.golang.org/protobuf/proto"
)

type HandleContextKey string
type HandleFuncType[ReqT proto.Message, RspT proto.Message] func(ctx context.Context, req ReqT) (RspT, error)
type OnHandleFinishedFuncType func(ctx context.Context, req, rsp proto.Message, callErr error, isExpectedError bool)

type HandleFuncInfo struct {
	ReqName string
	NewReq  func() proto.Message
	Call    func(context.Context, proto.Message) (proto.Message, error)
}

func (o *HandleFuncInfo) CallHandlePanic(ctx context.Context, req proto.Message) (rsp proto.Message, err error) {
	defer func() {
		if v := recover(); v != nil {
			stack := debug.Stack()
			goroutines, _ := gostackparse.Parse(bytes.NewReader(stack))

			rsp = nil
			err = serr.Errors(map[string]interface{}{
				"goroutines": goroutines,
				"value":      v,
			}, "panic")
		}
	}()

	rsp, err = o.Call(ctx, req)
	return
}

func IsHandleFuncType(f interface{}) error {
	fType := reflect.TypeOf(f)
	if fType.Kind() != reflect.Func {
		return serr.New("f is not func")
	}

	if fType.NumIn() != 2 {
		return serr.New("f in num error")
	}

	if fType.In(0) != reflect.TypeOf((*context.Context)(nil)).Elem() {
		return serr.New("f in[0] type error")
	}

	messageType := reflect.TypeOf((*proto.Message)(nil)).Elem()
	if !fType.In(1).Implements(messageType) {
		return serr.New("f in[1] type error")
	}

	if fType.NumOut() != 2 {
		return serr.New("f out num error")
	}

	if !fType.Out(0).Implements(messageType) {
		return serr.New("f out[0] type error")
	}

	if fType.Out(1) != reflect.TypeOf((*error)(nil)).Elem() {
		return serr.New("f out[1] type error")
	}

	return nil
}

func GetHandleFuncMethods(s interface{}) []interface{} {
	sv := reflect.ValueOf(s)

	funcs := []interface{}{}
	for i := 0; i < sv.NumMethod(); i++ {
		method := sv.Method(i)
		f := method.Interface()
		if err := IsHandleFuncType(f); err == nil {
			funcs = append(funcs, f)
		}
	}

	return funcs
}

// f type must be HandleFuncType
func GetHandleFuncInfo(f interface{}) (*HandleFuncInfo, error) {
	if err := IsHandleFuncType(f); err != nil {
		return nil, err
	}

	fType := reflect.TypeOf(f)
	newReq := func() proto.Message {
		return reflect.New(fType.In(1).Elem()).Interface().(proto.Message)
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

	return &HandleFuncInfo{
		ReqName: string(reqName),
		NewReq:  newReq,
		Call:    call,
	}, nil
}

func GetHandleFuncInfoByFunc[ReqT proto.Message, RspT proto.Message](handle func(context.Context, ReqT) (RspT, error)) *HandleFuncInfo {
	f := func(ctx context.Context, msg proto.Message) (proto.Message, error) {
		req, ok := msg.(ReqT)
		if !ok {
			return nil, serr.New("msg type error")
		}

		rsp, err := handle(ctx, req)
		if err != nil {
			return nil, err
		}

		return rsp, nil
	}

	var reqPointer ReqT
	reqPbName := reqPointer.ProtoReflect().Descriptor().FullName()

	newReq := func() proto.Message {
		return reqPointer.ProtoReflect().New().Interface()
	}

	return &HandleFuncInfo{
		ReqName: string(reqPbName),
		Call:    f,
		NewReq:  newReq,
	}
}
