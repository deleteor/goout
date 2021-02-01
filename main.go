package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/keep-fool/goout/cmd"
)

func main() {
	cmd.Execute()
}

func main1() {
	server, err := net.Listen("tcp", ":1081")
	if err != nil {
		fmt.Printf("Listen failed: %v\n", err)
		return
	}

	for {
		client, err := server.Accept()
		if err != nil {
			fmt.Printf("Accept failed: %v", err)
			continue
		}
		go process(client)
	}
}

// func process(client net.Conn) {
// 	remoteAddr := client.RemoteAddr().String()
// 	fmt.Printf("Connection from %s\n", remoteAddr)
// 	client.Write([]byte("Hello world!\n"))
// 	client.Close()
// }

func process(client net.Conn) {
	if err := Socks5Auth1(client); err != nil {
		fmt.Println("auth error:", err)
		client.Close()
		return
	}
	target, err := Socks5Connect1(client)
	if err != nil {
		fmt.Println("connect error:", err)
		client.Close()
		return
	}
	Socks5Forward1(client, target)
}

/*Socks5Auth1 认证
VER 本次请求的协议版本号，取固定值 0x05（表示socks 5）
NMETHODS 客户端支持的认证方式数量，可取值 1~255
METHODS 可用的认证方式列表
*/
func Socks5Auth1(client net.Conn) (err error) {
	buf := make([]byte, 256)
	// 读取 VER 和 NMETHODS
	n, err := io.ReadFull(client, buf[:2])
	// fmt.Println("buf =", buf[:2])
	if n != 2 {
		return errors.New("reading header: " + err.Error())
	}

	ver, nMethods := int(buf[0]), int(buf[1])
	// fmt.Printf("ver = %d, nMethods = %d\n", ver, nMethods)
	if ver != 5 {
		return errors.New("invalid version")
	}

	// 读取 METHODS 列表
	n, err = io.ReadFull(client, buf[:nMethods])
	// fmt.Println("buf =", buf[:nMethods])
	if n != nMethods {
		return errors.New("reading methods: " + err.Error())
	}

	// SOCKS5协议提供5种认证方式：
	// 0x00 不需要认证
	// 0x01 GSSAPI
	// 0x02 用户名、密码认证
	// 0x03 - 0x7F由IANA分配（保留）
	// 0x80 - 0xFE为私人方法保留
	n, err = client.Write([]byte{0x05, 0x00})
	if n != 2 || err != nil {
		return errors.New("write rsp err: " + err.Error())
	}
	return nil
}

/*Socks5Connect1 代理
VER 0x05 Socks5
CMD 连接方式，0x01=CONNECT, 0x02=BIND, 0x03=UDP ASSOCIATE
RSV 保留字段
ATYP 地址类型，0x01=IPv4，0x03=域名，0x04=IPv6
DST.ADDR 目标地址
DST.PORT 目标端口，2字节，网络字节序（network octec order）
*/
func Socks5Connect1(client net.Conn) (net.Conn, error) {
	buf := make([]byte, 256)

	n, err := io.ReadFull(client, buf[:4])
	if n != 4 {
		return nil, errors.New("read header: " + err.Error())
	}

	ver, cmd, _, atyp := buf[0], buf[1], buf[2], buf[3]
	if ver != 5 || cmd != 1 {
		return nil, errors.New("invalid ver/cmd")
	}

	addr := ""
	switch atyp {
	case 1:
		n, err = io.ReadFull(client, buf[:4])
		if n != 4 {
			return nil, errors.New("invalid IPv4: " + err.Error())
		}
		addr = fmt.Sprintf("%d.%d.%d.%d", buf[0], buf[1], buf[2], buf[3])
	case 3:
		n, err = io.ReadFull(client, buf[:1])
		if n != 1 {
			return nil, errors.New("invalid hostname: " + err.Error())
		}
		addrLen := int(buf[0])

		n, err = io.ReadFull(client, buf[:addrLen])
		if n != addrLen {
			return nil, errors.New("invalid hostname: " + err.Error())
		}
		addr = string(buf[:addrLen])
	case 4:
		return nil, errors.New("IPv6: no supported yet")
	default:
		return nil, errors.New("invalid atyp")
	}
	fmt.Println("addr:", addr)
	n, err = io.ReadFull(client, buf[:2])
	if n != 2 {
		return nil, errors.New("read port: " + err.Error())
	}

	port := binary.BigEndian.Uint16(buf[:2])
	destAddrPort := fmt.Sprintf("%s:%d", addr, port)
	dest, err := net.Dial("tcp", destAddrPort)
	if err != nil {
		return nil, errors.New("dial dst: " + err.Error())
	}
	n, err = client.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
	if err != nil {
		dest.Close()
		return nil, errors.New("write rsp: " + err.Error())
	}
	return dest, nil
}

// Socks5Forward1 转发
func Socks5Forward1(client, target net.Conn) {
	forward := func(src, dest net.Conn) {
		defer src.Close()
		defer dest.Close()
		io.Copy(src, dest)
	}
	go forward(client, target)
	go forward(target, client)
}
