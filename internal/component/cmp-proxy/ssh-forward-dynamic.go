package cmp_proxy

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/helays/utils/v2/close/vclose"
)

func (p *proxyConnect) forwardDynamic(conn net.Conn) {
	defer vclose.Close(conn)
	var b [1024]byte
	n, err := conn.Read(b[:])
	if err != nil {
		if !errors.Is(err, io.EOF) {
			p.error("forwardDynamic 目标数据读取失败 %v", err)
		}
		return
	}
	_, _ = conn.Write([]byte{0x05, 0x00})
	n, err = conn.Read(b[:])
	if err != nil {
		if !errors.Is(err, io.EOF) {
			p.error("forwardDynamic 第一次添加数据 %v", err)
		}
		return
	}
	var addr string
	switch b[3] {
	case 0x01:
		sip := sockIP{}
		if err = binary.Read(bytes.NewReader(b[4:n]), binary.BigEndian, &sip); err != nil {
			p.error("forwardDynamic 0x01 请求解析错误 %v", err)
			return
		}
		addr = sip.toAddr()
	case 0x03:
		host := string(b[5 : n-2])
		var port uint16
		err = binary.Read(bytes.NewReader(b[n-2:n]), binary.BigEndian, &port)
		if err != nil {
			p.error("forwardDynamic 0x03 错误 %v", err)
			return
		}
		addr = fmt.Sprintf("%s:%d", host, port)
	}
	if p.client == nil {
		return
	}

	server, err := p.client.Dial("tcp", addr)
	defer vclose.Close(server)
	if err != nil {
		p.error("forwardDynamic 动态转发连接目标失败 %v", err)
		return
	}
	_, _ = conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

	// 双向数据转发
	transfer(server, conn)
}
