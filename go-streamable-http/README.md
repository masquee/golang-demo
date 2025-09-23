# Go Streamable HTTP Demo

这个项目展示了如何在 Go 中实现和使用流式 HTTP 响应，包括三种不同的流式传输方式：

1. **JSON 流式响应** - 逐个发送 JSON 对象
2. **文本流式响应** - 实时发送文本数据  
3. **Server-Sent Events (SSE)** - 使用 SSE 协议的事件流

## 项目结构

```
go-streamable-http/
├── go.mod              # 主模块文件
├── server.go           # HTTP 服务器，提供流式响应
├── types.go            # 共享的数据类型定义
├── server              # 编译后的服务器可执行文件
├── start-server.sh     # 服务器启动脚本
├── test-endpoints.sh   # API 端点测试脚本
├── cmd/
│   └── client/
│       ├── go.mod      # 客户端模块文件
│       ├── client.go   # Go 客户端，演示如何消费流式响应
│       ├── client      # 编译后的客户端可执行文件
│       └── run-client.sh # 客户端运行脚本
└── README.md           # 项目说明文档
```

## 功能特性

### 服务器端 (server.go)
- **流式 JSON 响应** (`/stream/json`): 发送带时间戳的 JSON 数据流
- **流式文本响应** (`/stream/text`): 发送实时文本数据
- **Server-Sent Events** (`/sse`): 使用 SSE 协议发送事件流
- **Web 界面** (`/`): 提供交互式的网页演示界面

### 客户端 (cmd/client/client.go)
- **JSON 流消费**: 逐行解析 JSON 数据流
- **文本流消费**: 实时读取文本数据
- **SSE 流消费**: 解析 Server-Sent Events 格式
- **逐字节读取**: 演示低级别的流式数据处理

## 快速开始

### 1. 启动服务器

```bash
# 方式一：直接运行源代码
go run server.go types.go

# 方式二：编译后运行
go build -o server server.go types.go
./server

# 方式三：使用启动脚本
chmod +x start-server.sh
./start-server.sh
```

服务器将在 `http://localhost:8080` 启动。

### 2. 测试方式

#### 方式一：使用网页界面（推荐）
打开浏览器访问 http://localhost:8080，你可以：
- 点击按钮测试不同类型的流式响应
- 实时看到数据的接收过程
- 交互式地启动和停止数据流

#### 方式二：使用 Go 客户端
在新终端中运行：

```bash
# 进入客户端目录
cd cmd/client

# 直接运行源代码
go run client.go

# 或编译后运行
go build -o client client.go
./client

# 或使用运行脚本
chmod +x run-client.sh
./run-client.sh
```

#### 方式三：使用 curl 命令行测试

```bash
# 使用测试脚本
chmod +x test-endpoints.sh
./test-endpoints.sh

# 或手动测试各个端点
curl http://localhost:8080/stream/json?count=5
curl http://localhost:8080/stream/text
curl http://localhost:8080/sse
```

## API 端点详情

### GET /stream/json
**描述**: 流式发送 JSON 数据
**参数**: 
- `count` (可选): 指定发送的数据条数，默认为 10

**响应格式**: 每行一个 JSON 对象
```json
{"timestamp":1695456789,"message":"Stream message #1","count":1}
{"timestamp":1695456790,"message":"Stream message #2","count":2}
```

### GET /stream/text
**描述**: 流式发送文本数据
**响应格式**: 纯文本，实时发送
```
[14:23:45] Streaming line 1 - Current time: 2024-09-23 14:23:45
[14:23:46] Streaming line 2 - Current time: 2024-09-23 14:23:46
```

### GET /sse
**描述**: 使用 Server-Sent Events 协议发送事件流
**响应格式**: SSE 标准格式
```
data: Connected to SSE stream

id: 1
data: {"id":1,"timestamp":1695456789,"data":"SSE Event #1","random":123}

event: close
data: Stream ended
```

## 技术要点

### 流式响应的关键实现
1. **设置正确的 HTTP 头**:
   ```go
   w.Header().Set("Content-Type", "application/json")
   w.Header().Set("Cache-Control", "no-cache")
   w.Header().Set("Connection", "keep-alive")
   ```

2. **立即发送响应头**:
   ```go
   w.WriteHeader(http.StatusOK)
   ```

3. **强制刷新缓冲区**:
   ```go
   if flusher, ok := w.(http.Flusher); ok {
       flusher.Flush()
   }
   ```

### 客户端流式消费的关键技术
1. **使用 bufio.Scanner** 逐行读取
2. **使用 io.Reader** 逐块读取
3. **使用 EventSource API** (浏览器端) 处理 SSE

## 使用场景

- **实时日志流**: 实时查看应用程序日志
- **进度更新**: 长时间运行任务的进度通知
- **实时数据**: 股票价格、传感器数据等实时信息
- **聊天应用**: 实时消息推送
- **文件下载**: 大文件的流式下载

## 注意事项

1. **内存管理**: 流式响应不会一次性加载所有数据到内存
2. **连接管理**: 长连接需要适当的超时和错误处理
3. **缓冲控制**: 及时刷新缓冲区确保数据实时传输
4. **错误处理**: 网络中断时的优雅处理

## 运行示例

1. **启动服务器**:
   ```bash
   ./start-server.sh
   ```

2. **在另一个终端测试 API**:
   ```bash
   ./test-endpoints.sh
   ```

3. **运行 Go 客户端**:
   ```bash
   cd cmd/client && ./run-client.sh
   ```

## 扩展功能

你可以基于这个 demo 扩展以下功能：
- 添加身份验证
- 实现背压控制
- 添加数据压缩
- 支持 WebSocket 协议
- 添加监控和指标收集

## 许可证

MIT License - 随意使用和修改此演示代码。
