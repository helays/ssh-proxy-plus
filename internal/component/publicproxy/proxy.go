package publicproxy

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/helays/ssh-proxy-plus/configs"
	cmp_proxy "github.com/helays/ssh-proxy-plus/internal/component/cmp-proxy"
	dal_proxy "github.com/helays/ssh-proxy-plus/internal/dal/dal-proxy"
	"golang.org/x/net/proxy"
	"helay.net/go/utils/v3/close/httpClose"
	"helay.net/go/utils/v3/close/vclose"
	"helay.net/go/utils/v3/logger/ulogs"
	"helay.net/go/utils/v3/safe"
)

func init() {
	proxy.RegisterDialerType("http", func(u *url.URL, dialer proxy.Dialer) (proxy.Dialer, error) {
		return &httpProxyDialer{
			proxyURI:  u,
			forwarder: dialer,
		}, nil
	})
	proxy.RegisterDialerType("https", func(u *url.URL, dialer proxy.Dialer) (proxy.Dialer, error) {
		return &httpProxyDialer{
			proxyURI:  u,
			forwarder: dialer,
		}, nil
	})
	// 注册SOCKS4代理
	proxy.RegisterDialerType("socks4", func(u *url.URL, d proxy.Dialer) (proxy.Dialer, error) {
		return newSocks4Dialer(u, d)
	})
}

var publicProxy = safe.NewResource(new(""))

func SetPublicProxy(addr string) {
	publicProxy.Write(&addr)
}
func GetPublicProxy() string {
	return *(publicProxy.Read())
}

type proxyServer struct {
	socks5 string
	http   string
}

func RunProxy(ctx context.Context) {
	cfg := configs.Get().Common
	if !cfg.EnablePublicProxy {
		return
	}
	if addr, err := dal_proxy.BestProxy(); err == nil {
		SetPublicProxy(addr)
		ulogs.Info("最优代理", addr)
	}
	go check(ctx) // 运行检测服务
	serv := proxyServer{
		socks5: cfg.Socks5Port,
		http:   cfg.HttpPort,
	}

	if serv.socks5 != "" {
		ulogs.Info("公共socks5代理启用", serv.socks5)
		go serv.startSocks5Server(ctx)
	}

	if serv.http != "" {
		ulogs.Info("公共http代理启用", serv.http)
		go serv.startHttpServer(ctx)
	}
}

func (p *proxyServer) startSocks5Server(ctx context.Context) {
	lc := net.ListenConfig{}
	listener, lErr := lc.Listen(ctx, "tcp", p.socks5)
	if lErr != nil {
		panic(fmt.Errorf("socks5公共代理端口监听失败 %v", lErr))
		return
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go p.handleSOCKS5(conn)

	}
}

func (p *proxyServer) handleSOCKS5(clientConn net.Conn) {
	defer vclose.Close(clientConn)
	px := GetPublicProxy()
	if px == "" {
		ulogs.Error("公共代理未设置")
		return
	}
	_, dialer, err := parseProxyAddr(px, 0)
	if err != nil {
		ulogs.Errorf("公共代理 %s 拨号器创建失败 %v\n", px, err)
		return
	}
	if err = cmp_proxy.ForwardDynamic(clientConn, dialer); err != nil {
		ulogs.Errorf("公共代理 %s 转发失败 %v\n", px, err)
	}
}

func (p *proxyServer) startHttpServer(ctx context.Context) {
	server := http.Server{
		Addr: p.http,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			px := GetPublicProxy()
			if px == "" {
				ulogs.Error("公共代理未设置")
				return
			}
			_, dialer, err := parseProxyAddr(px, 0)
			if err != nil {
				ulogs.Errorf("公共代理 %s 拨号器创建失败 %v\n", px, err)
				return
			}
			if r.Method == http.MethodConnect {
				cmp_proxy.HandleTunneling(w, r, dialer)
			} else {
				cmp_proxy.HandleHTTP(w, r, dialer)
			}
		}),
	}
	go func() {
		<-ctx.Done()
		httpClose.Server(&server)
	}()
	if err := server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			ulogs.Error("HTTP代理启动失败", err)
		}
	}
}
