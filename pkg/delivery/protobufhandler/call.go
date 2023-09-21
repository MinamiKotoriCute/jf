package protobufhandler

import (
	"context"

	"github.com/MinamiKotoriCute/jf/pkg/delivery"
	"github.com/rotisserie/eris"
	"google.golang.org/protobuf/proto"
)

func Call[RspT proto.Message, ReqT proto.Message](o *ProtobufHandler, ctx context.Context, req ReqT) (RspT, error) {
	var a RspT
	rspData, err := o.Call(ctx, req)
	if err != nil {
		return a, err
	}

	rsp, ok := rspData.(RspT)
	if !ok {
		return a, eris.New("response type not match")
	}

	return rsp, nil
}

func (o *ProtobufHandler) Call(ctx context.Context, req proto.Message) (proto.Message, error) {
	funcInfo, ok := o.handleFuncs[string(req.ProtoReflect().Descriptor().FullName())]
	if !ok {
		return nil, eris.New("handle function not found")
	}

	return funcInfo.Call(ctx, req)
}

func (o *ProtobufHandler) GetHandleFuncInfo(reqPbName string) *delivery.HandleFuncInfo {
	funcInfo, ok := o.handleFuncs[reqPbName]
	if !ok {
		return nil
	}

	return funcInfo
}
