package publicproxy

import (
	"encoding/binary"
	"fmt"
	"net"
	"net/url"

	"golang.org/x/net/proxy"
	"helay.net/go/utils/v3/close/vclose"
)

// SOCKS4代理拨号器
type socks4Dialer struct {
	proxyURI  *url.URL
	forwarder proxy.Dialer
	username  string
}

func newSocks4Dialer(u *url.URL, forwarder proxy.Dialer) (*socks4Dialer, error) {
	d := &socks4Dialer{
		proxyURI:  u,
		forwarder: forwarder,
	}
	if u.User != nil {
		d.username = u.User.Username()
	}
	return d, nil
}

func (d *socks4Dialer) Dial(network, addr string) (net.Conn, error) {
	if network != "tcp" && network != "tcp4" {
		return nil, fmt.Errorf("socks4 only supports tcp connections")
	}

	// 解析目标地址
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	// 解析端口
	port, err := net.LookupPort("tcp", portStr)
	if err != nil {
		return nil, err
	}

	// 解析主机IP
	var ip net.IP
	if ip = net.ParseIP(host); ip == nil {
		// 如果是域名，需要先解析（SOCKS4不支持域名）
		ips, err := net.LookupIP(host)
		if err != nil {
			return nil, err
		}
		ip = ips[0]
	}
	ip = ip.To4()
	if ip == nil {
		return nil, fmt.Errorf("socks4 only supports IPv4")
	}

	// 连接到SOCKS4代理
	conn, err := d.forwarder.Dial("tcp", d.proxyURI.Host)
	if err != nil {
		return nil, err
	}

	// 构建SOCKS4请求包
	// +----+----+----+----+----+----+----+----+----+----+....+----+
	// | VN | CD | DSTPORT |      DSTIP        | USERID       |NULL|
	// +----+----+----+----+----+----+----+----+----+----+....+----+
	//  VN: 0x04, CD: 0x01 (CONNECT)
	packet := make([]byte, 0, 9+len(d.username))
	packet = append(packet, 0x04)                                // VN
	packet = append(packet, 0x01)                                // CD
	packet = binary.BigEndian.AppendUint16(packet, uint16(port)) // DSTPORT
	packet = append(packet, ip...)
	packet = append(packet, []byte(d.username)...)
	packet = append(packet, 0x00)

	// 发送请求
	if _, err = conn.Write(packet); err != nil {
		vclose.Close(conn)
		return nil, err
	}

	// 读取响应
	resp := make([]byte, 8)
	if _, err = conn.Read(resp); err != nil {
		vclose.Close(conn)
		return nil, err
	}

	// 检查响应
	if resp[0] != 0x00 {
		vclose.Close(conn)
		return nil, fmt.Errorf("socks4 server returned invalid VN: %d", resp[0])
	}

	switch resp[1] {
	case 0x5a:
		// 请求成功
		return conn, nil
	case 0x5b:
		vclose.Close(conn)
		return nil, fmt.Errorf("socks4 request rejected or failed")
	case 0x5c:
		vclose.Close(conn)
		return nil, fmt.Errorf("socks4 request failed because client is not running identd (or not reachable from server)")
	case 0x5d:
		vclose.Close(conn)
		return nil, fmt.Errorf("socks4 request failed because client's identd could not confirm the user ID string")
	default:
		vclose.Close(conn)
		return nil, fmt.Errorf("socks4 server returned unknown CD: %d", resp[1])
	}
}
