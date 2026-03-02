package model

import (
	"time"

	"github.com/helays/ssh-proxy-plus/internal/types"
	"helay.net/go/utils/v3/dataType"
)

type ProxyInfo struct {
	Id        int                  `json:"id" gorm:"primaryKey;int;autoIncrement"`
	Address   string               `json:"address" gorm:"size:128;not null;uniqueIndex;comment:代理地址"`
	Type      types.ProxyType      `json:"type" gorm:"not null;comment:代理类型"`
	Latency   time.Duration        `json:"latency" gorm:"comment:延迟"`
	Speed     int64                `json:"speed" gorm:"comment:下载速度 (bytes/ms)"`
	Score     float64              `json:"score" gorm:"precision:10;scale:4;comment:综合评分"`
	LastCheck *dataType.CustomTime `json:"last_check" gorm:"comment:最后检测时间"`
	IsAlive   dataType.Bool        `json:"is_alive" gorm:"not null;index;default:0;comment:是否可用"`
	Message   string               `json:"message" gorm:"type:text;comment:错误信息"`

	CreateTime *dataType.CustomTime `json:"create_time,omitempty" gorm:"autoCreateTime:true;index;not null;default:current_timestamp;comment:记录创建时间"`
	UpdateTime *dataType.CustomTime `json:"update_time,omitempty" gorm:"autoUpdateTime:true;comment:记录更新时间"`
}

func (ProxyInfo) TableName() string {
	return "proxy_infos"
}
