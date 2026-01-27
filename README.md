# ssh-proxy-plus

`ssh-proxy-plus`利用ssh的加密隧道能力，创建正向、反向、Socks以及HTTP代理。
当前版本提供前端配置界面，采用Vue3+Element Plus框架实现。

### 特性
- 支持多层级代理
- 支持正向代理
- 支持反向代理
- 支持Socks5代理
- 支持HTTP代理
- 支持快速创建阿里ECS并建立隧道
- 采用SQLite存储
- 前端静态文件编译到二进制包，简化部署流程

v2版本优化了代理链的重连与关断。

### 源码编译
``` shell
git clone https://github.com/helays/ssh-proxy-plus
make build
```