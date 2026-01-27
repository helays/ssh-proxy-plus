package cmp_proxy

import (
	"context"
	"github.com/helays/ssh-proxy-plus/internal/types"
	"net"
	"strconv"
	"strings"

	"github.com/helays/utils/v2/close/vclose"
)

func (p *proxyConnect) forwardRemote() {
	defer p.logs("隧道反向代理停止")
	p.logs("开始启动远程代理 %s", p.connect.Remote)
	if p.client == nil {
		return
	}
	var remotePort string
	if remotePorts := strings.Split(p.connect.Remote, ":"); len(remotePorts) == 2 {
		remotePort = remotePorts[1]
	}

	if remotePort != "" {
		port, err := strconv.Atoi(remotePort)
		if err != nil {
			p.error("端口格式错误 %v", err)
			return
		}
		pid, err := checkPortAndGetPID(p.client, port)
		if err != nil {
			p.error("检查远程端口占用失败 %v", err)
			return
		}
		if pid > 0 {
			if err = killProcess(p.client, pid); err != nil {
				p.error("kill进程失败 %v", err)
				return
			}
		}
	}

	server, err := p.client.Listen("tcp", p.connect.Remote)
	defer vclose.Close(server)
	if err != nil {
		vclose.Close(p.client)
		p.reConnectCancel()
		p.SetStatus(types.ConnectStatusRe)
		p.error("远程地址 %s 监听失败 %v", p.connect.Remote, err)
		return
	}
	p.logs("远程代理启动成功 %s", p.connect.Remote)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go p.quit(ctx, func() {
		cancel()
		vclose.Close(server)
	})
	for {
		client, accErr := server.Accept()
		if accErr != nil {
			p.error("远程转发 %s TCP 远程数据接收失败 %v", p.connect.Remote, err)
			return
		}
		p.logs("远程转发收到请求 远程[%s] 本地[%s]", client.RemoteAddr().String(), client.LocalAddr().String())
		go func(conn net.Conn) {
			defer vclose.Close(conn)
			_s, _e := net.Dial("tcp", p.connect.Listen)
			defer vclose.Close(_s)
			if _e != nil {
				p.error("远程转发 %s TCP 监听失败 %v", p.connect.Listen, _e)
				return
			}
			transfer(_s, conn)

		}(client)
	}
}
