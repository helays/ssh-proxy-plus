package controller

import (
	"encoding/json"
	"fmt"
	"github.com/helays/ssh-proxy-plus/configs"
	"github.com/helays/ssh-proxy-plus/internal/api/dto"
	"github.com/helays/ssh-proxy-plus/internal/cache"
	"net/http"
	"time"

	"github.com/dchest/captcha"
	"helay.net/go/utils/v3/crypto/rsaV2"
	"helay.net/go/utils/v3/net/http/httpServer/router"
	"helay.net/go/utils/v3/net/http/response"
	"helay.net/go/utils/v3/net/http/session"
)

func (c *Controller) CtlLogin(w http.ResponseWriter, r *http.Request) {
	cfg := configs.Get()
	if r.Method == http.MethodGet {
		c.generateRSA(w, r, cfg.Router.Router.SessionLoginName)
		return
	}
	loginInfo := router.LoginInfo{}
	sessionMgr := session.GetSession()
	err := sessionMgr.Flashes(w, r, cfg.Router.SessionLoginName, &loginInfo)
	if err != nil {
		response.SetReturnErrorDisableLog(w, err, http.StatusInternalServerError, "先获取加密密钥")
		return
	}

	var postData dto.LoginReq
	jd := json.NewDecoder(r.Body)
	if err = jd.Decode(&postData); err != nil {
		response.SetReturnErrorDisableLog(w, err, http.StatusInternalServerError, "参数解析失败")
		return
	}
	if postData.Captcha == "" {
		response.SetReturnErrorDisableLog(w, fmt.Errorf("请输入验证码"), http.StatusForbidden)
		return
	}
	var captchaId string
	if err = sessionMgr.Get(w, r, router.CaptchaID, &captchaId); err != nil {
		response.SetReturnErrorDisableLog(w, err, http.StatusForbidden, "验证码失效")
		return
	}
	if !captcha.VerifyString(captchaId, postData.Captcha) {
		response.SetReturnErrorDisableLog(w, fmt.Errorf("验证码错误"), http.StatusForbidden)
		return
	}
	_pass, err := rsaV2.DecryptWithPrivateKey(loginInfo.RsaPrivateKey, postData.Pass)
	if err != nil {
		response.SetReturnErrorDisableLog(w, err, http.StatusForbidden, "密码数据异常")
		return
	}
	if sysPass, ok := cache.SysConfig.Load("sys_pass"); !ok {
		response.SetReturnErrorDisableLog(w, fmt.Errorf("系统密码未配置"), http.StatusForbidden)
		return
	} else if sysPass.Value != string(_pass) {
		response.SetReturnErrorDisableLog(w, fmt.Errorf("登录失败，密码错误"), http.StatusForbidden)
		return
	}
	// 登录成功就存 session
	_ = sessionMgr.Set(w, r, &session.Value{
		Field: cfg.Router.SessionLoginName,
		Value: router.LoginInfo{LoginTime: time.Now(), IsLogin: true},
		TTL:   30 * 24 * time.Hour,
	})

	response.SetReturnData(w, 0, "成功", "登录成功")
}

func (c *Controller) CtlLogout(w http.ResponseWriter, r *http.Request) {
	_ = session.GetSession().Destroy(w, r)
	response.SetReturnData(w, 0, "成功", "退出成功")
}

func (c *Controller) generateRSA(w http.ResponseWriter, r *http.Request, sessionName string) {
	pri, pub, err := rsaV2.GenRsaPriPubKey(2048)
	if err != nil {
		response.SetReturnError(w, r, err, http.StatusForbidden)
		return
	}
	lgInfo := router.LoginInfo{RsaPrivateKey: pri, RsaPublickKey: pub}
	err = session.GetSession().Set(w, r, &session.Value{
		Field: sessionName,
		Value: lgInfo,
		TTL:   2 * time.Minute,
	})
	if err != nil {
		response.SetReturnWithoutError(w, r, err, http.StatusInternalServerError, "session设置失败")
		return
	}
	response.SetReturnData(w, 0, "成功", string(pub))
}
