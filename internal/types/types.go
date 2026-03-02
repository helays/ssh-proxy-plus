package types

import (
	"database/sql/driver"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"helay.net/go/utils/v3/dataType"
)

// noinspection all
type ProxyType string

// noinspection all
func (p *ProxyType) Scan(val any) (err error) {
	return dataType.HelperStringScan(val, p)
}

// noinspection all
func (p ProxyType) Value() (driver.Value, error) {
	return string(p), nil
}

// GormDBDataType gorm db data type
// noinspection all
func (ProxyType) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return dataType.HelperStringGormDBDataType(db, field, 24)
}

// noinspection all
func (m ProxyType) String() string {
	return string(m)
}

// noinspection all
func (m ProxyType) Valid() error {
	switch m {
	case ProxySocks5, ProxySocks4, ProxyHttp, ProxyHttps:
		return nil
	}
	return fmt.Errorf("无效的代理类型 %s", m)
}

const (
	ProxySocks5 ProxyType = "socks5"
	ProxySocks4 ProxyType = "socks4"
	ProxyHttp   ProxyType = "http"
	ProxyHttps  ProxyType = "https"
)
