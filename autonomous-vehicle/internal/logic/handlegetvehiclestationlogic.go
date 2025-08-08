package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"autonomous-vehicle/internal/svc"
	"autonomous-vehicle/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type HandleGetVehicleStationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHandleGetVehicleStationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleGetVehicleStationLogic {
	return &HandleGetVehicleStationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HandleGetVehicleStationLogic) HandleGetVehicleStation(req *types.GetVehicleStationReq) (*types.GetVehicleStationResq, error) {
	timestamp, nonce, signature, token, err := l.svcCtx.GenSignParams()
	if err != nil {
		return nil, err
	}

	// 测试环境:https://scapi.test.neolix.net/ 正式环境:https://scapi.neolix.net/
	url := fmt.Sprintf("https://scapi.neolix.net/openapi-server/slvapi/GetVehicleStation/v1?signature=%s&timeStamp=%s&nonce=%s&access_token=%s&vin=%s",
		signature, timestamp, nonce, token, req.Vin)

	httpReq, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("X-From", l.svcCtx.Config.XFrom)
	httpReq.Header.Set("X-Version", l.svcCtx.Config.XVersion)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result types.GetVehicleStationResq
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
