#!/bin/bash

# 启动脚本

set -e

echo "Starting Flamingo Server..."

# 进入项目目录
cd "$(dirname "$0")/.."

# 创建必要的目录
mkdir -p bin
mkdir -p files
mkdir -p images
mkdir -p configs

# 检查二进制文件是否存在
if [ ! -f "bin/chatserver" ] || [ ! -f "bin/fileserver" ] || [ ! -f "bin/imgserver" ]; then
    echo "Binaries not found. Running build script..."
    ./scripts/build.sh
fi

# 启动聊天服务
echo "Starting ChatServer..."
./bin/chatserver -ip 0.0.0.0 -port 8000 -name ChatServer &
CHAT_SERVER_PID=$!

echo "ChatServer started with PID: $CHAT_SERVER_PID"

# 启动文件服务
echo "Starting FileServer..."
./bin/fileserver -ip 0.0.0.0 -port 8001 -name FileServer -fileRoot ./files &
FILE_SERVER_PID=$!

echo "FileServer started with PID: $FILE_SERVER_PID"

# 启动图片服务
echo "Starting ImgServer..."
./bin/imgserver -ip 0.0.0.0 -port 8002 -name ImgServer -imgRoot ./images &
IMG_SERVER_PID=$!

echo "ImgServer started with PID: $IMG_SERVER_PID"

echo "All servers started successfully!"
echo "Press Ctrl+C to stop all servers."

# 等待信号
wait $CHAT_SERVER_PID $FILE_SERVER_PID $IMG_SERVER_PID
