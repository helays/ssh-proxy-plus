package auto_migrate

import (
	"github.com/helays/ssh-proxy-plus/internal/dal"
	"github.com/helays/ssh-proxy-plus/internal/model"

	"helay.net/go/utils/v3/db/userDb"
)

func AutoMigrate() {
	db := dal.GetDB()
	userDb.AutoCreateTableWithStruct(db, model.Connect{}, "初始化隧道连接配置表失败")
	userDb.AutoCreateTableWithStruct(db, model.SysConfig{}, "初始化系统配置表失败")
	userDb.AutoCreateTableWithStruct(db, model.AliEcsOrder{}, "初始化阿里云 ECS 订单表失败")

}
