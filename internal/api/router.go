package api

import (
	"github.com/helays/ssh-proxy-plus/configs"
	"github.com/helays/ssh-proxy-plus/dist"
	"github.com/helays/ssh-proxy-plus/internal/api/controller"
	"regexp"
)

func InitRouter() {
	cfg := configs.Get()
	// noinspection GoBoolExpressions
	if dist.EnableStaticFs {
		cfg.Router.SetStaticEmbedFs("/", &dist.StaticFS, "html")
	}
	htmlGroup := cfg.HttpServer.Group("")
	htmlGroup.Get("/", cfg.Router.Index)

	rootGroup := cfg.HttpServer.Group("/api/v1/")

	ctl := controller.New(cfg.Router)

	rootGroup.Get("/run.menu.lists", ctl.CtlMenuLists)

	rootGroup.Get("/run.sysconfig", ctl.CtlSysConfig)
	rootGroup.Post("/run.sysconfig", ctl.CtlSysConfig)

	rootGroup.Get("/edit.api", ctl.CtlForward)
	rootGroup.Post("/edit.api", ctl.CtlForward)
	rootGroup.Put("/edit.api", ctl.CtlForward)
	rootGroup.Delete("/edit.api", ctl.CtlForward)
	rootGroup.Patch("/edit.api", ctl.CtlForward)

	rootGroup.Ws("/data.api", ctl.CtlWSDataApi)

	// 阿里 ecs 接口
	if cfg.Common.EnableAliEcs {
		rootGroup.Get("/describe.regions", ctl.CtlDescribeRegions)
		rootGroup.Get("/describe.available.resource", ctl.CtlDescribeAvailableResource)
		rootGroup.Get("/describe.v.switches", ctl.CtlDescribeVSwitches)
		rootGroup.Get("/describe.security.groups", ctl.CtlDescribeSecurityGroups)
		rootGroup.Get("/ali.describe.instances", ctl.CtlDescribeInstances)
		rootGroup.Post("/ali.run.instances", ctl.CtlCreateRunInstances)
		rootGroup.Post("/ali.del.instances", ctl.CtlDelInstances)
	}

	if cfg.Common.EnablePass {
		rootGroup.Get("/run.captcha", cfg.Router.Router.Captcha)
		rootGroup.Get("/run.login", ctl.CtlLogin)
		rootGroup.Post("/run.login", ctl.CtlLogin)
		rootGroup.Get("/run.logout", ctl.CtlLogout)

		cfg.Router.Router.MustLoginPathRegexp = []*regexp.Regexp{
			regexp.MustCompile("^.*?/run\\.sysconfig$"),
			regexp.MustCompile("^.*?/edit\\.api$"),
			regexp.MustCompile("^.*?/data\\.api$"),
			regexp.MustCompile("^.*?/ali$"),
			regexp.MustCompile("^.*?/describe$"),
		}
	}

}
