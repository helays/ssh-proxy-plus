package cmp_proxy

import (
	"context"
	"fmt"
	"github.com/helays/ssh-proxy-plus/configs"
	"github.com/helays/ssh-proxy-plus/internal/model"
	"github.com/helays/ssh-proxy-plus/internal/types"
	"net"
	"sync"
	"time"

	"github.com/helays/utils/v2/close/vclose"
	"github.com/helays/utils/v2/tools/ringbuffer"
	"golang.org/x/crypto/ssh"
)

func GetLogRingBuffer(id string) *ringbuffer.RingBuffer[Logs] {
	client, ok := connectMap.Load(id)
	if !ok {
		return nil
	}
	return client.buffer
}

// 转发部分的核心逻辑：
// 1. 启动时候，根据配置，做好层级关系
// 2. 每一层都创建三个上下文，一个停止，一个重连，一个继承。
// 3. 收到停止信号、重连信号，都发送继承取消信号，子级就会退出循环
// 4. 父级别，重新启动，并根据层级关系，自动重启子级别。

func Stop(id string) {
	client, ok := connectMap.Load(id)
	if !ok {
		return
	}

	client.logs("人工触发隧道中断，停止代理")
	client.SetStatus(types.ConnectStatusDel)
	client.stopCancel()
	signal := make(chan struct{})
	go func() {
		tck := time.NewTicker(time.Second)
		defer tck.Stop()
		for range tck.C {
			// 检测到缓存中的数据已删除，说明已经退出完成。
			if _, ok = connectMap.Load(id); !ok {
				signal <- struct{}{}
				return
			}
		}
	}()
	<-signal
}

func StartList(cs []model.Connect) {
	for _, c := range cs {
		Start(c)
	}
}

func Start(c model.Connect) {
	if c.Active == types.TextStatusDisable {
		return
	}
	cfg := configs.Get()
	go func() {
		var pc = newProxy(c)
		defer pc.Close()

		for {
			select {
			case <-pc.parentReConnectCtx.Done(): // 监听父级重连信号，收到后，当前层级退出，并发送stop信号
				// 收到父级的重连信号，当前层级退出，并通过stopChannel 发送信号，子级别会自动中断
				// 如果通过递归一层一层的取消，效率太低
				pc.stopCancel()
				pc.logs("隧道收到父停止信号")
				return
			case <-pc.stopCtx.Done():
				pc.logs("隧道收到停止信号")
				return
			case <-pc.reConnectCtx.Done():
				pc.logs("隧道收到重连信号，%v 后自动重连", cfg.Common.HeartBeat)
				time.Sleep(cfg.Common.HeartBeat)
				if pc.stopCtx.Err() != nil || pc.parentReConnectCtx.Err() != nil {
					continue
				}
				pc = newProxy(c) // 重连的时候，需要重新初始化配置信息
				continue
			default:
				if pc.stopCtx.Err() != nil || pc.parentReConnectCtx.Err() != nil {
					continue
				}
				pc.logs("启动代理")
				pc.scheduler()
			}

		}
	}()
}

type proxyConnect struct {
	connect    model.Connect
	client     *ssh.Client         // ssh 客户端
	sshConfig  *ssh.ClientConfig   // 连接ssh 的配置
	status     types.ConnectStatus // 连接状态
	statusLock sync.RWMutex

	// 重连和停止上下文不继承，只控制当前层级

	reConnectCtx    context.Context // 用于控制重连的上下文
	reConnectCancel context.CancelFunc

	stopCtx    context.Context // 用于控制退出的上下文
	stopCancel context.CancelFunc

	parentReConnectCtx    context.Context // 父级重连信号，子级收到后 直接中断
	parentReConnectCancel context.CancelFunc

	buffer *ringbuffer.RingBuffer[Logs] // 环形缓冲区
}

func newProxy(c model.Connect) *proxyConnect {
	pc := proxyConnect{connect: c}
	pc.reConnectCtx, pc.reConnectCancel = context.WithCancel(context.Background())

	if parent, ok := connectMap.Load(c.Pid); ok {
		pc.parentReConnectCtx, pc.parentReConnectCancel = context.WithCancel(parent.reConnectCtx)
		// stop信号可以继承，避免一层一层中断
		pc.stopCtx, pc.stopCancel = context.WithCancel(parent.stopCtx)
	} else {
		pc.stopCtx, pc.stopCancel = context.WithCancel(context.Background())

		pc.parentReConnectCtx, pc.parentReConnectCancel = context.WithCancel(context.Background())
	}
	pc.buffer = ringbuffer.MustNew[Logs](configs.Get().Common.RingBufferLogSize)
	if _pc, ok := connectMap.Load(c.Id); ok {
		_pc.Close()
		_pc = nil
		connectMap.Delete(c.Id)
	}
	connectMap.Store(c.Id, &pc)
	return &pc
}

func (p *proxyConnect) scheduler() {
	p.SetStatus(types.ConnectStatusIng)
	p.sshConfig = new(ssh.ClientConfig)
	defer vclose.Close(p.client)
	if err := p.login(); err != nil {
		p.error("%v", err)
		p.SetStatus(types.ConnectStatusRe)
		p.reConnectCancel()
		return // 这里不停止，需要自动重连
	}
	cfg := configs.Get()
	// 启动心跳
	go func() {
		defer p.logs("隧道心跳停止")
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		tck := time.NewTicker(cfg.Common.HeartBeat)
		go p.quit(ctx, func() { tck.Stop(); cancel() })
		for {
			select {
			case <-tck.C:
				if err := p.heartBeat(); err != nil {
					p.error("%v", err)
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	p.SetStatus(types.ConnectStatusOk)
	p.logs("SSH连接已建立，开始下一步的转发动作")
	switch p.connect.Connect {
	case types.ForwardTypeLocal:
		p.localAndDynamic(p.forwardLocal)
	case types.ForwardTypeRemote:
		p.forwardRemote()
	case types.ForwardTypeDynamic:
		p.localAndDynamic(p.forwardDynamic)
	case types.ForwardTypeHTTP:
		p.forwardHTTP()

	}
}

func (p *proxyConnect) login() error {
	if p.client != nil {
		vclose.Close(p.client)
	}
	p.config()
	var err error
	p.client, err = ssh.Dial("tcp", p.connect.Saddr, p.sshConfig)
	if err != nil {
		p.SetStatus(types.ConnectStatusFail)
		return fmt.Errorf("目标地址连接失败 %s %v", p.connect.Saddr, err)
	}
	return nil
}

func (p *proxyConnect) config() {
	var auth ssh.AuthMethod
	if p.connect.Stype == types.SSHValidKey {
		pk, err := ssh.ParsePrivateKey([]byte(p.connect.Passwd))
		if err != nil {
			p.error("密钥文件解析失败 %v", err)
			return
		}
		auth = ssh.PublicKeys(pk)
	} else {
		auth = ssh.Password(p.connect.Passwd)
	}
	p.sshConfig.Auth = []ssh.AuthMethod{
		auth,
	}
	p.sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	p.sshConfig.Timeout = configs.Get().Common.SshTimeout
	p.sshConfig.User = p.connect.User
}

func (p *proxyConnect) heartBeat() error {
	if p.client == nil {
		return fmt.Errorf("ssh 连接已断开")
	}
	sess, err := p.client.NewSession()
	defer vclose.Close(sess)
	if err != nil {
		return fmt.Errorf("创建session失败 %v", err)
	}
	if err = sess.Run(""); err != nil {
		return fmt.Errorf("心跳失败 %v", err)
	}
	return nil
}

// 统一的退出以及回调函数
func (p *proxyConnect) quit(ctx context.Context, fs ...func()) {
	defer func() {
		if len(fs) > 0 {
			p.logs("隧道内部监听上下文停止信号，回调清理函数触发")
			for i := range fs {
				fs[i]()
			}
		}
	}()
	select {
	case <-p.parentReConnectCtx.Done():
		p.logs("隧道内部父级停止信号")
		return
	case <-p.stopCtx.Done():
		p.logs("隧道内部停止")
		return
	case <-p.reConnectCtx.Done():
		p.logs("隧道内部收到重连信号")
		return
	case <-ctx.Done():
		p.logs("隧道内部区块停止")
		p.reConnectCancel() // 发送重连信号
		return
	}
}

type proxyFunc func(conn net.Conn)

func (p *proxyConnect) localAndDynamic(f proxyFunc) {
	defer p.logs("隧道正向代理或者动态代理停止")
	p.logs("本地转发或者动态转发启动")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go p.quit(ctx, cancel)
	lc := net.ListenConfig{KeepAlive: time.Minute}
	localServer, err := lc.Listen(ctx, "tcp", p.connect.Listen)
	defer vclose.Close(localServer)
	if err != nil {
		p.error("代理端口监听失败 %v", err)
		p.SetStatus(types.ConnectStatusRe) // 监听失败重新监听
		return
	}
	p.logs("本地转发或动态转发启动成功")
	if len(p.connect.Son) > 0 {
		time.Sleep(1 * time.Second)
		p.logs("开始创建子系统")
		StartList(p.connect.Son)
	}
	for {
		if ctx.Err() != nil {
			return
		}
		// 设置超时以便定期检查上下文
		if tcpListener, ok := localServer.(*net.TCPListener); ok {
			err = tcpListener.SetDeadline(time.Now().Add(500 * time.Millisecond))
			if err != nil {
				return
			}
		}
		client, accErr := localServer.Accept()
		if accErr != nil {

			if isTimeoutError(accErr) {
				continue // 超时，继续循环
			}

			if isClosedError(accErr) {
				p.SetStatus(types.ConnectStatusRe) // 监听失败重新监听
				p.logs("监听器已关闭，停止接受连接")
				return
			}
			p.error("Accept 数据获取失败 %v", accErr)
			// 对于非超时错误，短暂等待后重试
			select {
			case <-ctx.Done():
				p.SetStatus(types.ConnectStatusRe) // 监听失败重新监听
				return
			case <-time.After(time.Second):
				continue
			}

		}
		go f(client)
	}
}

func (p *proxyConnect) Close() {
	connectMap.Delete(p.connect.Id) // 删除缓存中的数据
	vclose.Close(p.client)
	p.sshConfig = nil

	// 清理环形缓冲区
	if p.buffer != nil {
		p.buffer.Clear()
		p.buffer = nil
	}
}
