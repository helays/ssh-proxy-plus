package dal

import (
	"github.com/helays/ssh-proxy-plus/configs"
	"gorm.io/gorm"
	"helay.net/go/utils/v3/db/userDb/connect/sqliteconnect"
	"helay.net/go/utils/v3/logger/ulogs"
)

var db *gorm.DB

func Init() {
	cfg := configs.Get()
	var err error
	db, err = sqliteconnect.InitDB(&cfg.Db)
	ulogs.DieCheckerr(err, "数据库连接失败")
	ulogs.Info("数据库连接成功...")
}

func GetDB() *gorm.DB {
	return db.Session(&gorm.Session{})
}
