package controller

import (
	"github.com/helays/ssh-proxy-plus/configs"
	"github.com/helays/ssh-proxy-plus/internal/api/dto"
	"net/http"

	"github.com/helays/utils/v2/net/http/httpServer/router"
	"github.com/helays/utils/v2/net/http/response"
	"github.com/helays/utils/v2/net/http/session"
)

var frontendLists = []dto.FrontedResp{
	{
		Path:      "",
		Name:      "首页",
		Mod:       "default",
		Component: "HomeView",
		Meta: map[string]any{
			"title":        "首页",
			"require_auth": true,
		},
	},
	{
		Path:      "/page/server",
		Name:      "服务器开通",
		Mod:       "ali",
		Component: "ServerView",
		Meta: map[string]any{
			"title":        "服务器开通",
			"require_auth": true,
		},
	},
	{
		Path:      "/page/sysconfig",
		Name:      "系统配置",
		Mod:       "default",
		Component: "SysconfigView",
		Meta: map[string]any{
			"title":        "系统配置",
			"require_auth": true,
		},
	},
}

func (c *Controller) CtlMenuLists(w http.ResponseWriter, r *http.Request) {
	var resp []dto.FrontedResp
	cfg := configs.Get()
	for _, v := range frontendLists {
		if !cfg.Common.EnableAliEcs && v.Mod == "ali" {
			continue
		}
		if !cfg.Common.EnablePass && v.Mod == "login" {
			continue
		}
		v.Meta["mod"] = v.Mod
		v.Mod = ""
		resp = append(resp, v)
	}
	var userInfo router.LoginInfo
	if cfg.Common.EnablePass {
		_ = session.GetSession().Get(w, r, cfg.Router.SessionLoginName, &userInfo)
	}

	enablePass := "on"
	if !cfg.Common.EnablePass {
		enablePass = "off"
	} else {
		if !userInfo.IsLogin {
			resp = []dto.FrontedResp{}
		}
	}
	respData := map[string]any{
		"menu":        resp,
		"enable_pass": enablePass,
	}

	if cfg.Common.EnablePass {
		respData["islogin"] = userInfo.IsLogin
	}
	response.SetReturnCode(w, r, 0, "成功", respData)
}
