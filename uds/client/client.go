package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	socketPath := "/tmp/echo.sock"

	// 连接到服务器的 Unix 域套接字
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		fmt.Println("连接服务器失败:", err)
		return
	}
	defer conn.Close()

	// 从标准输入读取用户输入
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("请输入要发送的消息: ")
	message, _ := reader.ReadString('\n')

	// 发送消息到服务器
	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Println("发送消息失败:", err)
		return
	}

	// 接收服务器的回复
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("接收回复失败:", err)
		return
	}

	response := string(buffer[:n])
	fmt.Println("服务器回复:", response)
}
