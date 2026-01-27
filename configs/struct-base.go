package configs

import (
	"time"

	"github.com/helays/utils/v2/db"
	"github.com/helays/utils/v2/net/http/httpServer"
	"github.com/helays/utils/v2/net/http/httpServer/router"
	"github.com/helays/utils/v2/net/http/session"
	"github.com/helays/utils/v2/safe/cachemgr"
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
}

type SessionConfig struct {
	SessionEngine   cachemgr.Driver `ini:"session_engine" yaml:"session_engine" json:"session_engine"`          // session引擎 memory file redis db
	SessionFilePath string          `ini:"session_file_path" yaml:"session_file_path" json:"session_file_path"` // session文件路径
}
