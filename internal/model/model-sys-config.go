package model

import (
	"github.com/helays/utils/v2/dataType"
	"gorm.io/datatypes"
)

type SysConfig struct {
	Prop       string              `json:"prop" gorm:"primaryKey;type:varchar(128)"`
	Label      string              `json:"label" gorm:"type:varchar(128);not null"`
	Value      string              `json:"value" gorm:"type:text;not null;default:''"`
	Name       string              `json:"name" gorm:"type:varchar(128);not null;default:'el-input'"`
	Type       string              `json:"type" gorm:"type:varchar(128);not null;default:'text'"`
	Component  datatypes.JSONMap   `json:"component" gorm:"type:text"`
	CreateTime dataType.CustomTime `json:"create_time" gorm:"index;not null;type:timestamp;default:current_timestamp"`     // 创建时间
	UpdateTime dataType.CustomTime `json:"update_time" gorm:"index;not null;type:timestamp;default:'0000-00-00 00:00:00'"` // 更新时间
}

func (c *SysConfig) TableName() string {
	return "sys_config"
}
