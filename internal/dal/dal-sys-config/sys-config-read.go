package dal_sys_config

import (
	"github.com/helays/ssh-proxy-plus/internal/cache"
	"github.com/helays/ssh-proxy-plus/internal/dal"
	"github.com/helays/ssh-proxy-plus/internal/model"

	"helay.net/go/utils/v3/logger/ulogs"
)

// ReadSysConfig2Cache 读取系统配置 到内存
func ReadSysConfig2Cache() {
	db := dal.GetDB()
	var lst []model.SysConfig
	if err := db.Find(&lst).Error; err != nil {
		ulogs.Errorf("读取系统配置失败 %v", err)
		return
	}
	for _, v := range lst {
		cache.SysConfig.Store(v.Prop, &v)
	}
}
