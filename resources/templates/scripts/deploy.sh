#!/bin/bash
# xbuilder 部署脚本示例
# 用法: 在 xbuilder.yaml 中配置 local_script: "./scripts/deploy.sh"

set -e

echo "================================================"
echo "  开始部署"
echo "================================================"

# 进入部署目录
cd /opt/services

# 拉取最新镜像
echo "拉取最新镜像..."
docker-compose pull

# 停止旧服务
echo "停止旧服务..."
docker-compose down

# 启动新服务
echo "启动新服务..."
docker-compose up -d

# 清理旧镜像
echo "清理旧镜像..."
docker system prune -f

# 检查服务状态
echo "服务状态:"
docker-compose ps

echo "================================================"
echo "  部署完成!"
echo "================================================"
