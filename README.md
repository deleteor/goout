# goout

科学上网工具 一个简单的socks5协议来实现代理转发

`请勿用于非法用途!!!`

## 环境

* Golang 1.14+
* make

### 编译

```.
make local
```

### 使用

* 客户端

./build/goout client

* 服务端

./build/goout server

### 配置

* 客户端(默认读取configs/client.yaml)

local: 127.0.0.1:1081  本地 监听地址:端口
server: 1.1.1.1:443  远程地址
password: 123456  密码
encrytype: random  加密方式

* 服务端(默认读取configs/server.yaml)

local: :443
password: 123456
encrytype: random

暂无法防御GFW的主动探测
