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

	// 测试环境:https://scapi.test.neolix.net/ 正式环境:https://scapi.neolix.net/
	url := fmt.Sprintf("https://scapi.neolix.net/openapi-server/slvapi/getVehicleList?signature=%s&timeStamp=%s&nonce=%s&access_token=%s",
		signature, timestamp, nonce, token)
	if req.UserId != "" {
		url += fmt.Sprintf("&userId=%s", req.UserId)
	}

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("X-From", l.svcCtx.Config.XFrom)
	httpReq.Header.Set("X-Version", l.svcCtx.Config.XVersion)

	// 输出组装好的 http 报文到日志
	l.Infof("HTTP Request: %s %s", httpReq.Method, httpReq.URL.String())
	for k, v := range httpReq.Header {
		l.Infof("Header: %s: %v", k, v)
	}

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
