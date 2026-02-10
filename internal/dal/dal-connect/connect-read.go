package dal_connect

import (
	"fmt"
	"github.com/helays/ssh-proxy-plus/internal/cache"
	"github.com/helays/ssh-proxy-plus/internal/dal"
	"github.com/helays/ssh-proxy-plus/internal/model"

	"helay.net/go/utils/v3/dataType"
	"helay.net/go/utils/v3/logger/ulogs"
)

func SaveData(data *model.Connect, isNew bool) error {
	db := dal.GetDB()
	if isNew {
		if err := db.Create(data).Error; err != nil {
			return fmt.Errorf("隧道连接信息保存失败 %v", err)
		}
		return nil
	}
	tx := db.Where("id like ?", data.Id)
	data.UpdateTime = dataType.NewCustomTimeNow()
	err := tx.Omit("id", "active", "create_time").Updates(data).Error
	if err != nil {
		return fmt.Errorf("隧道连接信息更新失败 %v", err)
	}
	return nil
}

// ReadConnect2Cache 读取数据库所有连接信息到缓存
func ReadConnect2Cache() {
	db := dal.GetDB()
	var lst []model.Connect
	if err := db.Find(&lst).Error; err != nil {
		ulogs.Errorf("读取数据库所有连接信息失败 %v", err)
		return
	}
	cache.ConnectList.Update(func(_ []model.Connect) []model.Connect {
		return dbsToFormatData(lst, "")
	})
}

// 将 连接 保存为 上下级关系的格式。
func dbsToFormatData(dblists []model.Connect, pid string) []model.Connect {
	var swapList []model.Connect
	for _, value := range dblists {
		var swap = value.Db2Connect()
		if value.Pid == pid {
			swap.Son = dbsToFormatData(dblists, value.Id)
			swapList = append(swapList, swap)
		}
	}

	return swapList
}
