package cmp_proxy

import (
	"context"
	"fmt"
	"net"

	"github.com/helays/utils/v2/safe"
	"golang.org/x/crypto/ssh"
)

var connectMap *safe.Map[string, *proxyConnect] // 记录数据库中所有的连接客户端信息

func Init(ctx context.Context) {
	connectMap = safe.NewMap[string, *proxyConnect](ctx, safe.StringHasher{})
}

type sockIP struct {
	A, B, C, D byte
	PORT       uint16
}

func (ip sockIP) toAddr() string {
	return fmt.Sprintf("%d.%d.%d.%d:%d", ip.A, ip.B, ip.C, ip.D, ip.PORT)
}

// sshDialer 实现proxy.Dialer接口
type sshDialer struct {
	sshClient *ssh.Client
}

func (d *sshDialer) Dial(network, addr string) (net.Conn, error) {
	return d.sshClient.Dial(network, addr)
}
