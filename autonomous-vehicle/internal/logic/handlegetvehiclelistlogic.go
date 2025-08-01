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

type HandleGetVehicleListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHandleGetVehicleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleGetVehicleListLogic {
	return &HandleGetVehicleListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HandleGetVehicleListLogic) HandleGetVehicleList(req *types.GetVehicleListReq) (*types.GetVehicleListResp, error) {
	timestamp, nonce, signature, token, err := l.svcCtx.GenSignParams()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://scapi.test.neolix.net/openapi-server/slvapi/getVehicleList?signature=%s&timeStamp=%s&nonce=%s&access_token=%s&userId=%s",
		signature, timestamp, nonce, token, req.UserId)

	httpReq, err := http.NewRequest("GET", url, nil)
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

	var result types.GetVehicleListResp
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
