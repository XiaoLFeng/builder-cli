#!/bin/bash
# xbuilder 更新脚本
# 支持 Linux, macOS, FreeBSD
# 使用方法:
#   curl -fsSL https://raw.githubusercontent.com/XiaoLFeng/builder-cli/master/scripts/update.sh | bash
#   curl -fsSL https://raw.githubusercontent.com/XiaoLFeng/builder-cli/master/scripts/update.sh | bash -s v1.0.0

set -e

# 配置
REPO="XiaoLFeng/builder-cli"
BINARY_NAME="xbuilder"
INSTALL_DIR="$HOME/.local/bin"
GITHUB_API="https://api.github.com/repos/${REPO}/releases"
GITHUB_DOWNLOAD="https://github.com/${REPO}/releases/download"

# 颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

info() { echo -e "${BLUE}[INFO]${NC} $1"; }
success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

# 检测操作系统
detect_os() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    case "$OS" in
        linux*)   OS="linux" ;;
        darwin*)  OS="darwin" ;;
        freebsd*) OS="freebsd" ;;
        *)        error "不支持的操作系统: $OS" ;;
    esac
    echo "$OS"
}

# 检测架构
detect_arch() {
    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64|amd64)   ARCH="amd64" ;;
        aarch64|arm64)  ARCH="arm64" ;;
        *)              error "不支持的架构: $ARCH" ;;
    esac
    echo "$ARCH"
}

# 获取当前版本
get_current_version() {
    local binary="$INSTALL_DIR/$BINARY_NAME"
    if [ -f "$binary" ]; then
        "$binary" --version 2>/dev/null | grep -oE 'v?[0-9]+\.[0-9]+\.[0-9]+' | head -1 || echo ""
    else
        echo ""
    fi
}

# 获取最新版本
get_latest_version() {
    local version
    version=$(curl -sL "${GITHUB_API}/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$version" ]; then
        error "无法获取最新版本"
    fi
    echo "$version"
}

# 版本比较 (返回: 0=相等, 1=v1>v2, 2=v1<v2)
compare_versions() {
    local v1=${1#v}
    local v2=${2#v}

    if [ "$v1" = "$v2" ]; then
        echo 0
        return
    fi

    local IFS=.
    local i v1_arr=($v1) v2_arr=($v2)

    for ((i=0; i<${#v1_arr[@]} || i<${#v2_arr[@]}; i++)); do
        local n1=${v1_arr[i]:-0}
        local n2=${v2_arr[i]:-0}
        if ((n1 > n2)); then
            echo 1
            return
        elif ((n1 < n2)); then
            echo 2
            return
        fi
    done

    echo 0
}

# 下载文件
download_file() {
    local url=$1
    local output=$2
    local retry=0
    local max_retries=3

    while [ $retry -lt $max_retries ]; do
        if curl -fsSL "$url" -o "$output"; then
            return 0
        fi
        retry=$((retry + 1))
        warn "下载失败，重试 ($retry/$max_retries)..."
        sleep 2
    done
    error "下载失败: $url"
}

# 验证校验和
verify_checksum() {
    local file=$1
    local expected_sha=$2

    if command -v sha256sum &> /dev/null; then
        actual_sha=$(sha256sum "$file" | awk '{print $1}')
    elif command -v shasum &> /dev/null; then
        actual_sha=$(shasum -a 256 "$file" | awk '{print $1}')
    else
        warn "无法验证校验和"
        return 0
    fi

    if [ "$actual_sha" != "$expected_sha" ]; then
        error "校验和不匹配!"
    fi
    success "校验和验证通过"
}

# 主更新函数
main() {
    echo -e "${CYAN}"
    echo "╔═══════════════════════════════════════╗"
    echo "║       xbuilder 更新脚本               ║"
    echo "╚═══════════════════════════════════════╝"
    echo -e "${NC}"

    # 检测环境
    OS=$(detect_os)
    ARCH=$(detect_arch)
    info "检测到系统: ${OS}/${ARCH}"

    # 检查当前安装
    CURRENT_VERSION=$(get_current_version)
    if [ -z "$CURRENT_VERSION" ]; then
        warn "${BINARY_NAME} 未安装，将执行全新安装"
        CURRENT_VERSION="v0.0.0"
    else
        info "当前版本: ${CURRENT_VERSION}"
    fi

    # 获取目标版本
    TARGET_VERSION=${1:-$(get_latest_version)}
    info "目标版本: ${TARGET_VERSION}"

    # 版本比较
    CMP=$(compare_versions "$TARGET_VERSION" "$CURRENT_VERSION")
    case $CMP in
        0)
            success "已是最新版本 (${CURRENT_VERSION})"
            exit 0
            ;;
        1)
            info "将从 ${CURRENT_VERSION} 升级到 ${TARGET_VERSION}"
            ;;
        2)
            warn "目标版本 (${TARGET_VERSION}) 低于当前版本 (${CURRENT_VERSION})"
            read -p "确认降级? [y/N] " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                info "取消操作"
                exit 0
            fi
            ;;
    esac

    # 创建临时目录
    TMP_DIR=$(mktemp -d)
    trap "rm -rf $TMP_DIR" EXIT

    # 下载新版本
    BINARY_FILE="${BINARY_NAME}-${OS}-${ARCH}"
    DOWNLOAD_URL="${GITHUB_DOWNLOAD}/${TARGET_VERSION}/${BINARY_FILE}"
    CHECKSUM_URL="${GITHUB_DOWNLOAD}/${TARGET_VERSION}/checksums.txt"

    info "下载 ${BINARY_FILE}..."
    download_file "$DOWNLOAD_URL" "$TMP_DIR/$BINARY_FILE"

    info "下载校验和..."
    download_file "$CHECKSUM_URL" "$TMP_DIR/checksums.txt"

    EXPECTED_SHA=$(grep "$BINARY_FILE" "$TMP_DIR/checksums.txt" | awk '{print $1}')
    if [ -n "$EXPECTED_SHA" ]; then
        verify_checksum "$TMP_DIR/$BINARY_FILE" "$EXPECTED_SHA"
    fi

    # 备份旧版本
    BINARY_PATH="$INSTALL_DIR/$BINARY_NAME"
    if [ -f "$BINARY_PATH" ]; then
        info "备份旧版本..."
        cp "$BINARY_PATH" "${BINARY_PATH}.backup"
    fi

    # 安装新版本
    info "安装新版本..."
    mkdir -p "$INSTALL_DIR"
    mv "$TMP_DIR/$BINARY_FILE" "$BINARY_PATH"
    chmod +x "$BINARY_PATH"

    # 验证安装
    NEW_VERSION=$("$BINARY_PATH" --version 2>/dev/null | grep -oE 'v?[0-9]+\.[0-9]+\.[0-9]+' | head -1 || echo "unknown")
    if [ -f "${BINARY_PATH}.backup" ]; then
        rm -f "${BINARY_PATH}.backup"
    fi

    success "${BINARY_NAME} 已更新到 ${NEW_VERSION}!"
}

main "$@"
