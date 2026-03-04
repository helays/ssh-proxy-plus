package controller

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/helays/ssh-proxy-plus/internal/api/dto"
	"github.com/helays/ssh-proxy-plus/internal/component/publicproxy"
	"github.com/helays/ssh-proxy-plus/internal/dal"
	dal_proxy "github.com/helays/ssh-proxy-plus/internal/dal/dal-proxy"
	"github.com/helays/ssh-proxy-plus/internal/model"
	"github.com/helays/ssh-proxy-plus/internal/types"
	"helay.net/go/utils/v3/db/userDb"
	"helay.net/go/utils/v3/net/http/request"
	"helay.net/go/utils/v3/net/http/response"
	"helay.net/go/utils/v3/net/http/response/dbresponse"
	"helay.net/go/utils/v3/tools/decode/json_decode_tee"
)

func (c *Controller) CtlProxyList(w http.ResponseWriter, r *http.Request) {
	dbresponse.RespListsWithFilter[model.ProxyInfo, *dto.ProxyResp](w, r, dal.GetDB(), userDb.Curd{}, dbresponse.Pager{
		Order: "score desc",
	})
}

func (c *Controller) CtlProxyCreate(w http.ResponseWriter, r *http.Request) {
	var postData dto.ProxyCreateReq
	if err := json_decode_tee.JsonDecode(r.Body, &postData); err != nil {
		response.SetReturnErrorDisableLog(w, err, http.StatusInternalServerError)
		return
	}
	u, err := url.Parse(postData.Address)
	if err != nil {
		response.SetReturnErrorDisableLog(w, err, http.StatusInternalServerError)
		return
	}
	proxyType := types.ProxyType(u.Scheme)
	if err = proxyType.Valid(); err != nil {
		response.SetReturnErrorDisableLog(w, err, http.StatusInternalServerError)
		return
	}
	saveData := model.ProxyInfo{
		Address: fmt.Sprintf("%s://%s", proxyType, u.Host),
		Type:    proxyType,
	}
	if err = dal_proxy.SaveProxy(&saveData); err != nil {
		response.SetReturnErrorDisableLog(w, err, http.StatusInternalServerError)
		return
	}
	response.SetReturnData(w, 0, "成功")
}

func (c *Controller) CtlProxyUpdateBest(w http.ResponseWriter, r *http.Request) {
	publicproxy.UpdateBestProxy()
	response.SetReturnData(w, 0, publicproxy.GetPublicProxy())
}

func (c *Controller) CtlProxyTest(w http.ResponseWriter, r *http.Request) {
	id, ok := request.GetQueryValueFromRequest2Int(r, "id")
	if !ok {
		response.SetReturnErrorDisableLog(w, errors.New("id不能为空"), http.StatusForbidden)
		return
	}
	var p model.ProxyInfo
	if err := dal_proxy.GetProxy(id, &p); err != nil {
		response.SetReturnErrorDisableLog(w, err, http.StatusInternalServerError)
		return
	}
	if err := publicproxy.DoCheckProxy(r.Context(), p); err != nil {
		response.SetReturnErrorDisableLog(w, err, http.StatusInternalServerError)
		return
	}
	response.SetReturnData(w, 0, "测试成功")
}

func (c *Controller) CtlProxyDelete(w http.ResponseWriter, r *http.Request) {
	id, ok := request.GetQueryValueFromRequest2Int(r, "id")
	if !ok {
		response.SetReturnErrorDisableLog(w, errors.New("id不能为空"), http.StatusForbidden)
		return
	}
	if err := dal_proxy.DeleteProxy(id); err != nil {
		response.SetReturnErrorDisableLog(w, err, http.StatusInternalServerError)
		return
	}
	response.SetReturnData(w, 0, "删除成功")
}
