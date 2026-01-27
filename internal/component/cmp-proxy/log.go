package cmp_proxy

import (
	"fmt"
	"time"

	"github.com/helays/utils/v2/crypto/md5"
	"github.com/helays/utils/v2/logger/ulogs"
)

type Logs struct {
	Time time.Time
	Msg  string
	Hash string
	Flag string
}

func (p *proxyConnect) logs(format string, a ...any) {
	prefix := fmt.Sprintf("【%s】[%s] 隧道，服务地址[%s]，监听端口[%s]", p.connect.Lname, p.connect.Connect.String(), p.connect.Saddr, p.connect.Listen)
	info := fmt.Sprintf(format, a...)
	ulogs.Infof("%s %s", prefix, info)
	now := time.Now()
	msg := fmt.Sprintf("【INFO】%s %s %s", now.Format(time.DateTime), prefix, info)
	p.buffer.Push(Logs{
		Time: now,
		Msg:  msg,
		Hash: md5.Md5string(msg),
		Flag: "INFO",
	})
}

func (p *proxyConnect) error(format string, a ...any) {
	prefix := fmt.Sprintf("【%s】[%s] 隧道，服务地址[%s]，监听端口[%s]", p.connect.Lname, p.connect.Connect.String(), p.connect.Saddr, p.connect.Listen)
	info := fmt.Sprintf(format, a...)
	ulogs.Errorf("%s %s", prefix, info)
	now := time.Now()
	msg := fmt.Sprintf("【ERROR】%s %s %s", time.Now().Format(time.DateTime), prefix, info)
	p.buffer.Push(Logs{
		Time: now,
		Msg:  msg,
		Hash: md5.Md5string(msg),
		Flag: "ERROR",
	})
}
