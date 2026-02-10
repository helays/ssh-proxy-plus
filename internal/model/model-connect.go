package model

import (
	"github.com/helays/ssh-proxy-plus/internal/types"
	"helay.net/go/utils/v3/dataType"
)

type Connect struct {
	Id         string              `json:"id" gorm:"type:varchar(24);primaryKey"`
	Pid        string              `json:"pid" gorm:"type:varchar(24);index"`
	Lname      string              `json:"lname" gorm:"type:varchar(128);not null;default:'';"`                            // 连接名称
	Saddr      string              `json:"saddr" gorm:"type:varchar(128);not null;default:''"`                             // 目标地址
	User       string              `json:"user" gorm:"type:varchar(128);not null;index"`                                   // 用户
	Stype      types.SSHValidType  `json:"type" gorm:"type:int;not null;default:1"`                                        // 验证方式 1、密码验证 2、密钥验证
	Passwd     string              `json:"passwd" gorm:"type:text;not null;default:''"`                                    // 密码、密钥路径
	Remote     string              `json:"remote" gorm:"type:varchar(128);not null;default:''"`                            // 远程地址
	Listen     string              `json:"listen" gorm:"type:varchar(128);not null;default:''"`                            // 本地监听地址
	Connect    types.ForwardType   `json:"connect" gorm:"type:varchar(128);not null;default:''"`                           // 连接类型 L 本地转发 R 远程转发 D 动态转发 H HTTP代理
	Active     types.TextStatus    `json:"active" gorm:"type:char(1);index;not null;default:'Y'"`                          // 是否启用
	CreateTime dataType.CustomTime `json:"create_time" gorm:"index;not null;type:timestamp;default:current_timestamp"`     // 创建时间
	UpdateTime dataType.CustomTime `json:"update_time" gorm:"index;not null;type:timestamp;default:'0000-00-00 00:00:00'"` // 更新时间
	Son        []Connect           `json:"son" gorm:"-:all"`                                                               // 子连接
}

func (c *Connect) TableName() string {
	return "connect"
}

func (c *Connect) Db2Connect() Connect {
	return Connect{
		Son:        nil,
		Lname:      c.Lname,
		Saddr:      c.Saddr,
		User:       c.User,
		Stype:      c.Stype,
		Passwd:     c.Passwd,
		Remote:     c.Remote,
		Listen:     c.Listen,
		Connect:    c.Connect,
		Id:         c.Id,
		Pid:        c.Pid,
		Active:     c.Active,
		CreateTime: c.CreateTime,
		UpdateTime: c.UpdateTime,
	}
}
