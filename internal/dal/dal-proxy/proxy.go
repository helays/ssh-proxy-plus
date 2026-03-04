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
	tx := db.Where("id=?", proxy.Id).Select("latency", "speed", "score", "last_check", "is_alive", "message", "port_open")
	return tx.Updates(proxy).Error
}

func DeleteProxy(id int) error {
	db := dal.GetDB()
	tx := db.Where(clause.Eq{Column: "id", Value: id}).Delete(&model.ProxyInfo{})
	return tx.Error
}

func GetProxy(id int, info *model.ProxyInfo) error {
	db := dal.GetDB()
	tx := db.Where(clause.Eq{Column: "id", Value: id})
	if err := tx.Take(&info).Error; err != nil {
		return fmt.Errorf("查询代理 %d 失败 %v", id, err)
	}
	return nil
}

func FindAllProxes() ([]model.ProxyInfo, error) {
	db := dal.GetDB()
	var lst []model.ProxyInfo
	if err := db.Find(&lst).Error; err != nil {
		return nil, fmt.Errorf("查询所有代理失败 %v", err)
	}
	return lst, nil
}

func BestProxy() (addr string, err error) {
	db := dal.GetDB()
	var info model.ProxyInfo
	tx := db.Where(clause.And(
		clause.Eq{Column: "is_alive", Value: 1},
		clause.Eq{Column: "port_open", Value: 1},
	))
	tx.Order(clause.OrderByColumn{Column: clause.Column{Name: "score"}, Desc: true})
	if err = tx.Take(&info).Error; err != nil {
		return
	}
	return info.Address, nil
}
