#!/bin/bash

echo "🚀 启动 Go Streamable HTTP 服务器..."
echo "============================================"

# 检查是否存在编译后的服务器文件
if [ -f "./server" ]; then
    echo "使用已编译的服务器文件..."
    ./server
else
    echo "编译并启动服务器..."
    go run server.go types.go
fi
