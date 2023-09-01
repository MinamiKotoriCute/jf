package delivery

import (
	"context"
	"reflect"

	"github.com/rotisserie/eris"
	"google.golang.org/protobuf/proto"
)

type HandleFuncType[ReqT proto.Message, RspT proto.Message] func(ctx context.Context, req ReqT) (RspT, error)

type HandleFuncInfo struct {
	ReqName string
	NewReq  func() proto.Message
	Call    func(context.Context, proto.Message) (proto.Message, error)
}

func IsHandleFuncType(f interface{}) error {
	fType := reflect.TypeOf(f)
	if fType.Kind() != reflect.Func {
		return eris.New("f is not func")
	}

	if fType.NumIn() != 2 {
		return eris.New("f in num error")
	}

	if fType.In(0) != reflect.TypeOf((*context.Context)(nil)).Elem() {
		return eris.New("f in[0] type error")
	}

	messageType := reflect.TypeOf((*proto.Message)(nil)).Elem()
	if !fType.In(1).Implements(messageType) {
		return eris.New("f in[1] type error")
	}

	if fType.NumOut() != 2 {
		return eris.New("f out num error")
	}

	if !fType.Out(0).Implements(messageType) {
		return eris.New("f out[0] type error")
	}

	if fType.Out(1) != reflect.TypeOf((*error)(nil)).Elem() {
		return eris.New("f out[1] type error")
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
