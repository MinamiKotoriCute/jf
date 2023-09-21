package protobufhandler

import (
	"github.com/MinamiKotoriCute/jf/pkg/delivery"
	"github.com/rotisserie/eris"
)

func (o *ProtobufHandler) Regist(f interface{}) error {
	funcInfo, err := delivery.GetHandleFuncInfo(f)
	if err != nil {
		return eris.Wrap(err, "")
	}

	o.handleFuncs[funcInfo.ReqName] = funcInfo
	return nil
}

func (o *ProtobufHandler) Regists(fs ...interface{}) error {
	for _, f := range fs {
		if err := o.Regist(f); err != nil {
			return err
		}
	}

	return nil
}
