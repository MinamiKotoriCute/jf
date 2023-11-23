package webserver

import (
	"context"

	"github.com/MinamiKotoriCute/jf/pkg/delivery"
	"github.com/MinamiKotoriCute/serr"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/encoding/protojson"
)

func (o *WebServer) handle(ctx context.Context, funcInfo *delivery.HandleFuncInfo, data []byte) ([]byte, error) {
	req := funcInfo.NewReq()
	if len(data) != 0 {
		err := protojson.UnmarshalOptions{
			DiscardUnknown: true,
		}.Unmarshal(data, req)
		if err != nil {
			return nil, serr.Wrap(err)
		}
	}

	rsp, callErr := funcInfo.Call(ctx, req)
	isExpectedError := true
	if callErr != nil {
		if o.CreateInternalErrorRspFunc == nil {
			return nil, callErr
		}

		if rsp2, isExpectedError2, err2 := o.CreateInternalErrorRspFunc(callErr, string(req.ProtoReflect().Descriptor().FullName())); err2 != nil {
			logrus.WithContext(ctx).WithField("error", serr.ToJSON(err2, true)).Warning()
			return nil, serr.Wrap(callErr)
		} else {
			rsp = rsp2
			isExpectedError = isExpectedError2
		}
	}

	m := &protojson.MarshalOptions{
		UseProtoNames: true,
	}
	rspData, err := m.Marshal(rsp)
	if err != nil {
		return nil, serr.Wrap(err)
	}

	if o.OnHandleFinishedFunc != nil {
		o.OnHandleFinishedFunc(ctx, req, rsp, callErr, isExpectedError)
	}

	return rspData, nil
}
