package cmp_proxy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/helays/ssh-proxy-plus/internal/cache"
	"github.com/helays/ssh-proxy-plus/internal/model"

	"golang.org/x/crypto/ssh"
	"helay.net/go/utils/v3/close/vclose"
	"helay.net/go/utils/v3/logger/ulogs"
)

func checkPortAndGetPID(client *ssh.Client, port int) (int, error) {
	session, err := client.NewSession()
	if err != nil {
		return 0, fmt.Errorf("failed to create session: %v", err)
	}
	defer vclose.Close(session)

	// 先在Go中计算十六进制端口
	portHex := fmt.Sprintf("%04X", port)
	portHexLower := strings.ToLower(portHex)

	cmd := fmt.Sprintf(`
        # 同时检查IPv4和IPv6
        for file in /proc/net/tcp /proc/net/tcp6; do
            [ -f "$file" ] || continue
            # 查找本地端口(第二列的后4位是端口号)
            awk -v port=":%s" '
                function hex2dec(str) {
                    return sprintf("%%d", "0x" str)
                }
                {
                    split($2, parts, ":");
                    local_port_hex = parts[2];
                    if (local_port_hex == port) {
                        print $10;  # 输出inode
                    }
                }
            ' "$file"
        done | 
        # 去除重复inode
        sort -u | 
        while read inode; do
            # 查找使用该inode的进程
            find /proc/[0-9]*/fd/ -lname "socket:\[$inode\]" 2>/dev/null |
            awk -F/ '{print $3}' |
            sort -u
        done`, portHexLower)

	var stdout bytes.Buffer
	session.Stdout = &stdout
	err = session.Run(cmd)

	if err != nil && !strings.Contains(err.Error(), "exit status 1") {
		return 0, fmt.Errorf("failed to check /proc: %v", err)
	}

	output := strings.TrimSpace(stdout.String())
	if output == "" {
		return 0, nil
	}

	// 可能有多个PID，取第一个
	pids := strings.Split(output, "\n")
	pid, err := strconv.Atoi(pids[0])
	if err != nil {
		return 0, fmt.Errorf("failed to parse PID: %v", err)
	}

	return pid, nil
}

// 终止进程
func killProcess(client *ssh.Client, pid int) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}
	defer vclose.Close(session)

	// 执行kill命令
	cmd := fmt.Sprintf("kill -9 %d", pid)
	if err = session.Run(cmd); err != nil {
		return fmt.Errorf("failed to kill process: %v", err)
	}
	return nil
}

// 这是双向数据转发，对于其中一方Copy结束时，另一方都应该关闭资源。
func transfer(dst net.Conn, src net.Conn) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		if _, err := io.Copy(dst, src); err != nil {
			ulogs.Errorf("【transfer】 src = > dst %v", err)
		}
		cancel()
	}()
	go func() {
		if _, err := io.Copy(src, dst); err != nil {
			ulogs.Errorf("【transfer】 dst = > src %v", err)
		}
		cancel()
	}()
	<-ctx.Done()
	time.Sleep(100 * time.Millisecond) // 短暂等待100 ms

}

func FindConnectByID(id string) (result *model.Connect) {
	cache.ConnectList.ReadWith(func(connects []model.Connect) {
		stack := make([]*model.Connect, 0, len(connects))
		for i := range connects {
			stack = append(stack, &connects[i])
		}
		// 开始DFS搜索
		for len(stack) > 0 {
			// 弹出栈顶元素
			current := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			// 检查当前元素是否匹配
			if current.Id == id {
				result = current
				break
			}
			for i := range current.Son {
				stack = append(stack, &current.Son[i])
			}
		}
	})
	return
}

func FindConnectStatus() map[string]interface{} {
	var statusMap = make(map[string]interface{})
	connectMap.Range(func(key string, value *proxyConnect) bool {
		statusMap[key] = value.GetStatus()
		return true
	})
	return statusMap
}

// 辅助函数：检查是否为超时错误
func isTimeoutError(err error) bool {
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return opErr.Timeout()
	}
	return false
}

// 辅助函数：检查是否为连接关闭错误
func isClosedError(err error) bool {
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return opErr.Err.Error() == "use of closed network connection"
	}
	return strings.Contains(err.Error(), "use of closed network connection")
}
