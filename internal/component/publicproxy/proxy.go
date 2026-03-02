package publicproxy

import (
	"context"
)

func RunProxy(ctx context.Context) {
	check(ctx) // 运行检测服务
}
