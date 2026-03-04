package publicproxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/helays/ssh-proxy-plus/configs"
	"github.com/helays/ssh-proxy-plus/internal/dal/dal-proxy"
	"github.com/helays/ssh-proxy-plus/internal/model"
	"golang.org/x/net/proxy"
	"helay.net/go/utils/v3/close/httpClose"
	"helay.net/go/utils/v3/close/vclose"
	"helay.net/go/utils/v3/dataType"
	"helay.net/go/utils/v3/logger/ulogs"
	"helay.net/go/utils/v3/net/http/client/simpleHttpClient"
)

const connectURI = "google.com:80"

// const speedTestURI = "http://speedtest.tele2.net/1MB.zip"
const speedTestURI = "https://www.google.com"

func check(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		doCheck(ctx)
		time.Sleep(time.Minute * 10)
	}
}

func doCheck(ctx context.Context) {
	// 获取所有代理
	proxyList, err := dal_proxy.FindAllProxes()
	if err != nil {
		ulogs.Errorf("获取所有代理失败 %v", err)
		return
	}
	proxyTotals := len(proxyList)
	startTime := time.Now()
	wg := sync.WaitGroup{}
	maxConcurrent := 10 // 设置最大并发数，可以根据需要调整
	semaphore := make(chan struct{}, maxConcurrent)
	for i, _ := range proxyList {
		semaphore <- struct{}{}
		wg.Add(1)
		go func(idx int) {
			defer func() {
				wg.Done()
				<-semaphore
			}()
			p := proxyList[idx]
			if upErr := DoCheckProxy(ctx, p); upErr != nil {
				ulogs.Errorf("更新代理失败 %v", upErr)
			}
			//ulogs.Infof("代理 %s 检测完成,结果 %s,可用 %v,延迟 %v,速度 %d,综合评分 %.2f\n", p.Address, p.Message, p.IsAlive, p.Latency, p.Speed, p.Score)
		}(i)

	}
	if proxyTotals > 0 {
		wg.Wait()
	}
	UpdateBestProxy() // 更新最优代理
	ulogs.Infof("代理检测完成,总数 %d,耗时 %v\n", proxyTotals, time.Since(startTime))
}

// DoCheckProxy 检测代理
func DoCheckProxy(ctx context.Context, p model.ProxyInfo) error {
	p.IsAlive = dataType.NewBool(false)
	p.PortOpen = dataType.NewBool(false)
	p.Latency = 0
	p.Speed = 0
	p.Score = 0
	p.LastCheck = new(dataType.NewCustomTimeNow())
	cfg := configs.Get().Common
	pc := proxyCheck{ctx: ctx, proxy: &p, connectTimeout: cfg.ConnectTimeout, speedTestTimeout: cfg.SpeedTestTimeout}
	pc.check()
	return dal_proxy.UpdateProxy(pc.proxy)
}

type proxyCheck struct {
	ctx              context.Context
	proxy            *model.ProxyInfo
	connectTimeout   time.Duration // 连接超时
	speedTestTimeout time.Duration // 测速超时
}

// 解析代理地址
func parseProxyAddr(addr string, timeout time.Duration) (*url.URL, proxy.Dialer, error) {
	var dialer proxy.Dialer
	u, err := url.Parse(addr)
	if err != nil {
		return nil, nil, fmt.Errorf("代理地址格式错误 %v", err)
	}
	var forward proxy.Dialer = proxy.Direct
	if timeout > 0 {
		forward = &net.Dialer{
			Timeout: timeout,
			Resolver: &net.Resolver{
				PreferGo: true,
				Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
					d := net.Dialer{Timeout: timeout}
					return d.DialContext(ctx, network, address)
				},
			},
		}
	}
	dialer, err = proxy.FromURL(u, forward)
	if err != nil {
		return nil, nil, fmt.Errorf("代理创建失败 %v", err)
	}
	return u, dialer, nil
}

func (p *proxyCheck) check() {
	u, dialer, err := parseProxyAddr(p.proxy.Address, p.connectTimeout)
	if err != nil {
		p.proxy.Message = err.Error()
		return
	}
	// 测试代理服务是否可用
	if err = p.testProxyPort(u); err != nil {
		p.proxy.Message = err.Error()
		return
	}

	p.proxy.PortOpen = dataType.NewBool(true)
	p.proxy.Latency, err = p.testConnect(dialer)
	if err != nil {
		p.proxy.Message = err.Error()
		return
	}
	// 测试下载速度
	p.proxy.Speed, err = p.testSpeed()
	if err != nil {
		p.proxy.Message = err.Error()
		return
	}
	p.proxy.IsAlive = dataType.NewBool(true)
	p.proxy.Message = "代理可用"
	p.proxy.Score = p.calculateScore() // 计算综合评分

}

func (p *proxyCheck) testProxyPort(u *url.URL) error {
	conn, err := net.DialTimeout("tcp", u.Host, p.connectTimeout)
	defer vclose.Close(conn)
	if err != nil {
		return fmt.Errorf("代理服务已停用 %v", err)
	}
	return nil
}

func (p *proxyCheck) testConnect(dialer proxy.Dialer) (time.Duration, error) {
	var testF = func() (time.Duration, error) {
		ctx, cancel := context.WithTimeout(p.ctx, p.connectTimeout)
		defer cancel()
		ctxDialer, _ := dialer.(proxy.ContextDialer)
		// 计算连接延迟
		startTime := time.Now()
		conn, err := ctxDialer.DialContext(ctx, "tcp", connectURI)
		defer vclose.Close(conn)
		if err != nil {
			return 0, fmt.Errorf("检测失败 %v", err)
		}
		return time.Since(startTime), nil
	}
	var (
		latency time.Duration
		err     error
	)
	for i := 0; i < 3; i++ {
		latency, err = testF()
		if err == nil {
			return latency, nil
		}
		time.Sleep(5 * time.Second)
	}
	return 0, err
}

func (p *proxyCheck) testSpeed() (int64, error) {
	httpClient, _ := simpleHttpClient.New(p.speedTestTimeout, p.proxy.Address)
	var testF = func() (int64, error) {
		ctx, cancel := context.WithTimeout(p.ctx, p.speedTestTimeout)
		defer cancel()
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, speedTestURI, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36")
		startTime := time.Now()
		resp, err := httpClient.Do(req)
		defer httpClose.CloseResp(resp)
		if err != nil {
			return 0, fmt.Errorf("请求测速地址失败 %v", err)
		}
		n, err := io.Copy(io.Discard, resp.Body)
		if err != nil {
			return 0, fmt.Errorf("测速失败 %v", err)
		}
		// 计算速度
		duration := time.Since(startTime).Milliseconds()
		if duration == 0 {
			duration = 1
		}
		return n / duration, nil
	}
	var (
		speed int64
		err   error
	)
	for i := 0; i < 3; i++ {
		speed, err = testF()
		if err == nil {
			return speed, nil
		}
		time.Sleep(5 * time.Second)
	}
	return 0, err
}

// calculateScore 计算代理得分
func (p *proxyCheck) calculateScore() float64 {
	// 得分 = 1000/延迟(ms) + 速度/10
	// 延迟越低、速度越快，得分越高
	latencyScore := 1000.0 / float64(p.proxy.Latency.Milliseconds()+1)
	speedScore := float64(p.proxy.Speed) / 10.0

	return latencyScore + speedScore
}
