package webserver

import (
	"context"

	"github.com/MinamiKotoriCute/jf/pkg/delivery"
	"github.com/rotisserie/eris"
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
			return nil, eris.Wrap(err, "")
		}
	}

	rsp, err := funcInfo.Call(ctx, req)
	if err != nil {
		if o.CreateInternalErrorRspFunc == nil {
			return nil, err
		}

		if rsp2, isExpectedError, err2 := o.CreateInternalErrorRspFunc(err, string(req.ProtoReflect().Descriptor().FullName())); err2 != nil {
			logrus.WithContext(ctx).WithField("error", eris.ToJSON(err2, true)).Warning()
			return nil, eris.Wrap(err, "")
		} else {
			if !isExpectedError {
				logrus.WithContext(ctx).WithField("error", eris.ToJSON(err, true)).Warning()
			}
			rsp = rsp2
		}
	}

	m := &protojson.MarshalOptions{
		UseProtoNames: true,
	}
	rspData, err := m.Marshal(rsp)
	if err != nil {
		return nil, eris.Wrap(err, "")
	}

	if o.OnHandleFinishedFunc != nil {
		o.OnHandleFinishedFunc(ctx, req, rsp)
	}

	return rspData, nil
}
