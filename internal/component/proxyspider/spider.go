package proxyspider

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/helays/ssh-proxy-plus/configs"
	"github.com/helays/ssh-proxy-plus/internal/cache"
	"github.com/helays/ssh-proxy-plus/internal/dal/dal-proxy"
	"github.com/helays/ssh-proxy-plus/internal/model"
	"github.com/helays/ssh-proxy-plus/internal/types"
	"helay.net/go/utils/v3/close/httpClose"
	"helay.net/go/utils/v3/logger/ulogs"
	"helay.net/go/utils/v3/net/http/client/simpleHttpClient"
	"helay.net/go/utils/v3/tools"
)

type spider struct {
	enable bool
	name   string
	addr   string
	ticket time.Duration // 爬虫间隔
}

var spiderList = []spider{
	{
		enable: true,
		name:   "tomcat1235",
		addr:   "https://tomcat1235.nyc.mn/proxy_list",
		ticket: time.Minute * 10,
	},
}

var httpClient *http.Client

func RunSpider(ctx context.Context) {
	if !configs.Get().Common.EnablePublicProxy {
		return
	}
	ulogs.Info("公共代理服务启用，启动代理爬虫服务")
	proxyAddr := ""
	if info, ok := cache.SysConfig.Load(types.ProxyAddr); ok && info != nil && info.Value != "" {
		proxyAddr = info.Value
	}
	var err error
	httpClient, err = simpleHttpClient.New(time.Minute*2, proxyAddr)
	if err != nil {
		panic(fmt.Errorf("代理爬虫http client初始化失败 %v", err))
		return
	}
	for _, s := range spiderList {
		tools.RunAsyncTickerWithContext(ctx, s.enable, s.ticket, s.run)
	}
}
func (s *spider) error(format string, a ...any) {
	ulogs.Errorf("代理爬虫 %s %s\n", s.addr, fmt.Sprintf(format, a...))
}

func (s *spider) run(ctx context.Context) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, s.addr, nil)
	resp, err := httpClient.Do(req)
	defer httpClose.CloseResp(resp)
	if err != nil {
		s.error("请求失败 %v", err)
		return
	}
	byt, err := io.ReadAll(resp.Body)
	if err != nil {
		s.error("读取响应失败 %v", err)
		return
	}
	body := string(byt)
	if resp.StatusCode != http.StatusOK {
		s.error("响应状态码错误 %d %s", resp.StatusCode, body)
		return
	}
	switch s.name {
	case "tomcat1235":
		s.tomcat1235(body)
	}
}

func (s *spider) tomcat1235(body string) {

	var (
		trRegexp        = regexp.MustCompile(`<tr>([\s\S]+?)</tr>`)
		proxyTypeRegexp = regexp.MustCompile(`badge badge-type text-uppercase">(\w+?)</span>`)
		proxyAddrRegexp = regexp.MustCompile(`data-ip="([\d.]+?)" data-port="(\d+?)"`)
	)

	for _, tr := range trRegexp.FindAllStringSubmatch(body, -1) {
		if len(tr) != 2 {
			continue
		}
		var proxyInfo = model.ProxyInfo{}
		if proxyTypeResult := proxyTypeRegexp.FindStringSubmatch(tr[1]); len(proxyTypeResult) != 2 {
			continue
		} else {
			proxyInfo.Type = types.ProxyType(strings.TrimSpace(proxyTypeResult[1]))
		}
		if proxyAddrResult := proxyAddrRegexp.FindStringSubmatch(tr[1]); len(proxyAddrResult) != 3 {
			continue
		} else {
			proxyInfo.Address = fmt.Sprintf("%s://%s:%s", proxyInfo.Type, proxyAddrResult[1], proxyAddrResult[2])
		}
		if err := dal_proxy.SaveProxy(&proxyInfo); err != nil {
			s.error("保存代理失败 %v", err)
		}
	}
	ulogs.Infof("代理爬虫 %s 爬取完成", s.addr)

}
