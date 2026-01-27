package cmp_proxy

import "github.com/helays/ssh-proxy-plus/internal/types"

func (p *proxyConnect) SetStatus(s types.ConnectStatus) {
	p.statusLock.Lock()
	defer p.statusLock.Unlock()
	p.status = s
}
func (p *proxyConnect) GetStatus() types.ConnectStatus {
	p.statusLock.RLock()
	defer p.statusLock.RUnlock()
	return p.status
}
