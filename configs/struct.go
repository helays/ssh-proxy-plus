package configs

import (
	"time"

	"helay.net/go/utils/v3/db"
	"helay.net/go/utils/v3/net/http/httpServer"
	"helay.net/go/utils/v3/net/http/httpServer/router"
	"helay.net/go/utils/v3/net/http/session"
	"helay.net/go/utils/v3/safe/cachemgr"
)

type Config struct {
	Common                `json:"common" yaml:"common"`
	httpServer.HttpServer `json:"http_server" yaml:"http_server"`
	Router                `json:"router" yaml:"router"`
	session.Options       `yaml:"session" json:"session"`
	Db                    db.Dbbase `yaml:"db"`
	SessionConfig         `yaml:"session_config" json:"session_config"`
}

type Router struct {
	router.Router    `json:"router" yaml:"router"`
	router.LoginInfo `json:"login_info" yaml:"login_info"`
}

type Common struct {
	Cache             string        `ini:"cache" json:"cache" yaml:"cache"`
	HeartBeat         time.Duration `ini:"heart_beat" json:"heart_beat" yaml:"heart_beat"`                               // 心跳检测走起
	SshTimeout        time.Duration `ini:"ssh_timeout" json:"ssh_timeout" yaml:"ssh_timeout"`                            // ssh连接超时
	EnablePass        bool          `ini:"enable_pass" json:"enable_pass" yaml:"enable_pass"`                            // 启用系统登录
	EnableAliEcs      bool          `ini:"enable_ali_ecs" json:"enable_ali_ecs" yaml:"enable_ali_ecs"`                   // 开启阿里ECS
	RingBufferLogSize int           `ini:"log_ring_buffer_size" yaml:"log_ring_buffer_size" json:"log_ring_buffer_size"` // 环形缓冲区 日志 大小
	EnablePublicProxy bool          `ini:"enable_public_proxy" json:"enable_public_proxy" yaml:"enable_public_proxy"`    // 启用公共代理
	CheckMaxThread    int           `ini:"check_max_thread" json:"check_max_thread" yaml:"check_max_thread"`             // 检测最大线程
	ConnectTimeout    time.Duration `ini:"connect_timeout" json:"connect_timeout" yaml:"connect_timeout"`                // 连接超时
	SpeedTestTimeout  time.Duration `ini:"speed_test_timeout" json:"speed_test_timeout" yaml:"speed_test_timeout"`       // 测速超时
	Socks5Port        string        `ini:"socks5_port" json:"socks5_port" yaml:"socks5_port"`                            // socks5端口
	HttpPort          string        `ini:"http_port" json:"http_port" yaml:"http_port"`                                  // http端口
}

type SessionConfig struct {
	SessionEngine   cachemgr.Driver `ini:"session_engine" yaml:"session_engine" json:"session_engine"`          // session引擎 memory file redis db
	SessionFilePath string          `ini:"session_file_path" yaml:"session_file_path" json:"session_file_path"` // session文件路径
}
