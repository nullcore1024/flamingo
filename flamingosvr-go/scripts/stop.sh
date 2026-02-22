#!/bin/bash

# 停止脚本

echo "Stopping Flamingo Server..."

# 查找并停止所有服务进程
CHAT_SERVER_PID=$(pgrep -f "bin/chatserver" || echo "")
FILE_SERVER_PID=$(pgrep -f "bin/fileserver" || echo "")
IMG_SERVER_PID=$(pgrep -f "bin/imgserver" || echo "")

if [ -n "$CHAT_SERVER_PID" ]; then
    echo "Stopping ChatServer (PID: $CHAT_SERVER_PID)..."
    kill $CHAT_SERVER_PID
fi

if [ -n "$FILE_SERVER_PID" ]; then
    echo "Stopping FileServer (PID: $FILE_SERVER_PID)..."
    kill $FILE_SERVER_PID
fi

if [ -n "$IMG_SERVER_PID" ]; then
    echo "Stopping ImgServer (PID: $IMG_SERVER_PID)..."
    kill $IMG_SERVER_PID
fi

echo "All servers stopped!"
