package main

import (
	"context"
	"encoding/gob"
	"fmt"

	"helay.net/go/utils/v3/close/vclose"
	"helay.net/go/utils/v3/config/parseCmd"
	"helay.net/go/utils/v3/logger/ulogs"
	"helay.net/go/utils/v3/net/http/httpServer/router"
	"helay.net/go/utils/v3/net/http/session"
	"helay.net/go/utils/v3/net/http/session/storage/carrier_file"
	"helay.net/go/utils/v3/net/http/session/storage/carrier_memory"
	"helay.net/go/utils/v3/net/http/session/storage/carrier_rdbms"
	"helay.net/go/utils/v3/safe/cachemgr"
	"helay.net/go/utils/v3/signalTools"
	"helay.net/go/utils/v3/tools"

	"github.com/helays/ssh-proxy-plus/internal/api"
	"github.com/helays/ssh-proxy-plus/internal/cache"
	cmp_proxy "github.com/helays/ssh-proxy-plus/internal/component/cmp-proxy"
	"github.com/helays/ssh-proxy-plus/internal/dal"
	auto_migrate "github.com/helays/ssh-proxy-plus/internal/dal/auto-migrate"
	dal_connect "github.com/helays/ssh-proxy-plus/internal/dal/dal-connect"
	dal_sys_config "github.com/helays/ssh-proxy-plus/internal/dal/dal-sys-config"
	"github.com/helays/ssh-proxy-plus/internal/model"

	"github.com/helays/ssh-proxy-plus/configs"
)

//
// ━━━━━━神兽出没━━━━━━
// 　　 ┏┓     ┏┓
// 　　┏┛┻━━━━━┛┻┓
// 　　┃　　　　　 ┃
// 　　┃　　━　　　┃
// 　　┃　┳┛　┗┳  ┃
// 　　┃　　　　　 ┃
// 　　┃　　┻　　　┃
// 　　┃　　　　　 ┃
// 　　┗━┓　　　┏━┛　Code is far away from bug with the animal protecting
// 　　　 ┃　　　┃    神兽保佑,代码无bug
// 　　　　┃　　　┃
// 　　　　┃　　　┗━━━┓
// 　　　　┃　　　　　　┣┓
// 　　　　┃　　　　　　┏┛
// 　　　　┗┓┓┏━┳┓┏┛
// 　　　　 ┃┫┫ ┃┫┫
// 　　　　 ┗┻┛ ┗┻┛
//
// ━━━━━━感觉萌萌哒━━━━━━
//
//
// User helay
// Date: 2024/6/29 13:33
//

func init() {
	parseCmd.Parseparams()
	configs.Init()
	dal.Init()
	auto_migrate.AutoMigrate()
	auto_migrate.InitSysConfigData()

}

func initSession(ctx context.Context) {
	cfg := configs.Get()
	if !cfg.Common.EnablePass {
		return
	}
	var (
		storage session.StorageDriver
		err     error
	)
	gob.Register(router.LoginInfo{})

	switch cfg.SessionConfig.SessionEngine {
	case cachemgr.DriverMemory:
		storage = carrier_memory.New(ctx)
	case cachemgr.DriverFile:
		storage, err = carrier_file.New(cfg.SessionConfig.SessionFilePath)
		if err != nil {
			panic(fmt.Errorf("初始化session 存储失败 %v", err))
		}
	case cachemgr.DriverRdbms:
		storage = carrier_rdbms.New(dal.GetDB())
	default:
		ulogs.Error("错误的session 引擎配置", cfg.SessionConfig.SessionEngine)
	}
	session.StartSession(ctx, storage, &cfg.Options)
	go tools.RunOnContextDone(ctx, func() { vclose.Close(storage) })
	ulogs.Info("session 模块初始化成功")
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	cfg := configs.Get()
	go signalTools.SignalHandle(func() {
		cancel()
	})
	initSession(ctx)
	cache.Init(ctx)
	cmp_proxy.Init(ctx)
	dal_sys_config.ReadSysConfig2Cache()
	dal_connect.ReadConnect2Cache()
	// 用这个方法来读取
	cache.ConnectList.ReadWith(func(connects []model.Connect) {
		cmp_proxy.StartList(connects)
	})

	api.InitRouter()
	cfg.HttpServer.HttpServerStart(ctx)

}
