#!/bin/bash
# xbuilder 卸载脚本
# 支持 Linux, macOS, FreeBSD
# 使用方法:
#   curl -fsSL https://raw.githubusercontent.com/XiaoLFeng/builder-cli/master/scripts/uninstall.sh | bash

set -e

# 配置
BINARY_NAME="xbuilder"
INSTALL_DIR="$HOME/.local/bin"
CONFIG_DIR="$HOME/.xbuilder"

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

main() {
    echo -e "${CYAN}"
    echo "╔═══════════════════════════════════════╗"
    echo "║       xbuilder 卸载脚本               ║"
    echo "╚═══════════════════════════════════════╝"
    echo -e "${NC}"

    BINARY_PATH="$INSTALL_DIR/$BINARY_NAME"
    FOUND_BINARY=false
    FOUND_CONFIG=false

    # 检查二进制文件
    if [ -f "$BINARY_PATH" ]; then
        FOUND_BINARY=true
        info "找到二进制文件: $BINARY_PATH"
    else
        warn "未找到二进制文件: $BINARY_PATH"
        # 尝试在 PATH 中查找
        OTHER_PATH=$(which "$BINARY_NAME" 2>/dev/null || true)
        if [ -n "$OTHER_PATH" ]; then
            warn "在其他位置找到 $BINARY_NAME: $OTHER_PATH"
            echo "如需删除，请手动执行: rm $OTHER_PATH"
        fi
    fi

    # 检查配置目录
    if [ -d "$CONFIG_DIR" ]; then
        FOUND_CONFIG=true
        info "找到配置目录: $CONFIG_DIR"
    fi

    # 如果什么都没找到
    if [ "$FOUND_BINARY" = false ] && [ "$FOUND_CONFIG" = false ]; then
        warn "${BINARY_NAME} 似乎未安装"
        exit 0
    fi

    # 确认卸载
    echo ""
    echo "将删除以下内容:"
    if [ "$FOUND_BINARY" = true ]; then
        echo "  - 二进制文件: $BINARY_PATH"
    fi
    if [ "$FOUND_CONFIG" = true ]; then
        echo "  - 配置目录: $CONFIG_DIR (包含所有配置和数据)"
    fi
    echo ""

    read -p "确认卸载? [y/N] " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        info "取消卸载"
        exit 0
    fi

    # 删除二进制文件
    if [ "$FOUND_BINARY" = true ]; then
        info "删除二进制文件..."
        rm -f "$BINARY_PATH"
        success "已删除: $BINARY_PATH"
    fi

    # 询问是否删除配置
    if [ "$FOUND_CONFIG" = true ]; then
        echo ""
        read -p "是否同时删除配置目录 ($CONFIG_DIR)? [y/N] " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            info "删除配置目录..."
            rm -rf "$CONFIG_DIR"
            success "已删除: $CONFIG_DIR"
        else
            info "保留配置目录: $CONFIG_DIR"
        fi
    fi

    echo ""
    success "${BINARY_NAME} 卸载完成!"
    echo ""
    echo "如果你是通过 Homebrew 安装的，请使用:"
    echo -e "  ${CYAN}brew uninstall ${BINARY_NAME}${NC}"
}

main "$@"
