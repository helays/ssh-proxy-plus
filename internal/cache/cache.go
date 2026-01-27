package cache

import (
	"context"
	"github.com/helays/ssh-proxy-plus/internal/model"

	"github.com/helays/utils/v2/safe"
)

var (
	ConnectList = safe.NewResourceRWMutex([]model.Connect{}) // 记录数据库中所有的连接关系信息

	SysConfig *safe.Map[string, *model.SysConfig]
)

func Init(ctx context.Context) {
	SysConfig = safe.NewMap[string, *model.SysConfig](ctx, safe.StringHasher{})
}
