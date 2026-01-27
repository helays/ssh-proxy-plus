package service

import (
	"context"
	"encoding/json"
	"github.com/helays/ssh-proxy-plus/internal/api/dto"
	"github.com/helays/ssh-proxy-plus/internal/cache"
	cmp_proxy "github.com/helays/ssh-proxy-plus/internal/component/cmp-proxy"
	"github.com/helays/ssh-proxy-plus/internal/model"
	"io"
	"strings"
	"time"

	"github.com/helays/utils/v2/crypto/md5"
	"github.com/helays/utils/v2/logger/ulogs"
	"github.com/helays/utils/v2/net/http/request"
	"github.com/helays/utils/v2/safe"
	"golang.org/x/net/websocket"
)

type ServWS struct {
	ws           *websocket.Conn
	lastTime     *safe.ResourceRWMutex[time.Time]
	logCtx       context.Context
	logCtxCancel context.CancelFunc
}

func NewWS(ws *websocket.Conn) *ServWS {
	return &ServWS{ws: ws, lastTime: safe.NewResourceRWMutex(time.Now())}
}

func (s *ServWS) Service(_ctx context.Context) {
	ctx, cancel := context.WithCancel(_ctx)
	defer cancel()
	go s.responseForwardLst(ctx)
	go s.receive(ctx)
	// 检测是否掉线门限值
	const threshold = 10 * time.Second
	tck := time.NewTicker(threshold)
	defer tck.Stop()
	for {
		select {
		case <-tck.C:
			s.lastTime.ReadWith(func(t time.Time) {
				if time.Since(t) > threshold {
					ulogs.Infof("客户端[%s] 超时掉线", request.Getip(s.ws.Request()))
					cancel()
				}
			})
		case <-ctx.Done():
			return
		}
	}
}

// 定时更新列表数据
func (s *ServWS) responseForwardLst(ctx context.Context) {
	tck := time.NewTicker(time.Second * 1)
	defer tck.Stop()
	var respHash string
	for {
		select {
		case <-ctx.Done():
			return
		case <-tck.C:
			cache.ConnectList.ReadWith(func(connects []model.Connect) {
				res := dto.WsResp{Action: "list", Data: connects}
				byt, _ := json.Marshal(res)
				currentHash := md5.Md5(byt)
				if respHash == currentHash {
					return
				}
				respHash = currentHash
				_ = websocket.JSON.Send(s.ws, res)
			})
		}
	}
}

// 接收客户端数据
func (s *ServWS) receive(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			ulogs.Infof("客户端[%s] 连接断开", request.Getip(s.ws.Request()))
			return
		default:
			var rec string
			if err := websocket.Message.Receive(s.ws, &rec); err != nil {
				// 判断 err 是否An established connection was aborted by the software in your host machine.
				if err != io.EOF && !strings.Contains(err.Error(), "An established connection was aborted") {
					ulogs.Errorf("客户端[%s] 接收数据异常 %v", request.Getip(s.ws.Request()), err)
				}
				continue
			}
			s.lastTime.Write(time.Now())
			s.messageHandle(ctx, rec)
		}
	}
}

func (s *ServWS) messageHandle(ctx context.Context, message string) {
	receiveArr := strings.Split(message, "_")
	if len(receiveArr) == 2 {
		switch receiveArr[0] {
		case "show":
			s.logCtx, s.logCtxCancel = context.WithCancel(ctx)
			go s.responseLog(receiveArr[1])
			return
		case "close":
			if s.logCtxCancel != nil {
				s.logCtxCancel()
			}
			return
		}
	}
	res := dto.WsResp{Action: "status", Data: cmp_proxy.FindConnectStatus()}
	_ = websocket.JSON.Send(s.ws, res)
}

func (s *ServWS) responseLog(id string) {
	ringBuffer := cmp_proxy.GetLogRingBuffer(id)
	if ringBuffer == nil {
		_ = websocket.JSON.Send(s.ws, dto.WsResp{Action: "log", Data: "隧道日志查询失败"})
		return
	}
	var lastLog cmp_proxy.Logs
	for msg := range ringBuffer.Iterator() {
		// 【INFO】2025-09-27 13:09:49
		lastLog = msg
		_ = websocket.JSON.Send(s.ws, dto.WsResp{Action: "log", Data: msg.Msg})
	}
	// 接下来就需要动态监听日志，有更新就发送。
	tck := time.NewTicker(1 * time.Second)
	defer tck.Stop()
	for {
		select {
		case <-s.logCtx.Done():
			_ = websocket.JSON.Send(s.ws, dto.WsResp{Action: "log", Data: "日志查询结束"})
			return
		default:
			for _, msg := range ringBuffer.GetLast(100) {
				if msg.Time.After(lastLog.Time) {
					lastLog = msg
					_ = websocket.JSON.Send(s.ws, dto.WsResp{Action: "log", Data: msg.Msg})
				}
			}
		}
	}
}
