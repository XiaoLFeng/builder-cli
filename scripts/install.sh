#!/bin/bash
# xbuilder 安装脚本
# 支持 Linux, macOS, FreeBSD
# 使用方法:
#   curl -fsSL https://raw.githubusercontent.com/XiaoLFeng/builder-cli/master/scripts/install.sh | bash
#   curl -fsSL https://raw.githubusercontent.com/XiaoLFeng/builder-cli/master/scripts/install.sh | bash -s v1.0.0

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
NC='\033[0m' # No Color

# 日志函数
info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

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

# 获取最新版本
get_latest_version() {
    local version
    version=$(curl -sL "${GITHUB_API}/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$version" ]; then
        error "无法获取最新版本，请检查网络连接"
    fi
    echo "$version"
}

# 下载文件 (带重试)
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

    info "验证文件完整性..."

    if command -v sha256sum &> /dev/null; then
        actual_sha=$(sha256sum "$file" | awk '{print $1}')
    elif command -v shasum &> /dev/null; then
        actual_sha=$(shasum -a 256 "$file" | awk '{print $1}')
    else
        warn "无法验证校验和 (sha256sum/shasum 不可用)"
        return 0
    fi

    if [ "$actual_sha" != "$expected_sha" ]; then
        error "校验和不匹配!\n期望: $expected_sha\n实际: $actual_sha"
    fi

    success "校验和验证通过"
}

# 主安装函数
main() {
    echo -e "${CYAN}"
    echo "╔═══════════════════════════════════════╗"
    echo "║       xbuilder 安装脚本               ║"
    echo "║   Build & Deploy Pipeline CLI Tool    ║"
    echo "╚═══════════════════════════════════════╝"
    echo -e "${NC}"

    # 检测环境
    OS=$(detect_os)
    ARCH=$(detect_arch)
    info "检测到系统: ${OS}/${ARCH}"

    # 获取版本
    VERSION=${1:-$(get_latest_version)}
    info "安装版本: ${VERSION}"

    # 构建下载 URL
    BINARY_FILE="${BINARY_NAME}-${OS}-${ARCH}"
    DOWNLOAD_URL="${GITHUB_DOWNLOAD}/${VERSION}/${BINARY_FILE}"
    CHECKSUM_URL="${GITHUB_DOWNLOAD}/${VERSION}/checksums.txt"

    # 创建临时目录
    TMP_DIR=$(mktemp -d)
    trap "rm -rf $TMP_DIR" EXIT

    # 下载二进制文件
    info "下载 ${BINARY_FILE}..."
    download_file "$DOWNLOAD_URL" "$TMP_DIR/$BINARY_FILE"

    # 下载并验证校验和
    info "下载校验和文件..."
    download_file "$CHECKSUM_URL" "$TMP_DIR/checksums.txt"

    EXPECTED_SHA=$(grep "$BINARY_FILE" "$TMP_DIR/checksums.txt" | awk '{print $1}')
    if [ -n "$EXPECTED_SHA" ]; then
        verify_checksum "$TMP_DIR/$BINARY_FILE" "$EXPECTED_SHA"
    else
        warn "未找到校验和，跳过验证"
    fi

    # 创建安装目录
    info "安装到 ${INSTALL_DIR}..."
    mkdir -p "$INSTALL_DIR"

    # 安装二进制文件
    mv "$TMP_DIR/$BINARY_FILE" "$INSTALL_DIR/$BINARY_NAME"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"

    success "${BINARY_NAME} ${VERSION} 安装成功!"

    # 检查 PATH
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        echo ""
        warn "安装目录 ($INSTALL_DIR) 不在 PATH 中"
        echo ""
        echo -e "请将以下内容添加到你的 shell 配置文件 (~/.bashrc, ~/.zshrc 等):"
        echo ""
        echo -e "  ${CYAN}export PATH=\"\$HOME/.local/bin:\$PATH\"${NC}"
        echo ""
        echo "然后运行: source ~/.bashrc (或 source ~/.zshrc)"
    fi

    echo ""
    echo -e "使用方法: ${GREEN}${BINARY_NAME} --help${NC}"
    echo ""
}

# 运行
main "$@"
