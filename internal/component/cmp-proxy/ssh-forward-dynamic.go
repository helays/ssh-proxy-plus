package cmp_proxy

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"

	"golang.org/x/net/proxy"
	"helay.net/go/utils/v3/close/vclose"
)

func (p *proxyConnect) forwardDynamic(conn net.Conn, dialer proxy.Dialer) {
	err := ForwardDynamic(conn, dialer)
	if err != nil {
		p.error("%v", err)
	}
}

func ForwardDynamic(conn net.Conn, dialer proxy.Dialer) error {
	defer vclose.Close(conn)
	var b [1024]byte
	n, err := conn.Read(b[:])
	if err != nil {
		if !errors.Is(err, io.EOF) {
			return fmt.Errorf("forwardDynamic 目标数据读取失败 %v", err)
		}
		return nil
	}
	_, _ = conn.Write([]byte{0x05, 0x00})
	n, err = conn.Read(b[:])
	if err != nil {
		if !errors.Is(err, io.EOF) {
			return fmt.Errorf("forwardDynamic 第一次添加数据 %v", err)
		}
		return nil
	}
	var addr string
	switch b[3] {
	case 0x01:
		sip := sockIP{}
		if err = binary.Read(bytes.NewReader(b[4:n]), binary.BigEndian, &sip); err != nil {
			return fmt.Errorf("forwardDynamic 0x01 请求解析错误 %v", err)
		}
		addr = sip.toAddr()
	case 0x03:
		host := string(b[5 : n-2])
		var port uint16
		err = binary.Read(bytes.NewReader(b[n-2:n]), binary.BigEndian, &port)
		if err != nil {
			return fmt.Errorf("forwardDynamic 0x03 错误 %v", err)
		}
		addr = fmt.Sprintf("%s:%d", host, port)
	}
	if dialer == nil {
		return nil
	}

	server, err := dialer.Dial("tcp", addr)
	defer vclose.Close(server)
	if err != nil {
		return fmt.Errorf("forwardDynamic 动态转发连接目标失败 %v", err)
	}
	_, _ = conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

	// 双向数据转发
	Transfer(server, conn)
	return nil
}
