package logic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"autonomous-vehicle/internal/svc"
	"autonomous-vehicle/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type HandleGetCurrentMissionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHandleGetCurrentMissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleGetCurrentMissionLogic {
	return &HandleGetCurrentMissionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HandleGetCurrentMissionLogic) HandleGetCurrentMission(req *types.GetCurrentMissionReq) (*types.GetCurrentMissionResp, error) {
	timestamp, nonce, signature, token, err := l.svcCtx.GenSignParams()
	if err != nil {
		return nil, err
	}

	// 测试环境:https://scapi.test.neolix.net/ 正式环境:https://scapi.neolix.net/
	url := fmt.Sprintf("https://scapi.neolix.net/openapi-server/slvapi/getCurrentMission?signature=%s&timeStamp=%s&nonce=%s&access_token=%s",
		signature, timestamp, nonce, token)

	bodyBytes, _ := json.Marshal(map[string]string{"vin": req.Vin})
	httpReq, err := http.NewRequest("POST", url, io.NopCloser(bytes.NewReader(bodyBytes)))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-From", l.svcCtx.Config.XFrom)
	httpReq.Header.Set("X-Version", l.svcCtx.Config.XVersion)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	var result types.GetCurrentMissionResp
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
