#!/bin/bash

set -e  # 遇到错误立即退出

# 获取脚本所在目录（server 目录）
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# 检查 swag 是否安装，如果未安装则跳过
if command -v swag &> /dev/null; then
    echo "正在生成 Swagger 文档..."
    swag init
else
    echo "警告: swag 命令未找到，跳过 Swagger 文档生成"
    echo "如需生成 Swagger 文档，请安装: go install github.com/swaggo/swag/cmd/swag@latest"
fi

# 定义构建参数
MAIN_FILE="./main.go"
BIN_NAME="box-server"
BIN_DIR="$SCRIPT_DIR"

# 创建 bin 目录（如果不存在）
mkdir -p "$BIN_DIR"

echo "构建目录: $SCRIPT_DIR"
echo "输出目录: $BIN_DIR"
echo "开始构建 Linux amd64 二进制文件..."

# 构建二进制文件
env GOOS=linux GOARCH=amd64 go build -v -o "$BIN_DIR/$BIN_NAME" "$MAIN_FILE"

if [ $? -eq 0 ]; then
    echo "✓ 构建成功: $BIN_DIR/$BIN_NAME"
    
    # 显示文件信息
    ls -lh "$BIN_DIR/$BIN_NAME"
    
    # 可选：部署到服务器（取消注释以启用）
    #echo "正在部署到服务器..."
    #scp "$BIN_DIR/$BIN_NAME" target_ubuntu:/tmp
    #echo "部署完成"
else
    echo "✗ 构建失败"
    exit 1
fi

