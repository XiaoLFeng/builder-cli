package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// 完整配置文件模板
const fullConfigTemplate = `# xbuilder 完整配置示例
# 文档: https://github.com/xiaolfeng/builder-cli

version: "1.0"

# ─────────────────────────────────────────────────────────────
# 项目基本信息
# ─────────────────────────────────────────────────────────────
project:
  name: "my-microservices"
  description: "微服务项目构建配置"

# ─────────────────────────────────────────────────────────────
# 全局变量 (可在配置中使用 ${VAR_NAME} 引用)
# ─────────────────────────────────────────────────────────────
variables:
  APP_VERSION: "1.0.0"
  REGISTRY_PREFIX: "registry.example.com/myproject"
  DEPLOY_ENV: "production"

# ─────────────────────────────────────────────────────────────
# Docker Registry 配置
# ─────────────────────────────────────────────────────────────
registries:
  default:
    url: "registry.example.com"
    username: "${DOCKER_USERNAME}"      # 从环境变量读取
    password: "${DOCKER_PASSWORD}"

  aliyun:
    url: "registry.cn-hangzhou.aliyuncs.com"
    username: "${ALIYUN_USERNAME}"
    password: "${ALIYUN_PASSWORD}"

# ─────────────────────────────────────────────────────────────
# SSH 服务器配置
# ─────────────────────────────────────────────────────────────
servers:
  production:
    host: "192.168.1.100"
    port: 22
    username: "deploy"
    auth:
      type: "key"                       # "password" | "key"
      key_path: "~/.ssh/id_rsa"

  staging:
    host: "192.168.1.101"
    port: 22
    username: "deploy"
    auth:
      type: "password"
      password: "${SSH_PASSWORD}"

# ─────────────────────────────────────────────────────────────
# 构建流水线
# ─────────────────────────────────────────────────────────────
pipeline:
  # ─────────────────────────────────────────────────────────
  # 阶段 1: Maven 构建
  # ─────────────────────────────────────────────────────────
  - stage: "maven-build"
    name: "Maven 构建"
    tasks:
      - name: "编译打包"
        type: "maven"
        config:
          # 方式1: 直接命令
          command: "mvn clean package -DskipTests -P prod"
          # 方式2: 执行脚本
          # script: "./scripts/build.sh"
          working_dir: "."
          timeout: 600                  # 超时时间 (秒)

  # ─────────────────────────────────────────────────────────
  # 阶段 2: Docker 构建 (支持并行)
  # ─────────────────────────────────────────────────────────
  - stage: "docker-build"
    name: "Docker 镜像构建"
    parallel: true                      # 启用并行执行
    tasks:
      - name: "用户服务镜像"
        type: "docker-build"
        config:
          dockerfile: "./user-service/Dockerfile"
          context: "./user-service"
          image_name: "${REGISTRY_PREFIX}/user-service"
          tag: "${APP_VERSION}"
          build_args:
            JAR_FILE: "target/*.jar"
            BUILD_ENV: "${DEPLOY_ENV}"

      - name: "订单服务镜像"
        type: "docker-build"
        config:
          dockerfile: "./order-service/Dockerfile"
          context: "./order-service"
          image_name: "${REGISTRY_PREFIX}/order-service"
          tag: "${APP_VERSION}"

      - name: "网关服务镜像"
        type: "docker-build"
        config:
          dockerfile: "./gateway-service/Dockerfile"
          context: "./gateway-service"
          image_name: "${REGISTRY_PREFIX}/gateway-service"
          tag: "${APP_VERSION}"

  # ─────────────────────────────────────────────────────────
  # 阶段 3: Docker 推送
  # ─────────────────────────────────────────────────────────
  - stage: "docker-push"
    name: "推送镜像到 Registry"
    tasks:
      - name: "推送所有镜像"
        type: "docker-push"
        config:
          registry: "default"
          # 使用 auto 自动推送上一阶段构建的镜像
          auto: true
          # 或者手动指定镜像列表
          # images:
          #   - "${REGISTRY_PREFIX}/user-service:${APP_VERSION}"
          #   - "${REGISTRY_PREFIX}/order-service:${APP_VERSION}"
          #   - "${REGISTRY_PREFIX}/gateway-service:${APP_VERSION}"

  # ─────────────────────────────────────────────────────────
  # 阶段 4: 部署
  # ─────────────────────────────────────────────────────────
  - stage: "deploy"
    name: "部署到服务器"
    tasks:
      - name: "部署到生产环境"
        type: "ssh"
        config:
          server: "production"
          # 方式1: 内联命令
          commands:
            - "cd /opt/services"
            - "docker-compose pull"
            - "docker-compose up -d"
            - "docker system prune -f"
          # 方式2: 执行本地脚本 (会上传到服务器执行)
          # local_script: "./scripts/deploy.sh"
          timeout: 300

# ─────────────────────────────────────────────────────────────
# 钩子 (可选)
# ─────────────────────────────────────────────────────────────
hooks:
  pre_build:
    - "echo '🚀 开始构建...'"
    - "date"
  post_build:
    - "echo '✅ 构建完成!'"
    - "date"
  on_failure:
    - "echo '❌ 构建失败!'"
    # - "./scripts/notify-failure.sh"
`

// 构建脚本模板
const buildScriptTemplate = `#!/bin/bash
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
`

// 部署脚本模板
const deployScriptTemplate = `#!/bin/bash
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
`

// 通知脚本模板
const notifyScriptTemplate = `#!/bin/bash
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
`

// Dockerfile 模板
const dockerfileTemplate = `# xbuilder Dockerfile 示例
# 适用于 Spring Boot 应用

FROM openjdk:17-jdk-slim

LABEL maintainer="your-email@example.com"

# 设置工作目录
WORKDIR /app

# 设置时区
ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# 复制 jar 文件
ARG JAR_FILE=target/*.jar
COPY ${JAR_FILE} app.jar

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=60s --retries=3 \
  CMD curl -f http://localhost:8080/actuator/health || exit 1

# 启动命令
ENTRYPOINT ["java", "-jar", "app.jar"]
`

// docker-compose 模板
const dockerComposeTemplate = `# xbuilder docker-compose 示例
version: '3.8'

services:
  user-service:
    image: ${REGISTRY_PREFIX}/user-service:${APP_VERSION:-latest}
    container_name: user-service
    restart: always
    ports:
      - "8081:8080"
    environment:
      - SPRING_PROFILES_ACTIVE=prod
      - JAVA_OPTS=-Xmx512m
    networks:
      - app-network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/actuator/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  order-service:
    image: ${REGISTRY_PREFIX}/order-service:${APP_VERSION:-latest}
    container_name: order-service
    restart: always
    ports:
      - "8082:8080"
    environment:
      - SPRING_PROFILES_ACTIVE=prod
      - JAVA_OPTS=-Xmx512m
    networks:
      - app-network
    depends_on:
      - user-service

  gateway-service:
    image: ${REGISTRY_PREFIX}/gateway-service:${APP_VERSION:-latest}
    container_name: gateway-service
    restart: always
    ports:
      - "8080:8080"
    environment:
      - SPRING_PROFILES_ACTIVE=prod
    networks:
      - app-network
    depends_on:
      - user-service
      - order-service

networks:
  app-network:
    driver: bridge
`

var (
	genForce      bool
	genConfigOnly bool
)

// genCmd gen 命令
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "生成完整配置模板和示例脚本",
	Long: `生成完整的 xbuilder 配置模板和示例脚本文件。

将会创建以下文件:
  xbuilder.yaml           完整配置文件模板
  scripts/build.sh        构建脚本示例
  scripts/deploy.sh       部署脚本示例
  scripts/notify.sh       通知脚本示例
  Dockerfile.example      Dockerfile 示例
  docker-compose.example.yml  docker-compose 示例`,
	Example: `  xbuilder gen            # 生成所有模板文件
  xbuilder gen -f         # 强制覆盖已存在的文件
  xbuilder gen --config   # 只生成配置文件`,
	RunE: runGen,
}

func init() {
	rootCmd.AddCommand(genCmd)
	genCmd.Flags().BoolVarP(&genForce, "force", "f", false, "强制覆盖已存在的文件")
	genCmd.Flags().BoolVar(&genConfigOnly, "config", false, "只生成配置文件")
}

func runGen(cmd *cobra.Command, args []string) error {
	// 定义要生成的文件
	files := []struct {
		path    string
		content string
		desc    string
	}{
		{"xbuilder.yaml", fullConfigTemplate, "完整配置文件"},
	}

	// 如果不是只生成配置文件，添加其他文件
	if !genConfigOnly {
		files = append(files, []struct {
			path    string
			content string
			desc    string
		}{
			{"scripts/build.sh", buildScriptTemplate, "构建脚本"},
			{"scripts/deploy.sh", deployScriptTemplate, "部署脚本"},
			{"scripts/notify.sh", notifyScriptTemplate, "通知脚本"},
			{"Dockerfile.example", dockerfileTemplate, "Dockerfile 示例"},
			{"docker-compose.example.yml", dockerComposeTemplate, "docker-compose 示例"},
		}...)
	}

	fmt.Println("📦 生成 xbuilder 模板文件...")
	fmt.Println()

	createdCount := 0
	skippedCount := 0

	for _, f := range files {
		// 创建目录
		dir := filepath.Dir(f.path)
		if dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("创建目录 %s 失败: %w", dir, err)
			}
		}

		// 检查文件是否已存在
		if _, err := os.Stat(f.path); err == nil {
			if !genForce {
				fmt.Printf("  ⏭️  跳过 %s (已存在)\n", f.path)
				skippedCount++
				continue
			}
		}

		// 写入文件
		perm := os.FileMode(0644)
		if filepath.Ext(f.path) == ".sh" {
			perm = 0755 // 脚本文件添加执行权限
		}

		if err := os.WriteFile(f.path, []byte(f.content), perm); err != nil {
			return fmt.Errorf("创建文件 %s 失败: %w", f.path, err)
		}

		fmt.Printf("  ✅ 创建 %s (%s)\n", f.path, f.desc)
		createdCount++
	}

	fmt.Println()
	fmt.Printf("完成! 创建 %d 个文件", createdCount)
	if skippedCount > 0 {
		fmt.Printf(", 跳过 %d 个文件", skippedCount)
	}
	fmt.Println()

	if createdCount > 0 {
		fmt.Println()
		fmt.Println("下一步:")
		fmt.Println("  1. 编辑 xbuilder.yaml 配置你的构建流程")
		fmt.Println("  2. 根据需要修改 scripts/ 目录下的脚本")
		fmt.Println("  3. 运行 'xbuilder build' 开始构建")
	}

	return nil
}
