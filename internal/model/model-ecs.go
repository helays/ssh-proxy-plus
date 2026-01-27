package model

import (
	"database/sql/driver"
	"github.com/helays/utils/v2/dataType"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type Disk struct {
	Category string `json:"category"` // 系统盘类型
	Size     int32  `json:"size"`     // 系统盘大小
}

func (d *Disk) Scan(value any) error {
	return dataType.DriverScanWithJson(value, d)
}

func (d Disk) Value() (driver.Value, error) {
	return dataType.DriverValueWithJson(d)
}

// GormDataType gorm common data type
func (d Disk) GormDataType() string {
	return "json"
}

// GormDBDataType gorm db data type
func (Disk) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return dataType.JsonDbDataType(db, field)
}

type AliEcsOrder struct {
	Id                          int    `json:"id" gorm:"primaryKey;int;autoIncrement"`
	RegionId                    string `json:"region_id" gorm:"type:varchar(128);index;not null;comment:实例所属地"`              //实例所属的地域ID
	ImageId                     string `json:"image_id" gorm:"type:varchar(128);not null"`                                   // 镜像ID
	InstanceType                string `json:"instance_type" gorm:"type:varchar(128);not null;comment:资源规格"`                 // 实例的资源规格
	PasswordInherit             bool   `json:"password_inherit"`                                                             //是否使用镜像预设的密码
	AutoRenew                   bool   `json:"auto_renew"`                                                                   //是否自动续费
	AutoReleaseTime             int    `json:"auto_release_time" gorm:"type:int;not null;default:1"`                         // 释放时间
	InstanceChargeType          string `json:"instance_charge_type" gorm:"type:varchar(64);not null;default:PostPaid"`       // 实例计费类型
	AutoPay                     bool   `json:"auto_pay"`                                                                     // 是否自动付费
	InternetChargeType          string `json:"internet_charge_type" gorm:"type:varchar(64);not null;default:PayByBandwidth"` // 网络计费类型
	InternetMaxBandwidth        int32  `json:"internet_max_bandwidth" gorm:"type:int;default:1;not null;"`                   // 公网带宽最大值
	DryRun                      bool   `json:"dry_run"`                                                                      // 预检请求
	SecurityGroupId             string `json:"security_group_id" gorm:"type:varchar(128)"`                                   // 安全组ID
	Password                    string `json:"password" gorm:"type:varchar(32)"`                                             // 密码
	SystemDisk                  Disk   `json:"system_disk" gorm:"type:text;"`                                                // 系统盘
	IoOptimized                 string `json:"io_optimized" gorm:"type:varchar(32)"`                                         // 是否I/O优化
	SecurityEnhancementStrategy string `json:"security_enhancement_strategy" gorm:"type:varchar(32)"`                        // 安全增强
	VSwitchId                   string `json:"v_switch_id" gorm:"type:varchar(128)"`                                         // 交换机ID
	LocalListenAddr             string `json:"local_listen_addr" gorm:"type:varchar(128);not null;default:''"`               // 本地代理监听地址
	ConnectId                   string `json:"connect_id" gorm:"type:varchar(24);index;not null;default:''"`                 //

	OrderStatus int               `json:"order_status" gorm:"type:int;not null;default:-1;index"`        // 订单状态 -1，默认创建，200 创建成功 400 创建失败
	ErrMessage  string            `json:"err_message" gorm:"type:varchar(1024);not null;default:''"`     // 订单失败message
	ErrData     datatypes.JSONMap `json:"err_data" gorm:"type:text;"`                                    // 订单失败数据
	RequestId   string            `json:"request_id" gorm:"type:varchar(128);not null;default:''"`       // 订单请求ID
	OrderId     string            `json:"order_id" gorm:"type:varchar(128);not null;default:''"`         // 订单ID
	TradePrice  float64           `json:"trade_price" gorm:"type:decimal(10,4);not null;default:0.0000"` // 订单成交价
	InstanceId  string            `json:"instance_id" gorm:"type:varchar(128);not null;default:''"`      // 实例ID

	QueryStatus     int               `json:"query_status" gorm:"type:int;not null;default:-1;index"`
	QueryErrMessage string            `json:"query_err_message" gorm:"type:varchar(1024);not null;default:''"`
	QueryErrData    datatypes.JSONMap `json:"query_err_data" gorm:"type:text;"`
	RunStatus       string            `json:"run_status" gorm:"type:varchar(64);not null;default:''"`         // 运行状态
	PublicIpAddress string            `json:"public_ip_address" gorm:"type:varchar(128);not null;default:''"` // 实例IP

	CreateTime dataType.CustomTime `json:"create_time" gorm:"index;not null;type:timestamp;default:current_timestamp"`     // 创建时间
	UpdateTime dataType.CustomTime `json:"update_time" gorm:"index;not null;type:timestamp;default:'0000-00-00 00:00:00'"` // 更新时间
}

func (a *AliEcsOrder) TableName() string {
	return "ali_ecs_order"
}
