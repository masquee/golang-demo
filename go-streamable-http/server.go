package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

// 流式 JSON 响应处理器
func streamJSONHandler(w http.ResponseWriter, r *http.Request) {
	// 设置响应头以支持流式传输
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 立即发送响应头
	w.WriteHeader(http.StatusOK)

	// 获取客户端指定的流式数据数量，默认为10
	countParam := r.URL.Query().Get("count")
	count := 10
	if countParam != "" {
		if c, err := strconv.Atoi(countParam); err == nil && c > 0 {
			count = c
		}
	}

	// 流式发送数据
	for i := 1; i <= count; i++ {
		data := StreamData{
			Timestamp: time.Now().Unix(),
			Message:   fmt.Sprintf("Stream message #%d", i),
			Count:     i,
		}

		// 将数据编码为 JSON
		jsonData, err := json.Marshal(data)
		if err != nil {
			log.Printf("Error marshaling JSON: %v", err)
			return
		}

		// 写入响应
		if _, err := w.Write(jsonData); err != nil {
			log.Printf("Error writing response: %v", err)
			return
		}

		// 添加换行符分隔每个 JSON 对象
		if _, err := w.Write([]byte("\n")); err != nil {
			log.Printf("Error writing newline: %v", err)
			return
		}

		// 强制刷新缓冲区，确保数据立即发送
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}

		// 模拟处理时间
		time.Sleep(500 * time.Millisecond)
	}
}

// 流式文本响应处理器
func streamTextHandler(w http.ResponseWriter, _ *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	w.WriteHeader(http.StatusOK)

	// 流式发送文本数据
	for i := 1; i <= 20; i++ {
		message := fmt.Sprintf("[%s] Streaming line %d - Current time: %s\n",
			time.Now().Format("15:04:05"),
			i,
			time.Now().Format("2006-01-02 15:04:05"))

		if _, err := w.Write([]byte(message)); err != nil {
			log.Printf("Error writing response: %v", err)
			return
		}

		// 刷新缓冲区
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}

		time.Sleep(300 * time.Millisecond)
	}
}

// Server-Sent Events (SSE) 处理器
func sseHandler(w http.ResponseWriter, _ *http.Request) {
	// 设置 SSE 响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	w.WriteHeader(http.StatusOK)

	// 发送初始连接消息
	if _, err := fmt.Fprintf(w, "data: Connected to SSE stream\n\n"); err != nil {
		log.Printf("Error writing SSE message: %v", err)
		return
	}
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	// 持续发送事件
	for i := 1; i <= 15; i++ {
		event := map[string]interface{}{
			"id":        i,
			"timestamp": time.Now().Unix(),
			"data":      fmt.Sprintf("SSE Event #%d", i),
			"random":    time.Now().Nanosecond() % 1000,
		}

		jsonData, _ := json.Marshal(event)

		// SSE 格式：data: {json}\n\n
		if _, err := fmt.Fprintf(w, "id: %d\n", i); err != nil {
			log.Printf("Error writing SSE id: %v", err)
			return
		}
		if _, err := fmt.Fprintf(w, "data: %s\n\n", jsonData); err != nil {
			log.Printf("Error writing SSE data: %v", err)
			return
		}

		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}

		time.Sleep(800 * time.Millisecond)
	}

	// 发送结束事件
	if _, err := fmt.Fprintf(w, "event: close\n"); err != nil {
		log.Printf("Error writing SSE close event: %v", err)
		return
	}
	if _, err := fmt.Fprintf(w, "data: Stream ended\n\n"); err != nil {
		log.Printf("Error writing SSE close data: %v", err)
		return
	}
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

// 主页处理器
func indexHandler(w http.ResponseWriter, _ *http.Request) {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>Go Streamable HTTP Demo</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .container { max-width: 800px; margin: 0 auto; }
        .endpoint { margin: 20px 0; padding: 15px; border: 1px solid #ddd; border-radius: 5px; }
        .output { background: #f5f5f5; padding: 10px; margin-top: 10px; border-radius: 3px; height: 200px; overflow-y: auto; font-family: monospace; white-space: pre-wrap; }
        button { padding: 8px 16px; margin: 5px; background: #007cba; color: white; border: none; border-radius: 3px; cursor: pointer; }
        button:hover { background: #005a87; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Go Streamable HTTP Demo</h1>
        
        <div class="endpoint">
            <h3>1. JSON Stream (/stream/json)</h3>
            <p>流式传输 JSON 数据，每个对象以换行符分隔</p>
            <button onclick="fetchJSONStream()">开始 JSON 流</button>
            <div id="json-output" class="output"></div>
        </div>

        <div class="endpoint">
            <h3>2. Text Stream (/stream/text)</h3>
            <p>流式传输纯文本数据</p>
            <button onclick="fetchTextStream()">开始文本流</button>
            <div id="text-output" class="output"></div>
        </div>

        <div class="endpoint">
            <h3>3. Server-Sent Events (/sse)</h3>
            <p>使用 SSE 协议的事件流</p>
            <button onclick="startSSE()">开始 SSE 流</button>
            <button onclick="stopSSE()">停止 SSE 流</button>
            <div id="sse-output" class="output"></div>
        </div>
    </div>

    <script>
        let eventSource = null;

        async function fetchJSONStream() {
            const output = document.getElementById('json-output');
            output.textContent = 'Starting JSON stream...\n';
            
            try {
                const response = await fetch('/stream/json?count=8');
                const reader = response.body.getReader();
                const decoder = new TextDecoder();
                
                while (true) {
                    const { done, value } = await reader.read();
                    if (done) break;
                    
                    const text = decoder.decode(value, { stream: true });
                    const lines = text.split('\n').filter(line => line.trim());
                    
                    lines.forEach(line => {
                        try {
                            const json = JSON.parse(line);
                            output.textContent += JSON.stringify(json, null, 2) + '\n';
                        } catch (e) {
                            output.textContent += line + '\n';
                        }
                    });
                    
                    output.scrollTop = output.scrollHeight;
                }
            } catch (error) {
                output.textContent += 'Error: ' + error.message + '\n';
            }
        }

        async function fetchTextStream() {
            const output = document.getElementById('text-output');
            output.textContent = 'Starting text stream...\n';
            
            try {
                const response = await fetch('/stream/text');
                const reader = response.body.getReader();
                const decoder = new TextDecoder();
                
                while (true) {
                    const { done, value } = await reader.read();
                    if (done) break;
                    
                    const text = decoder.decode(value, { stream: true });
                    output.textContent += text;
                    output.scrollTop = output.scrollHeight;
                }
            } catch (error) {
                output.textContent += 'Error: ' + error.message + '\n';
            }
        }

        function startSSE() {
            const output = document.getElementById('sse-output');
            output.textContent = 'Connecting to SSE stream...\n';
            
            if (eventSource) {
                eventSource.close();
            }
            
            eventSource = new EventSource('/sse');
            
            eventSource.onopen = function(event) {
                output.textContent += 'SSE connection opened\n';
            };
            
            eventSource.onmessage = function(event) {
                try {
                    const data = JSON.parse(event.data);
                    output.textContent += 'Event: ' + JSON.stringify(data, null, 2) + '\n';
                } catch (e) {
                    output.textContent += 'Message: ' + event.data + '\n';
                }
                output.scrollTop = output.scrollHeight;
            };
            
            eventSource.addEventListener('close', function(event) {
                output.textContent += 'Stream closed: ' + event.data + '\n';
                eventSource.close();
                eventSource = null;
            });
            
            eventSource.onerror = function(event) {
                output.textContent += 'SSE error occurred\n';
                output.scrollTop = output.scrollHeight;
            };
        }

        function stopSSE() {
            if (eventSource) {
                eventSource.close();
                eventSource = null;
                document.getElementById('sse-output').textContent += 'SSE connection closed by user\n';
            }
        }
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(html)); err != nil {
		log.Printf("Error writing HTML response: %v", err)
	}
}

func main() {
	// 注册路由
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/stream/json", streamJSONHandler)
	http.HandleFunc("/stream/text", streamTextHandler)
	http.HandleFunc("/sse", sseHandler)

	port := ":8080"
	log.Printf("Starting streamable HTTP server on http://localhost%s", port)
	log.Printf("Available endpoints:")
	log.Printf("  - http://localhost%s/           (Demo page)", port)
	log.Printf("  - http://localhost%s/stream/json (JSON stream)", port)
	log.Printf("  - http://localhost%s/stream/text (Text stream)", port)
	log.Printf("  - http://localhost%s/sse         (Server-Sent Events)", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
