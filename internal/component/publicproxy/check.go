package publicproxy

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/helays/ssh-proxy-plus/configs"
	"github.com/helays/ssh-proxy-plus/internal/dal/dal-proxy"
	"github.com/helays/ssh-proxy-plus/internal/model"
	"github.com/helays/ssh-proxy-plus/internal/types"
	"golang.org/x/net/proxy"
	"helay.net/go/utils/v3/close/httpClose"
	"helay.net/go/utils/v3/close/vclose"
	"helay.net/go/utils/v3/dataType"
	"helay.net/go/utils/v3/logger/ulogs"
)

const connectURI = "github.com:80"
const speedTestURI = "http://speedtest.tele2.net/10MB.zip"

func check(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		doCheck(ctx)
		time.Sleep(time.Minute)
	}
}

func doCheck(ctx context.Context) {
	// 获取所有代理
	proxyList, err := dal_proxy.FindAllProxes()
	if err != nil {
		ulogs.Errorf("获取所有代理失败 %v", err)
		return
	}
	startTime := time.Now()
	cfg := configs.Get()
	for _, p := range proxyList {
		ulogs.Infof("正在检测代理 %s\n", p.Address)
		pc := proxyCheck{ctx: ctx, proxy: p, checkTimeout: cfg.ProxyCheckTimeout}
		pc.check()
		if err = dal_proxy.UpdateProxy(pc.proxy); err != nil {
			ulogs.Errorf("更新代理失败 %v", err)
		}
	}
	ulogs.Info("代理检测完成,总数 %d,耗时 %v", len(proxyList), time.Since(startTime))
}

type proxyCheck struct {
	ctx          context.Context
	proxy        *model.ProxyInfo
	checkTimeout time.Duration
}

func (p *proxyCheck) check() {
	// 计算连接延迟
	startTime := time.Now()
	ctx, cancel := context.WithTimeout(p.ctx, p.checkTimeout)
	defer cancel()

	var dialer proxy.Dialer
	var err error

	u, err := url.Parse(p.proxy.Address)
	if err != nil {
		p.proxy.IsAlive = dataType.NewBool(false)
		p.proxy.Message = "代理地址格式错误"
		return
	}

	switch p.proxy.Type {
	case types.ProxySocks5, types.ProxySocks4:
		dialer, err = proxy.FromURL(u, proxy.Direct)
	case types.ProxyHttp, types.ProxyHttps:
		dialer = p.createHTTPProxyDialer(u)
	default:
		p.proxy.IsAlive = dataType.NewBool(false)
	}
	if err != nil {
		p.proxy.IsAlive = dataType.NewBool(false)
		p.proxy.Message = "创建代理拨号器失败"
		return
	}
	conn, err := dialer.Dial("tcp", connectURI)
	defer vclose.Close(conn)
	if err != nil {
		p.proxy.IsAlive = dataType.NewBool(false)
		p.proxy.Message = fmt.Sprintf("代理 %s 检测失败 %v", p.proxy.Address, err)
		return
	}
	p.proxy.Latency = time.Since(startTime) // 计算时延
	ulogs.Infof("代理 %s 延迟 %v", p.proxy.Address, p.proxy.Latency)

	// 测试下载速度
	p.proxy.Speed, err = p.testSpeed(ctx, dialer)
	if err != nil {
		p.proxy.IsAlive = dataType.NewBool(false)
		p.proxy.Message = fmt.Sprintf("代理 %s 测速失败 %v", p.proxy.Address, err)
		return
	}
	p.proxy.IsAlive = dataType.NewBool(true)
	p.proxy.Message = "代理可用"
	p.proxy.Score = p.calculateScore() // 计算综合评分

}

func (p *proxyCheck) createHTTPProxyDialer(u *url.URL) proxy.Dialer {
	// 创建HTTP transport
	transport := &http.Transport{
		Proxy: http.ProxyURL(u),
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// 返回自定义dialer
	return &httpProxyDialer{transport: transport}
}

func (p *proxyCheck) testSpeed(ctx context.Context, dialer proxy.Dialer) (int64, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.Dial(network, addr)
			},
		},
		Timeout: p.checkTimeout,
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, speedTestURI, nil)
	startTime := time.Now()
	resp, err := httpClient.Do(req)
	defer httpClose.CloseResp(resp)
	if err != nil {
		return 0, fmt.Errorf("代理 %s 请求测速地址失败 %v", p.proxy.Address, err)
	}
	n, err := io.Copy(io.Discard, resp.Body)
	if err != nil {
		return 0, fmt.Errorf("代理 %s 测速失败 %v", p.proxy.Address, err)
	}
	// 计算速度
	duration := time.Since(startTime).Milliseconds()
	if duration == 0 {
		duration = 1
	}
	return n / duration, nil
}

// calculateScore 计算代理得分
func (p *proxyCheck) calculateScore() float64 {
	// 得分 = 1000/延迟(ms) + 速度/10
	// 延迟越低、速度越快，得分越高
	latencyScore := 1000.0 / float64(p.proxy.Latency.Milliseconds()+1)
	speedScore := float64(p.proxy.Speed) / 10.0

	return latencyScore + speedScore
}
