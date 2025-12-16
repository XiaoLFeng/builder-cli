#!/bin/bash
# xbuilder 通知脚本示例
# 用法: 在 hooks.on_failure 中配置

# 钉钉/企业微信通知示例
# WEBHOOK_URL="https://oapi.dingtalk.com/robot/send?access_token=xxx"

echo "构建失败通知"
echo "时间: $(date)"
echo "项目: ${PROJECT_NAME:-unknown}"

# curl -X POST "$WEBHOOK_URL" \
#   -H "Content-Type: application/json" \
#   -d '{
#     "msgtype": "text",
#     "text": {
#       "content": "构建失败: '"${PROJECT_NAME}"'"
#     }
#   }'
