package publicproxy

import (
	"fmt"
	"net"
	"net/url"

	"golang.org/x/net/proxy"
	"helay.net/go/utils/v3/close/vclose"
)

// HTTP代理拨号器
type httpProxyDialer struct {
	proxyURI  *url.URL
	forwarder proxy.Dialer
}

func (d *httpProxyDialer) Dial(network, addr string) (net.Conn, error) {
	// 连接到HTTP代理
	proxyConn, err := d.forwarder.Dial("tcp", d.proxyURI.Host)
	if err != nil {
		return nil, err
	}

	// 发送HTTP CONNECT请求
	req := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n\r\n", addr, addr)
	if _, err = proxyConn.Write([]byte(req)); err != nil {
		vclose.Close(proxyConn)
		return nil, err
	}

	// 读取响应
	resp := make([]byte, 1024)
	n, err := proxyConn.Read(resp)
	if err != nil {
		vclose.Close(proxyConn)
		return nil, err
	}

	// 检查响应状态
	if string(resp[:7]) != "HTTP/1.1 200" && string(resp[:7]) != "HTTP/1.0 200" {
		vclose.Close(proxyConn)
		return nil, fmt.Errorf("proxy returned non-200 status: %s", string(resp[:n]))
	}

	return proxyConn, nil
}
