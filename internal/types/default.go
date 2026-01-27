package types

type TextStatus string

const (
	TextStatusEnable  TextStatus = "Y"
	TextStatusDisable TextStatus = "N"
)

type SSHValidType int

func (s SSHValidType) String() string {
	switch s {
	case SSHValidPasswd:
		return "密码验证"
	case SSHValidKey:
		return "密钥验证"
	default:
		return "未知"
	}
}

const (
	SSHValidPasswd SSHValidType = 1
	SSHValidKey    SSHValidType = 2
)

type ConnectStatus int

func (c ConnectStatus) String() string {
	switch c {
	case ConnectStatusFail:
		return "连接失败"
	case ConnectStatusNew:
		return "新增"
	case ConnectStatusIng:
		return "连接中"
	case ConnectStatusOk:
		return "连接成功"
	case ConnectStatusStop:
		return "中断连接"
	case ConnectStatusRe:
		return "重新连接"
	case ConnectStatusDel:
		return "删除"
	default:
		return "未知"
	}
}

// 用于存储 每个连接的状态，1、新增 2、连接中 0、连接失败 3、连接成功 4、中断连接 5 重新连接
const (
	ConnectStatusFail ConnectStatus = 0
	ConnectStatusNew  ConnectStatus = 1
	ConnectStatusIng  ConnectStatus = 2
	ConnectStatusOk   ConnectStatus = 3
	ConnectStatusStop ConnectStatus = 4
	ConnectStatusRe   ConnectStatus = 5
	ConnectStatusDel  ConnectStatus = 6
)

type ForwardType string

func (f ForwardType) String() string {
	switch f {
	case ForwardTypeLocal:
		return "本地"
	case ForwardTypeRemote:
		return "远程"
	case ForwardTypeDynamic:
		return "动态"
	case ForwardTypeHTTP:
		return "HTTP"
	default:
		return "未知"
	}
}

const (
	ForwardTypeLocal   ForwardType = "L"
	ForwardTypeRemote  ForwardType = "R"
	ForwardTypeDynamic ForwardType = "D"
	ForwardTypeHTTP    ForwardType = "H"
)
