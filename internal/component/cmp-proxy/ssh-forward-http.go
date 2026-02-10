package cmp_proxy

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"time"

	"golang.org/x/net/proxy"
	"helay.net/go/utils/v3/close/httpClose"
	"helay.net/go/utils/v3/close/vclose"
	"helay.net/go/utils/v3/net/http/httpkit"
)

func (p *proxyConnect) forwardHTTP() {
	p.logs("开始启动HTTP代理")
	if p.client == nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// 创建基于SSH的SOCKS5拨号器
	dialer := &sshDialer{sshClient: p.client}
	server := http.Server{
		Addr: p.connect.Listen,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodConnect {
				p.handleTunneling(w, r, dialer)
			} else {
				p.handleHTTP(w, r, dialer)
			}
		}),
	}
	go p.quit(ctx, cancel)
	go func() {
		<-ctx.Done()
		httpClose.Server(&server)
	}()
	if err := server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			p.error("http启动失败 %v", err)
		}
	}
	p.logs("HTTP代理停止")
}

func (p *proxyConnect) handleTunneling(w http.ResponseWriter, r *http.Request, dialer proxy.Dialer) {
	destConn, err := dialer.Dial("tcp", r.Host)
	defer vclose.Close(destConn)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hijacker.Hijack()
	defer vclose.Close(clientConn)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
	transfer(destConn, clientConn)
}

func (p *proxyConnect) handleHTTP(w http.ResponseWriter, r *http.Request, dialer proxy.Dialer) {
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		},
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 20 * time.Second,
	}
	p.prepareProxyRequest(r)
	client := &http.Client{
		Transport: transport,
	}
	resp, err := client.Do(r)
	defer httpClose.CloseResp(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	httpkit.RespCloneHeader(w, resp.Header)
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)

}

func (p *proxyConnect) prepareProxyRequest(r *http.Request) {
	r.RequestURI = "" // 必须清空

	// 确保URL有完整的主机信息
	if r.URL.Host == "" {
		r.URL.Host = r.Host
	}

	// 设置Scheme（根据实际情况）
	if r.URL.Scheme == "" {
		if r.TLS != nil {
			r.URL.Scheme = "https"
		} else {
			r.URL.Scheme = "http"
		}
	}

	// 更安全的端口处理
	host, port, err := net.SplitHostPort(r.URL.Host)
	if err != nil {
		// 如果没有端口，添加默认端口
		switch r.URL.Scheme {
		case "https":
			r.URL.Host = net.JoinHostPort(r.URL.Host, "443")
		default:
			r.URL.Host = net.JoinHostPort(r.URL.Host, "80")
		}
	} else if port == "" {
		// 有冒号但没有端口号的情况
		switch r.URL.Scheme {
		case "https":
			r.URL.Host = net.JoinHostPort(host, "443")
		default:
			r.URL.Host = net.JoinHostPort(host, "80")
		}
	}
}
