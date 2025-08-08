package main

import (
	"autonomous-vehicle/internal/config"
	"autonomous-vehicle/internal/handler"
	"autonomous-vehicle/internal/logic"
	"autonomous-vehicle/internal/svc"
	"autonomous-vehicle/internal/types"
	"context"
	"flag"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/autonomousvehicle.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	defer ctx.Dao.Close()

	// 用于保存车辆列表的内存变量
	var vehicleList []types.VehicleList

	// 每24小时拉取一次车辆列表
	go func() {
		for {
			// fmt.Println("拉取车辆列表...")
			list, err := fetchVehicleList(ctx)
			if err != nil {
				fmt.Printf("拉取车辆列表失败: %+v\n", err)
			} else {
				vehicleList = list
				// fmt.Printf("车辆列表已更新，数量: %d\n", len(vehicleList))
			}
			time.Sleep(6 * time.Hour)
		}
	}()

	// 每1秒拉取一次所有车辆信息
	go func() {
		for {
			if len(vehicleList) == 0 {
				time.Sleep(1 * time.Second)
				continue
			}

			var wg sync.WaitGroup
			workerNum := runtime.NumCPU() // 可根据实际情况调整并发数
			sem := make(chan struct{}, workerNum)

			for _, v := range vehicleList {
				wg.Add(1)
				sem <- struct{}{}
				go func(v types.VehicleList) {
					defer wg.Done()
					defer func() { <-sem }()
					if err := fetchAndSaveVehicleInfo(ctx, v.Vin); err != nil {
						fmt.Printf("拉取车辆 %s 信息失败: %+v\n", v.Vin, err)
					}
				}(v)
			}
			wg.Wait()
			time.Sleep(1 * time.Second) // 每秒拉取一次所有车辆
		}
	}()

	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}

// 拉取车辆列表
func fetchVehicleList(ctx *svc.ServiceContext) ([]types.VehicleList, error) {
	logicList := logic.NewHandleGetVehicleListLogic(context.Background(), ctx)
	listResp, err := logicList.HandleGetVehicleList(&types.GetVehicleListReq{
		// UserId: "zsdxtest",
	})
	if err != nil {
		return nil, err
	}
	return listResp.Data, nil
}

// 拉取并保存单辆车信息（只保存在线车辆）
func fetchAndSaveVehicleInfo(ctx *svc.ServiceContext, vin string) error {
	logicInfo := logic.NewHandleGetVehicleInfoLogic(context.Background(), ctx)
	infoResp, err := logicInfo.HandleGetVehicleInfo(&types.GetVehicleInfoReq{Vin: vin})
	if err != nil {
		return err
	}
	if infoResp.Data.Vin != "" && infoResp.Data.PowerState {
		return ctx.Dao.SaveVehicleInfo(&infoResp.Data)
	}
	return nil
}
