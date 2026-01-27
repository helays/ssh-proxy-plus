package auto_migrate

import (
	"errors"
	"github.com/helays/ssh-proxy-plus/configs"
	"github.com/helays/ssh-proxy-plus/internal/dal"
	"github.com/helays/ssh-proxy-plus/internal/model"
	"github.com/helays/utils/v2/logger/ulogs"
	"gorm.io/gorm"
)

func InitSysConfigData() {
	cfg := configs.Get()
	if cfg.Common.EnableAliEcs {
		setDefaultSysConfig(model.SysConfig{
			Prop:  "access_key_id",
			Label: "阿里云RAM ID",
			Name:  "el-input",
			Type:  "text",
			Component: map[string]any{
				"props": map[string]any{
					"placeholder":  "请输入阿里云RAM ID",
					"autocomplete": "on",
				},
			},
		})
		setDefaultSysConfig(model.SysConfig{
			Prop:  "access_key_secret",
			Label: "阿里云RAM Secret",
			Name:  "el-input",
			Type:  "password",
			Component: map[string]any{
				"show-password": true,
				"props": map[string]any{
					"placeholder":  "请输入阿里云RAM Secret",
					"autocomplete": "on",
				},
			},
		})
	}
	setDefaultSysConfig(model.SysConfig{
		Prop:  "sys_pass",
		Label: "系统通行证",
		Name:  "el-input",
		Type:  "password",
		Component: map[string]any{
			"show-password": true,
			"props": map[string]any{
				"placeholder":  "请输入系统通行证",
				"autocomplete": "on",
			},
		},
	})
}

func setDefaultSysConfig(value model.SysConfig) {
	db := dal.GetDB()
	err := db.Model(model.SysConfig{}).Where("prop = ?", value.Prop).Take(&model.SysConfig{}).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return
	}
	ulogs.Checkerr(db.Create(&value).Error, "创建初始化数据失败", value.Prop)
}
