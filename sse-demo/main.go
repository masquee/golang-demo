package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/sse", SSEHandler)
	log.Println("Server started on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// SSEHandler 处理 SSE 请求
func SSEHandler(w http.ResponseWriter, r *http.Request) {
	// 设置响应头，表明这是一个 SSE 响应
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// 允许跨域请求
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 确保响应写入不会被缓冲
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// 模拟每 2 秒发送一次事件
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			message := "data: This is an SSE message\n\n" // 构造 SSE 消息
			_, err := w.Write([]byte(message))            // 写入消息到响应
			if err != nil {
				log.Printf("Error writing SSE message: %v", err)
				return
			}
			flusher.Flush() // 刷新响应，确保消息立即发送
		case <-r.Context().Done():
			// 客户端断开连接，退出循环
			log.Println("Client disconnected")
			return
		}
	}
}
