package main

import (
	"flag"
	"fmt"
	"time"

	"drone-api/internal/config"
	"drone-api/internal/handler"
	"drone-api/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/drone-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	defer ctx.Dao.Close()

	handler.RegisterHandlers(server, ctx)

	fmt.Printf("%s, Starting server at %s:%d...\n", time.Now(), c.Host, c.Port)
	server.Start()
}
