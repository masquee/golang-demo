#!/bin/bash

echo "🚀 启动 Go Streamable HTTP 客户端..."
echo "====================================="

# 检查是否存在编译后的客户端文件
if [ -f "./client" ]; then
    echo "使用已编译的客户端文件..."
    ./client
else
    echo "编译并启动客户端..."
    go run client.go
fi
