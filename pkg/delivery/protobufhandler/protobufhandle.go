package protobufhandler

import "github.com/MinamiKotoriCute/jf/pkg/delivery"

type ProtobufHandler struct {
	handleFuncs map[string]*delivery.HandleFuncInfo
}

func NewProtobufHandler() *ProtobufHandler {
	return &ProtobufHandler{
		handleFuncs: make(map[string]*delivery.HandleFuncInfo),
	}
}

func (o *ProtobufHandler) GetHandlers() map[string]*delivery.HandleFuncInfo {
	return o.handleFuncs
}
