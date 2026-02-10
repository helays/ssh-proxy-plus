package controller

import (
	"github.com/helays/ssh-proxy-plus/internal/dal"
	dal_sys_config "github.com/helays/ssh-proxy-plus/internal/dal/dal-sys-config"
	"github.com/helays/ssh-proxy-plus/internal/model"
	"net/http"

	"helay.net/go/utils/v3/net/http/request"
	"helay.net/go/utils/v3/net/http/response"
)

func (c *Controller) CtlSysConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		c.sysConfigRead(w, r)
		return
	}
	postData, err := request.JsonDecode[[]model.SysConfig](r)
	if err != nil {
		response.SetReturnErrorDisableLog(w, err, http.StatusInternalServerError)
		return
	}
	db := dal.GetDB()
	if err = db.Where("1=1").Delete(model.SysConfig{}).Error; err != nil {
		response.SetReturnErrorDisableLog(w, err, http.StatusInternalServerError)
		return
	}
	if err = db.Create(&postData).Error; err != nil {
		response.SetReturnErrorDisableLog(w, err, http.StatusInternalServerError)
		return
	}
	dal_sys_config.ReadSysConfig2Cache()
	response.SetReturnData(w, 0, "保存成功")
}

func (c *Controller) sysConfigRead(w http.ResponseWriter, r *http.Request) {
	var data []model.SysConfig
	db := dal.GetDB()
	if err := db.Find(&data).Error; err != nil {
		response.SetReturnErrorDisableLog(w, err, http.StatusInternalServerError)
		return
	}
	response.SetReturnData(w, 0, "查询成功", data)
}
