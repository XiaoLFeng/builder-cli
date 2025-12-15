# xbuilder

一个基于 Bubble Tea + Lipgloss 的美观 TUI 构建工具，支持 Maven 构建、Docker 镜像构建推送、SSH 远程部署。

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go)
![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)

## 特性

- **美观的 TUI 界面** - 实时任务队列、进度条、输出卡片
- **Maven 构建** - 支持自定义命令或脚本
- **Docker 镜像** - 并行构建多个 Dockerfile，自动扫描微服务
- **SSH 部署** - 远程服务器执行部署脚本
- **YAML 配置** - 简洁的配置文件，支持变量替换
- **并行执行** - 阶段内任务可并行运行
- **阶段控制** - 支持只运行指定阶段

## 安装

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

### 方式 1: 使用 init 命令 (推荐)

```bash
# 初始化最小配置文件
xbuilder init

# 编辑配置文件
vim xbuilder.yaml

# 运行构建
xbuilder build
```

### 方式 2: 使用 gen 命令生成完整模板

```bash
# 生成完整模板和示例脚本
xbuilder gen

# 会创建以下文件:
#   xbuilder.yaml              完整配置模板
#   scripts/build.sh           构建脚本示例
#   scripts/deploy.sh          部署脚本示例
#   scripts/notify.sh          通知脚本示例
#   Dockerfile.example         Dockerfile 示例
#   docker-compose.example.yml docker-compose 示例
```

## 命令详解

### 初始化配置

```bash
xbuilder init           # 创建最小配置文件
xbuilder init -f        # 强制覆盖已存在的文件
```

### 生成模板

```bash
xbuilder gen            # 生成完整模板和脚本
xbuilder gen -f         # 强制覆盖已存在的文件
xbuilder gen --config   # 只生成配置文件
```

### 运行构建

```bash
xbuilder build          # 运行全部阶段
xbuilder build 2        # 只运行第 2 个阶段
xbuilder build 1-3      # 运行第 1 到第 3 个阶段
xbuilder build 2-       # 从第 2 个阶段运行到最后
xbuilder build -3       # 从第 1 个阶段运行到第 3 个
xbuilder build -v       # 先验证配置，再运行

# 使用 --only 参数只执行指定的任务
xbuilder build --only "用户服务镜像"              # 只执行名为"用户服务镜像"的任务
xbuilder build -o "用户服务" -o "订单服务"        # 同时执行多个指定的任务
xbuilder build 2 --only "用户服务镜像"            # 在第 2 阶段中只执行指定任务
```

### 验证配置

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

## 界面预览

```
┌─────────────────────────────────────────────────────────────────────┐
│  ⚡ xbuilder v1.0.0                                    [q] 退出     │
├─────────────────────────────────────────────────────────────────────┤
│  📋 任务队列                                                        │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  ✓ Maven 构建 - 编译打包                          完成       │   │
│  │  ● Docker 镜像构建 - 用户服务镜像                 进行中     │   │
│  │  ○ 推送镜像到 Registry                            等待中     │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  📊 Overall Progress                                                │
│  ████████████░░░░░░░░░░░░░░░░░░░░░░░░░░░  2/5 (40%)               │
├─────────────────────────────────────────────────────────────────────┤
│  🔧 当前任务                                                        │
│  ┌─ 用户服务镜像 ──────────────────┐ ┌─ 订单服务镜像 ──────────────┐│
│  │ Step 2/5: Copying files...     │ │ Step 1/5: FROM openjdk:17  ││
│  │ COPY target/*.jar app.jar      │ │ ---> Using cache           ││
│  └────────────────────────────────┘ └────────────────────────────┘│
├─────────────────────────────────────────────────────────────────────┤
│  ⏱ 用时: 00:02:34  │  📦 阶段: docker-build  │  🔄 任务: 2/5        │
└─────────────────────────────────────────────────────────────────────┘
```

## 配置说明

### 最小配置 (`xbuilder init`)

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

# 项目信息
project:
  name: "my-microservices"
  description: "微服务项目"

# 变量定义 (可使用 ${VAR_NAME} 引用)
variables:
  APP_VERSION: "1.0.0"
  REGISTRY_PREFIX: "registry.example.com/myproject"

# Docker Registry 配置
registries:
  default:
    url: "registry.example.com"
    username: "${DOCKER_USERNAME}"    # 从环境变量读取
    password: "${DOCKER_PASSWORD}"

# SSH 服务器配置
servers:
  production:
    host: "192.168.1.100"
    port: 22
    username: "deploy"
    auth:
      type: "key"                     # "password" | "key"
      key_path: "~/.ssh/id_rsa"

# 构建流水线
pipeline:
  # 阶段 1: Maven 构建
  - stage: "maven-build"
    name: "Maven 构建"
    tasks:
      - name: "编译打包"
        type: "maven"
        config:
          command: "mvn clean package -DskipTests"
          working_dir: "."
          timeout: 600                # 超时 (秒)

  # 阶段 2: Docker 构建 (并行)
  - stage: "docker-build"
    name: "Docker 镜像构建"
    parallel: true
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

      - name: "订单服务镜像"
        type: "docker-build"
        config:
          dockerfile: "./order-service/Dockerfile"
          context: "./order-service"
          image_name: "${REGISTRY_PREFIX}/order-service"
          tag: "${APP_VERSION}"

  # 阶段 3: Docker 推送
  - stage: "docker-push"
    name: "推送镜像"
    tasks:
      - name: "推送所有镜像"
        type: "docker-push"
        config:
          registry: "default"
          auto: true                  # 自动推送上阶段构建的镜像

  # 阶段 4: 部署
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
          timeout: 300

# 钩子 (可选)
hooks:
  pre_build:
    - "echo '开始构建...'"
  post_build:
    - "echo '构建完成!'"
  on_failure:
    - "echo '构建失败!'"
```

### 任务类型

| 类型             | 说明            | 配置项                                                       |
| ---------------- | --------------- | ------------------------------------------------------------ |
| `maven`          | Maven 构建      | `command`, `script`, `working_dir`, `timeout`                |
| `docker-build`   | Docker 镜像构建 | `dockerfile`, `context`, `image_name`, `tag`, `build_args`   |
| `docker-push`    | Docker 镜像推送 | `registry`, `images`, `auto`                                 |
| `ssh`            | SSH 远程执行    | `server`, `commands`, `local_script`, `timeout`              |

### 变量替换

支持在配置中使用 `${VAR_NAME}` 语法引用变量:

- 优先从 `variables` 部分获取
- 如果未定义，则从环境变量获取
- 支持嵌套引用

## 快捷键 (TUI 模式)

| 按键      | 功能     |
| --------- | -------- |
| `Enter`   | 开始构建 |
| `q`       | 退出     |
| `?`       | 显示帮助 |

## 配置文件查找顺序

1. `xbuilder.yaml`
2. `xbuilder.yml`
3. `.xbuilder.yaml`
4. `.xbuilder.yml`

## 依赖

- [Cobra](https://github.com/spf13/cobra) - CLI 框架
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI 框架
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - 终端样式
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI 组件
- [golang.org/x/crypto/ssh](https://pkg.go.dev/golang.org/x/crypto/ssh) - SSH 客户端
- [gopkg.in/yaml.v3](https://pkg.go.dev/gopkg.in/yaml.v3) - YAML 解析

## 项目结构

```
builder-cli/
├── main.go                          # 程序入口
├── cmd/                             # CLI 命令
│   ├── root.go                      # 根命令
│   ├── init.go                      # init 命令
│   ├── gen.go                       # gen 命令
│   ├── build.go                     # build 命令
│   └── validate.go                  # validate 命令
├── internal/
│   ├── app/
│   │   └── app.go                   # 应用逻辑
│   ├── config/
│   │   ├── config.go                # 配置结构定义
│   │   ├── loader.go                # YAML 配置加载器
│   │   └── validator.go             # 配置验证器
│   ├── tui/
│   │   ├── model.go                 # 主 Model (Bubble Tea)
│   │   ├── view.go                  # View 渲染
│   │   ├── keys.go                  # 快捷键绑定
│   │   ├── messages.go              # 消息类型
│   │   └── components/
│   │       ├── todolist/            # Todo List 组件
│   │       ├── progressbar/         # 进度条组件
│   │       ├── taskcard/            # 任务卡片组件
│   │       └── statusbar/           # 状态栏组件
│   ├── executor/
│   │   ├── executor.go              # 执行器接口
│   │   ├── runner.go                # 通用命令运行器
│   │   ├── maven.go                 # Maven 执行器
│   │   ├── docker.go                # Docker 执行器
│   │   └── ssh.go                   # SSH 执行器
│   ├── pipeline/
│   │   ├── pipeline.go              # 流水线编排器
│   │   ├── stage.go                 # 阶段管理
│   │   └── task.go                  # 任务管理
│   ├── styles/
│   │   └── styles.go                # 共享样式定义
│   └── types/
│       └── messages.go              # 共享类型定义
└── pkg/
    └── version/
        └── version.go               # 版本信息
```

## License

MIT License
