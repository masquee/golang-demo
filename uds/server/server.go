package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	socketPath := "/tmp/echo.sock"

	// 如果套接字文件已存在，先删除
	if _, err := os.Stat(socketPath); err == nil {
		os.Remove(socketPath)
	}

	// 创建 Unix 域套接字监听器
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		fmt.Println("监听套接字失败:", err)
		return
	}
	defer listener.Close()
	fmt.Println("服务器已启动，监听", socketPath)

	for {
		// 接受客户端连接
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("接受连接失败:", err)
			continue
		}

		// 处理客户端连接
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	buffer := make([]byte, 1024)

	// 读取客户端发送的数据
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("读取数据失败:", err)
		return
	}

	received := string(buffer[:n])
	fmt.Println("收到客户端消息:", received)

	// 回复客户端
	response := "服务器收到: " + received
	_, err = conn.Write([]byte(response))
	if err != nil {
		fmt.Println("发送回复失败:", err)
		return
	}
}
