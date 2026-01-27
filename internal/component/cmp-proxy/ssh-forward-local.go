package cmp_proxy

import (
	"net"

	"github.com/helays/utils/v2/close/vclose"
)

// 本地转发
func (p *proxyConnect) forwardLocal(conn net.Conn) {
	defer vclose.Close(conn)
	if p.client == nil {
		return
	}

	server, err := p.client.Dial("tcp", p.connect.Remote)
	defer vclose.Close(server)
	if err != nil {
		p.error("本地转发，目标地址连接失败 %s %v", p.connect.Remote, err)
		return
	}
	// 双向数据转发
	transfer(server, conn)

}
