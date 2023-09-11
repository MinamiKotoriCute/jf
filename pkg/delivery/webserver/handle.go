package webserver

import (
	"context"

	"github.com/MinamiKotoriCute/jf/pkg/delivery"
	"github.com/rotisserie/eris"
	"google.golang.org/protobuf/encoding/protojson"
)

func handle(ctx context.Context, funcInfo *delivery.HandleFuncInfo, data []byte, onFinishedFunc delivery.OnHandleFinishedFuncType) ([]byte, error) {
	req := funcInfo.NewReq()
	if len(data) != 0 {
		if err := protojson.Unmarshal(data, req); err != nil {
			return nil, eris.Wrap(err, "")
		}
	}

	rsp, err := funcInfo.Call(ctx, req)
	if err != nil {
		return nil, err
	}

	m := &protojson.MarshalOptions{
		UseProtoNames: true,
	}
	rspData, err := m.Marshal(rsp)
	if err != nil {
		return nil, eris.Wrap(err, "")
	}

	if onFinishedFunc != nil {
		onFinishedFunc(ctx, req, rsp)
	}

	return rspData, nil
}
