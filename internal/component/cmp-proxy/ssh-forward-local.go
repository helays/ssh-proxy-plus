package cmp_proxy

import (
	"net"

	"golang.org/x/net/proxy"
	"helay.net/go/utils/v3/close/vclose"
)

// 本地转发
func (p *proxyConnect) forwardLocal(conn net.Conn, dialer proxy.Dialer) {
	defer vclose.Close(conn)
	if dialer == nil {
		return
	}

	server, err := dialer.Dial("tcp", p.connect.Remote)
	defer vclose.Close(server)
	if err != nil {
		p.error("本地转发，目标地址连接失败 %s %v", p.connect.Remote, err)
		return
	}
	// 双向数据转发
	Transfer(server, conn)

}
