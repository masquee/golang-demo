# Kubernetes Client-Go Demo

学习 Kubernetes watch 机制的演示程序，使用 client-go 默认的 HTTP/2 协议。

## 🚀 快速开始

有2种运行方式：

### 1. 基础版本
```bash
go run main.go
```
- 演示 List 和 Watch Pods
- 使用 client-go 默认配置（HTTP/2 over HTTPS）
- 流量被 TLS 加密

### 2. HTTP 日志版本（推荐🌟）
```bash
./run-with-logging.sh
```
- **显示完整的 HTTP 请求和响应**
- 可以看到 client-go 的真实网络调用
- 使用标准的 HTTP/2 协议
- 最适合学习 Kubernetes API 和 watch 机制

## 📦 文件说明

- `main.go` - 基础演示程序
- `main-with-logging.go` - 带详细 HTTP 日志记录
- `run-with-logging.sh` - 运行日志版本的脚本
- `captures/` - 保存抓包文件的目录

## 🔍 学习重点

**List vs Watch 的区别：**
- **List**: `GET /api/v1/pods` - 一次性获取所有 Pod (HTTP/1.1)
- **Watch**: `GET /api/v1/pods?watch=true` - 长连接接收事件流 (HTTP/2.0)

**client-go 的网络行为：**
- 自动选择 HTTP 协议版本（1.1 或 2.0）
- 使用 TLS 客户端证书认证
- Watch 请求使用 HTTP/2 流式传输
- 事件类型：ADDED, MODIFIED, DELETED

## ⚡ 推荐使用

**学习 watch 机制：**
```bash
./run-with-logging.sh
```

这是真正的 client-go 默认行为，使用标准的 kubeconfig 和 HTTP/2 协议。