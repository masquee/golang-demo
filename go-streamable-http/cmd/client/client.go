package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// StreamData 与服务器端保持一致的数据结构
type StreamData struct {
	Timestamp int64  `json:"timestamp"`
	Message   string `json:"message"`
	Count     int    `json:"count"`
}

// 消费 JSON 流式响应
func consumeJSONStream(url string) error {
	fmt.Printf("🔄 开始消费 JSON 流: %s\n", url)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("服务器返回错误状态: %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var data StreamData
		if err := json.Unmarshal([]byte(line), &data); err != nil {
			fmt.Printf("❌ JSON 解析错误: %v, 原始数据: %s\n", err, line)
			continue
		}

		fmt.Printf("📦 收到数据: Count=%d, Message=%s, Timestamp=%d\n",
			data.Count, data.Message, data.Timestamp)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取流时出错: %v", err)
	}

	fmt.Println("✅ JSON 流处理完成\n")
	return nil
}

// 消费文本流式响应
func consumeTextStream(url string) error {
	fmt.Printf("🔄 开始消费文本流: %s\n", url)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("服务器返回错误状态: %d", resp.StatusCode)
	}

	buffer := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			fmt.Print(string(buffer[:n]))
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("读取响应时出错: %v", err)
		}
	}

	fmt.Println("\n✅ 文本流处理完成\n")
	return nil
}

// 消费 Server-Sent Events 流
func consumeSSEStream(url string) error {
	fmt.Printf("🔄 开始消费 SSE 流: %s\n", url)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("服务器返回错误状态: %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	var eventID string
	var eventType string
	var eventData string

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			// 空行表示一个事件结束
			if eventData != "" {
				fmt.Printf("📡 SSE事件 [ID:%s, Type:%s]: %s\n", eventID, eventType, eventData)

				// 尝试解析 JSON 数据
				if eventData != "Connected to SSE stream" && eventData != "Stream ended" {
					var jsonData map[string]interface{}
					if err := json.Unmarshal([]byte(eventData), &jsonData); err == nil {
						fmt.Printf("   解析后的数据: %+v\n", jsonData)
					}
				}
			}

			// 重置事件字段
			eventID = ""
			eventType = ""
			eventData = ""
			continue
		}

		if len(line) > 5 {
			prefix := line[:5]
			value := line[5:]

			switch prefix {
			case "id: ":
				eventID = value
			case "event":
				if len(line) > 7 && line[:7] == "event: " {
					eventType = line[7:]
				}
			case "data:":
				if len(line) > 6 && line[:6] == "data: " {
					eventData = line[6:]
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取 SSE 流时出错: %v", err)
	}

	fmt.Println("✅ SSE 流处理完成\n")
	return nil
}

// 演示逐字节读取流式响应
func consumeStreamByteByByte(url string) error {
	fmt.Printf("🔄 开始逐字节消费流: %s\n", url)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("服务器返回错误状态: %d", resp.StatusCode)
	}

	buffer := make([]byte, 1)
	var receivedData []byte

	fmt.Print("📥 实时接收数据: ")
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			receivedData = append(receivedData, buffer[0])
			fmt.Print(string(buffer[0]))

			// 每收到一行数据就处理一次
			if buffer[0] == '\n' {
				fmt.Print("\n📥 继续接收: ")
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("读取响应时出错: %v", err)
		}
	}

	fmt.Printf("\n✅ 总共接收了 %d 字节数据\n\n", len(receivedData))
	return nil
}

func main() {
	baseURL := "http://localhost:8080"

	fmt.Println("🚀 Go Streamable HTTP Client Demo")
	fmt.Println("=====================================")
	fmt.Println("确保服务器正在运行: go run server.go")
	fmt.Println()

	// 等待用户确认服务器已启动
	fmt.Print("按 Enter 键开始测试客户端...")
	fmt.Scanln()

	// 测试不同类型的流式响应
	tests := []struct {
		name string
		fn   func(string) error
		url  string
	}{
		{
			name: "JSON 流式响应",
			fn:   consumeJSONStream,
			url:  baseURL + "/stream/json?count=5",
		},
		{
			name: "文本流式响应",
			fn:   consumeTextStream,
			url:  baseURL + "/stream/text",
		},
		{
			name: "Server-Sent Events",
			fn:   consumeSSEStream,
			url:  baseURL + "/sse",
		},
		{
			name: "逐字节读取流",
			fn:   consumeStreamByteByByte,
			url:  baseURL + "/stream/json?count=3",
		},
	}

	for i, test := range tests {
		fmt.Printf("\n%d. 测试 %s\n", i+1, test.name)
		fmt.Println(strings.Repeat("-", 50))

		if err := test.fn(test.url); err != nil {
			fmt.Printf("❌ 错误: %v\n", err)
		}

		// 在测试之间添加延迟
		if i < len(tests)-1 {
			fmt.Print("等待 2 秒后继续下一个测试...")
			time.Sleep(2 * time.Second)
			fmt.Println(" 继续")
		}
	}

	fmt.Println("\n🎉 所有测试完成!")
}
