# xbuilder

一个基于 Bubble Tea + Lipgloss 的美观 TUI 构建工具，支持 Maven/Go 构建、Docker 镜像构建推送、SSH 远程部署。

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go)
![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)

## 特性

- **美观的 TUI 界面** - 实时任务队列、进度条、输出卡片
- **多语言构建** - 支持 Maven、Go 构建
- **Docker 镜像** - 并行构建多个 Dockerfile，支持多平台构建
- **SSH 部署** - 远程服务器执行部署脚本
- **YAML 配置** - 简洁的配置文件，支持变量替换
- **并行执行** - 阶段内任务可并行运行
- **阶段控制** - 支持只运行指定阶段

## 安装

### Homebrew (macOS/Linux)

```bash
brew tap xiaolfeng/tap
brew install xbuilder
```

### 从源码构建

```bash
git clone https://github.com/xiaolfeng/builder-cli.git
cd builder-cli
go build -o xbuilder .
```

### 使用 Go Install

```bash
go install github.com/xiaolfeng/builder-cli@latest
```

## 快速开始

```bash
# 1. 初始化最小配置文件
xbuilder init

# 2. 编辑配置文件
vim xbuilder.yaml

# 3. 运行构建
xbuilder build
```

## 命令参考

### init - 初始化配置

```bash
xbuilder init           # 创建最小配置文件
xbuilder init -f        # 强制覆盖已存在的文件
```

### gen - 生成模板 (父命令)

`gen` 命令包含多个子命令，用于生成各种模板文件。

#### gen config - 生成完整配置

```bash
xbuilder gen config              # 生成完整配置文件
xbuilder gen config --scripts    # 同时生成示例脚本
xbuilder gen config -f           # 强制覆盖
xbuilder gen config -o custom.yaml  # 自定义输出路径
```

#### gen dockerfile - 生成 Dockerfile

```bash
xbuilder gen dockerfile              # 生成 Go Dockerfile (默认)
xbuilder gen dockerfile --lang java  # 生成 Java Dockerfile
xbuilder gen dockerfile -o Dockerfile.prod  # 自定义输出路径
xbuilder gen df -l go                # 使用别名
```

支持的语言:
- `go` (默认) - 多阶段构建，scratch 基础镜像
- `java` - Spring Boot 应用

#### gen dockercompose - 生成 docker-compose

```bash
xbuilder gen dockercompose              # 生成全部三个环境文件 (默认)
xbuilder gen dockercompose --scope dev  # 仅生成开发环境
xbuilder gen dockercompose -s test      # 仅生成测试环境
xbuilder gen dc -s prod                 # 仅生成生产环境 (别名)
```

默认生成:
- `docker-compose.dev.yaml` - 开发环境 (含 db/redis)
- `docker-compose.test.yaml` - 测试环境 (资源限制)
- `docker-compose.yaml` - 生产环境

#### gen makefile - 生成 Makefile

```bash
xbuilder gen makefile                                    # 使用默认值
xbuilder gen makefile --project myapp --registry ghcr.io/me  # 自定义项目和仓库
xbuilder gen mk -p myapp -r docker.io/user               # 使用别名
```

### build - 运行构建

```bash
xbuilder build          # 运行全部阶段
xbuilder build 2        # 只运行第 2 个阶段
xbuilder build 1-3      # 运行第 1 到第 3 个阶段
xbuilder build 2-       # 从第 2 个阶段运行到最后
xbuilder build -3       # 从第 1 个阶段运行到第 3 个
xbuilder build -v       # 先验证配置，再运行

# 只执行指定任务
xbuilder build --only "用户服务镜像"
xbuilder build -o "用户服务" -o "订单服务"
```

### validate - 验证配置

```bash
xbuilder validate                # 验证默认配置文件
xbuilder validate -c custom.yaml # 验证指定配置文件
```

### 全局选项

```bash
xbuilder -c config.yaml <command>  # 指定配置文件
xbuilder --version                  # 显示版本
xbuilder --help                     # 显示帮助
```

## 配置说明

### 最小配置

```yaml
version: "1.0"

project:
  name: "my-project"

pipeline:
  - stage: "build"
    name: "构建"
    tasks:
      - name: "Maven 打包"
        type: "maven"
        config:
          command: "mvn clean package -DskipTests"

  - stage: "docker"
    name: "Docker 构建"
    tasks:
      - name: "构建镜像"
        type: "docker-build"
        config:
          dockerfile: "./Dockerfile"
          context: "."
          image_name: "my-project"
          tag: "latest"
```

### 完整配置示例

```yaml
version: "1.0"

project:
  name: "my-microservices"
  description: "微服务项目"

# 变量定义 (使用 ${VAR_NAME} 引用)
variables:
  APP_VERSION: "1.0.0"
  REGISTRY_PREFIX: "registry.example.com/myproject"

# Docker Registry 配置
registries:
  default:
    url: "registry.example.com"
    username: "${DOCKER_USERNAME}"
    password: "${DOCKER_PASSWORD}"

# SSH 服务器配置
servers:
  production:
    host: "192.168.1.100"
    port: 22
    username: "deploy"
    auth:
      type: "key"
      key_path: "~/.ssh/id_rsa"

# 构建流水线
pipeline:
  - stage: "build"
    name: "Go 构建"
    tasks:
      - name: "编译"
        type: "go-build"
        config:
          goos: "linux"
          goarch: "amd64"
          output: "bin/app"
          ldflags: "-s -w"

  - stage: "docker-build"
    name: "Docker 镜像构建"
    parallel: true
    tasks:
      - name: "API 服务镜像"
        type: "docker-build"
        config:
          dockerfile: "./Dockerfile"
          context: "."
          image_name: "${REGISTRY_PREFIX}/api"
          tag: "${APP_VERSION}"
          platforms:
            - "linux/amd64"
            - "linux/arm64"

  - stage: "docker-push"
    name: "推送镜像"
    tasks:
      - name: "推送所有镜像"
        type: "docker-push"
        config:
          registry: "default"
          auto: true

  - stage: "deploy"
    name: "部署"
    tasks:
      - name: "部署到生产环境"
        type: "ssh"
        config:
          server: "production"
          commands:
            - "cd /opt/services"
            - "docker-compose pull"
            - "docker-compose up -d"
```

### 任务类型

| 类型 | 说明 | 主要配置项 |
|------|------|------------|
| `maven` | Maven 构建 | `command`, `script`, `working_dir`, `timeout` |
| `go-build` | Go 构建 | `goos`, `goarch`, `output`, `ldflags`, `tags` |
| `docker-build` | Docker 镜像构建 | `dockerfile`, `context`, `image_name`, `tag`, `platforms` |
| `docker-push` | Docker 镜像推送 | `registry`, `images`, `auto`, `push_latest` |
| `ssh` | SSH 远程执行 | `server`, `commands`, `local_script`, `timeout` |

### 多平台 Docker 构建

```yaml
- name: "多平台镜像"
  type: "docker-build"
  config:
    dockerfile: "./Dockerfile"
    image_name: "myapp"
    tag: "latest"
    platforms:
      - "linux/amd64"
      - "linux/arm64"
    push_on_build: true  # 多平台构建时自动推送 (默认 true)
```

## 界面预览

```
┌─────────────────────────────────────────────────────────────────────┐
│  ⚡ xbuilder v1.0.0                                    [q] 退出     │
├─────────────────────────────────────────────────────────────────────┤
│  📋 任务队列                                                        │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  ✓ Go 构建 - 编译                                 完成       │   │
│  │  ● Docker 镜像构建 - API 服务镜像                 进行中     │   │
│  │  ○ 推送镜像到 Registry                            等待中     │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  📊 Overall Progress                                                │
│  ████████████░░░░░░░░░░░░░░░░░░░░░░░░░░░  2/5 (40%)               │
├─────────────────────────────────────────────────────────────────────┤
│  ⏱ 用时: 00:02:34  │  📦 阶段: docker-build  │  🔄 任务: 2/5        │
└─────────────────────────────────────────────────────────────────────┘
```

## 快捷键

| 按键 | 功能 |
|------|------|
| `q` | 退出 |
| `?` | 显示帮助 |

## 项目结构

```
builder-cli/
├── main.go                 # 程序入口
├── cmd/                    # CLI 命令
│   ├── root.go
│   ├── init.go
│   ├── gen.go              # gen 父命令 + 子命令
│   ├── build.go
│   └── validate.go
├── resources/              # 嵌入式模板
│   ├── embed.go
│   └── templates/
│       ├── config/
│       ├── dockerfile/
│       ├── dockercompose/
│       ├── makefile/
│       └── scripts/
├── internal/
│   ├── config/             # 配置加载与验证
│   ├── executor/           # 任务执行器
│   ├── pipeline/           # 流水线编排
│   └── tui/                # TUI 界面
└── pkg/
    └── version/
```

## 依赖

- [Cobra](https://github.com/spf13/cobra) - CLI 框架
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI 框架
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - 终端样式
- [golang.org/x/crypto/ssh](https://pkg.go.dev/golang.org/x/crypto/ssh) - SSH 客户端
- [gopkg.in/yaml.v3](https://pkg.go.dev/gopkg.in/yaml.v3) - YAML 解析

## License

MIT License - see [LICENSE](LICENSE)
