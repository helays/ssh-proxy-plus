package configs

import (
	"path"
	"time"

	"helay.net/go/utils/v3/config/loadAuto"
	"helay.net/go/utils/v3/logger/ulogs"
	"helay.net/go/utils/v3/tools"
)

var conf = new(Config)

func Init() {
	loadAuto.Load(conf)
	setDefault()
	ulogs.Info("配置文件载入成功...")
}

func setDefault() {
	conf.Common.Cache = tools.Ternary(conf.Common.Cache == "", "cache", conf.Common.Cache)
	conf.Common.Cache = tools.Fileabs(conf.Common.Cache)
	ulogs.DieCheckerr(tools.Mkdir(conf.Common.Cache), "创建缓存目录失败！")
	if len(conf.Db.Host) == 0 {
		conf.Db.Host = []string{path.Join(conf.Common.Cache, "proxy.db")}
	} else {
		conf.Db.Host = []string{path.Join(conf.Common.Cache, conf.Db.Host[0])}
	}
	conf.Common.HeartBeat = tools.AutoTimeDuration(conf.Common.HeartBeat, time.Second, 10*time.Second)
	conf.Common.SshTimeout = tools.AutoTimeDuration(conf.Common.SshTimeout, time.Second, 30*time.Second)
	conf.Common.RingBufferLogSize = tools.Ternary(conf.Common.RingBufferLogSize == 0, 1024, conf.Common.RingBufferLogSize)
}

func Get() *Config {
	return conf
}
