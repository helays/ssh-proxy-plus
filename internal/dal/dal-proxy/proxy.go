package dal_proxy

import (
	"fmt"

	"github.com/helays/ssh-proxy-plus/internal/dal"
	"github.com/helays/ssh-proxy-plus/internal/model"
	"gorm.io/gorm/clause"
)

// SaveProxy 保存代理
func SaveProxy(proxy *model.ProxyInfo) error {
	db := dal.GetDB()
	var totals int64
	tx := db.Model(proxy).Where(clause.Eq{Column: "address", Value: proxy.Address}).Count(&totals)
	if err := tx.Error; err != nil {
		return fmt.Errorf("查询公共代理 %s 是否存在失败 %v", proxy.Address, err)
	}
	if totals > 0 {
		return nil
	}
	return tx.Create(proxy).Error
}

func UpdateProxy(proxy *model.ProxyInfo) error {
	db := dal.GetDB()
	return db.Model(proxy).Where(clause.Eq{Column: "address", Value: proxy.Address}).Updates(proxy).Error
}

func FindAllProxes() ([]*model.ProxyInfo, error) {
	db := dal.GetDB()
	var lst []*model.ProxyInfo
	if err := db.Find(&lst).Error; err != nil {
		return nil, fmt.Errorf("查询所有代理失败 %v", err)
	}
	return lst, nil
}
