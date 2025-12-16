#!/bin/bash
# xbuilder 构建脚本示例
# 用法: 在 xbuilder.yaml 中配置 script: "./scripts/build.sh"

set -e

echo "================================================"
echo "  开始 Maven 构建"
echo "================================================"

# 设置 Maven 选项
export MAVEN_OPTS="-Xmx1024m"

# 执行构建
mvn clean package -DskipTests -P prod

echo "================================================"
echo "  构建完成!"
echo "================================================"

# 列出生成的 jar 文件
echo "生成的文件:"
find . -name "*.jar" -path "*/target/*" -type f 2>/dev/null | head -20
