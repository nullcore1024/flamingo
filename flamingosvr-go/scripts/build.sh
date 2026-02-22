#!/bin/bash

# 构建脚本

set -e

echo "Building Flamingo Server..."

# 进入项目目录
cd "$(dirname "$0")/.."

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed"
    exit 1
fi

# 安装依赖
echo "Installing dependencies..."
go mod download
go mod tidy

# 构建聊天服务
echo "Building ChatServer..."
go build -o bin/chatserver ./cmd/chatserver

# 构建文件服务
echo "Building FileServer..."
go build -o bin/fileserver ./cmd/fileserver

# 构建图片服务
echo "Building ImgServer..."
go build -o bin/imgserver ./cmd/imgserver

echo "Build completed successfully!"
echo "Binaries are available in the bin directory."
