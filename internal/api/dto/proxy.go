package dto

import (
	"github.com/helays/ssh-proxy-plus/internal/model"
)

type ProxyResp []model.ProxyInfo

func (p ProxyResp) RespFilter() {
}

type ProxyCreateReq struct {
	Address string `json:"address"`
}
