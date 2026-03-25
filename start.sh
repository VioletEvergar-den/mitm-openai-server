#!/bin/bash

# MITM OpenAI Server 智能启动脚本
# 版本: 2.0
# 特性: 自动检测依赖、智能修复问题、新手友好

set -e

# ==================== 颜色定义 ====================
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

# ==================== 全局变量 ====================
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DATA_DIR="${SCRIPT_DIR}/data"
LOG_FILE="${SCRIPT_DIR}/server.log"
PORT=8081
VERBOSE=false
SKIP_BUILD=false
CLEAN_START=false

# ==================== 帮助信息 ====================
show_help() {
    echo -e "${BOLD}${CYAN}============================================${NC}"
    echo -e "${BOLD}${CYAN}    MITM OpenAI Server 智能启动脚本 v2.0${NC}"
    echo -e "${BOLD}${CYAN}============================================${NC}"
    echo ""
    echo -e "${BOLD}用法:${NC}"
    echo -e "  ./start.sh [选项]"
    echo ""
    echo -e "${BOLD}选项:${NC}"
    echo -e "  -h, --help       显示此帮助信息"
    echo -e "  -p, --port       指定服务器端口 (默认: 8081)"
    echo -e "  -v, --verbose    显示详细输出"
    echo -e "  -s, --skip-build 跳过构建步骤"
    echo -e "  -c, --clean      清理数据后启动 (谨慎使用)"
    echo -e "  --install-deps   自动安装缺失的依赖"
    echo ""
    echo -e "${BOLD}示例:${NC}"
    echo -e "  ./start.sh                    # 默认启动"
    echo -e "  ./start.sh -p 3000            # 使用端口3000"
    echo -e "  ./start.sh --install-deps     # 自动安装依赖"
    echo -e "  ./start.sh -s                 # 跳过构建快速启动"
    echo ""
    echo -e "${BOLD}登录信息:${NC}"
    echo -e "  启动后会显示用户名和密码"
    echo -e "  访问地址: http://localhost:${PORT}/ui/login"
    echo ""
}

# ==================== 日志函数 ====================
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
    if [ "$VERBOSE" = true ]; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] [INFO] $1" >> "$LOG_FILE"
    fi
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] [WARN] $1" >> "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] [ERROR] $1" >> "$LOG_FILE"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# ==================== 检测操作系统 ====================
detect_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$ID
        OS_VERSION=$VERSION_ID
    elif [ -f /etc/debian_version ]; then
        OS="debian"
    elif [ -f /etc/redhat-release ]; then
        OS="rhel"
    else
        OS="unknown"
    fi
    log_info "检测到操作系统: $OS $OS_VERSION"
}

# ==================== 检查并安装Go ====================
check_go() {
    log_step "检查 Go 环境..."
    
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
        log_info "Go 已安装: $GO_VERSION"
        
        # 检查版本是否满足要求 (>= 1.19)
        REQUIRED_VERSION="1.19"
        if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" = "$REQUIRED_VERSION" ]; then
            log_info "Go 版本满足要求 (>= $REQUIRED_VERSION)"
            return 0
        else
            log_warn "Go 版本过低，建议升级到 $REQUIRED_VERSION 或更高版本"
            log_warn "当前版本: $GO_VERSION"
        fi
        return 0
    fi
    
    log_error "未找到 Go 环境"
    echo ""
    echo -e "${YELLOW}请选择安装方式:${NC}"
    echo "  1) 自动安装 (需要 sudo 权限)"
    echo "  2) 显示手动安装说明"
    echo "  3) 跳过 (如果已经安装请检查 PATH)"
    echo ""
    read -p "请选择 [1-3]: " choice
    
    case $choice in
        1)
            install_go
            ;;
        2)
            show_go_install_guide
            exit 1
            ;;
        3)
            log_warn "跳过 Go 安装，如果构建失败请手动安装"
            return 1
            ;;
        *)
            log_error "无效选择"
            exit 1
            ;;
    esac
}

install_go() {
    log_step "正在安装 Go..."
    
    case $OS in
        ubuntu|debian)
            sudo apt update
            sudo apt install -y golang-go
            ;;
        centos|rhel|fedora)
            sudo yum install -y golang
            ;;
        *)
            log_error "不支持的操作系统，请手动安装 Go"
            show_go_install_guide
            exit 1
            ;;
    esac
    
    if command -v go &> /dev/null; then
        log_info "Go 安装成功: $(go version)"
    else
        log_error "Go 安装失败，请手动安装"
        exit 1
    fi
}

show_go_install_guide() {
    echo ""
    echo -e "${BOLD}${CYAN}========== Go 安装指南 ==========${NC}"
    echo ""
    echo -e "${BOLD}Ubuntu/Debian:${NC}"
    echo "  sudo apt update"
    echo "  sudo apt install -y golang-go"
    echo ""
    echo -e "${BOLD}CentOS/RHEL:${NC}"
    echo "  sudo yum install -y golang"
    echo ""
    echo -e "${BOLD}macOS (使用 Homebrew):${NC}"
    echo "  brew install go"
    echo ""
    echo -e "${BOLD}手动安装 (推荐最新版本):${NC}"
    echo "  1. 访问 https://go.dev/dl/"
    echo "  2. 下载适合您系统的安装包"
    echo "  3. 按照官方文档安装"
    echo ""
    echo -e "${BOLD}安装后验证:${NC}"
    echo "  go version"
    echo ""
}

# ==================== 检查并安装Node.js ====================
check_nodejs() {
    log_step "检查 Node.js 环境..."
    
    if command -v node &> /dev/null && command -v npm &> /dev/null; then
        NODE_VERSION=$(node -v | sed 's/v//')
        NPM_VERSION=$(npm -v)
        log_info "Node.js 已安装: $NODE_VERSION"
        log_info "npm 已安装: $NPM_VERSION"
        return 0
    fi
    
    log_error "未找到 Node.js 或 npm"
    echo ""
    echo -e "${YELLOW}请选择安装方式:${NC}"
    echo "  1) 自动安装 (需要 sudo 权限)"
    echo "  2) 显示手动安装说明"
    echo "  3) 跳过 (如果已经安装请检查 PATH)"
    echo ""
    read -p "请选择 [1-3]: " choice
    
    case $choice in
        1)
            install_nodejs
            ;;
        2)
            show_nodejs_install_guide
            exit 1
            ;;
        3)
            log_warn "跳过 Node.js 安装，如果构建失败请手动安装"
            return 1
            ;;
        *)
            log_error "无效选择"
            exit 1
            ;;
    esac
}

install_nodejs() {
    log_step "正在安装 Node.js..."
    
    case $OS in
        ubuntu|debian)
            curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
            sudo apt install -y nodejs
            ;;
        centos|rhel|fedora)
            curl -fsSL https://rpm.nodesource.com/setup_18.x | sudo bash -
            sudo yum install -y nodejs
            ;;
        *)
            log_error "不支持的操作系统，请手动安装 Node.js"
            show_nodejs_install_guide
            exit 1
            ;;
    esac
    
    if command -v node &> /dev/null; then
        log_info "Node.js 安装成功: $(node -v)"
    else
        log_error "Node.js 安装失败，请手动安装"
        exit 1
    fi
}

show_nodejs_install_guide() {
    echo ""
    echo -e "${BOLD}${CYAN}========== Node.js 安装指南 ==========${NC}"
    echo ""
    echo -e "${BOLD}Ubuntu/Debian:${NC}"
    echo "  curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -"
    echo "  sudo apt install -y nodejs"
    echo ""
    echo -e "${BOLD}CentOS/RHEL:${NC}"
    echo "  curl -fsSL https://rpm.nodesource.com/setup_18.x | sudo bash -"
    echo "  sudo yum install -y nodejs"
    echo ""
    echo -e "${BOLD}macOS (使用 Homebrew):${NC}"
    echo "  brew install node"
    echo ""
    echo -e "${BOLD}使用 nvm 安装 (推荐):${NC}"
    echo "  curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash"
    echo "  nvm install 18"
    echo ""
    echo -e "${BOLD}安装后验证:${NC}"
    echo "  node -v"
    echo "  npm -v"
    echo ""
}

# ==================== 检查数据目录权限 ====================
check_data_dir() {
    log_step "检查数据目录..."
    
    if [ ! -d "$DATA_DIR" ]; then
        log_info "创建数据目录: $DATA_DIR"
        mkdir -p "$DATA_DIR"
    fi
    
    # 检查写权限
    if [ ! -w "$DATA_DIR" ]; then
        log_error "数据目录无写权限: $DATA_DIR"
        fix_data_dir_permission
    else
        log_info "数据目录权限正常"
    fi
    
    # 检查数据库文件权限
    DB_FILE="${DATA_DIR}/mitm_server.db"
    if [ -f "$DB_FILE" ] && [ ! -w "$DB_FILE" ]; then
        log_error "数据库文件无写权限: $DB_FILE"
        fix_data_dir_permission
    fi
}

fix_data_dir_permission() {
    CURRENT_USER=$(whoami)
    log_warn "尝试修复数据目录权限..."
    
    if [ "$CURRENT_USER" != "root" ]; then
        echo -e "${YELLOW}需要 sudo 权限来修复目录权限${NC}"
        sudo chown -R "$CURRENT_USER:$CURRENT_USER" "$DATA_DIR"
        sudo chmod -R 755 "$DATA_DIR"
    else
        chown -R "$CURRENT_USER:$CURRENT_USER" "$DATA_DIR"
        chmod -R 755 "$DATA_DIR"
    fi
    
    log_info "数据目录权限已修复"
}

# ==================== 清理旧进程 ====================
kill_old_process() {
    log_step "检查并停止旧进程..."
    
    # 检查端口占用
    if command -v lsof &> /dev/null; then
        PIDS=$(lsof -ti:$PORT 2>/dev/null || true)
        if [ -n "$PIDS" ]; then
            log_warn "发现占用端口 $PORT 的进程: $PIDS"
            for PID in $PIDS; do
                kill -TERM $PID 2>/dev/null || true
            done
            sleep 2
            
            # 强制杀死仍在运行的进程
            REMAINING=$(lsof -ti:$PORT 2>/dev/null || true)
            if [ -n "$REMAINING" ]; then
                for PID in $REMAINING; do
                    kill -KILL $PID 2>/dev/null || true
                done
            fi
            log_info "端口 $PORT 已释放"
        else
            log_info "端口 $PORT 未被占用"
        fi
    else
        # 使用 fuser 作为备选
        if command -v fuser &> /dev/null; then
            fuser -k $PORT/tcp 2>/dev/null || true
            log_info "已尝试释放端口 $PORT"
        fi
    fi
    
    # 杀死可能存在的旧进程
    pkill -f "mitm-openai-server" 2>/dev/null || true
}

# ==================== 构建前端 ====================
build_frontend() {
    log_step "构建前端应用..."
    
    cd "${SCRIPT_DIR}/react-ui"
    
    # 检查 node_modules
    if [ ! -d "node_modules" ]; then
        log_info "安装前端依赖..."
        npm install --legacy-peer-deps
    else
        # 检查是否需要更新依赖
        if [ package.json -nt node_modules ]; then
            log_info "更新前端依赖..."
            npm install --legacy-peer-deps
        fi
    fi
    
    log_info "编译前端代码..."
    npm run build
    
    if [ ! -d "dist" ]; then
        log_error "前端构建失败"
        exit 1
    fi
    
    log_info "前端构建完成"
    cd "$SCRIPT_DIR"
}

# ==================== 构建后端 ====================
build_backend() {
    log_step "构建后端应用..."
    
    cd "$SCRIPT_DIR"
    
    # 下载依赖
    log_info "下载 Go 模块依赖..."
    go mod download
    
    # 编译
    log_info "编译 Go 代码..."
    go build -o mitm-openai-server
    
    if [ ! -f "./mitm-openai-server" ]; then
        log_error "后端构建失败"
        exit 1
    fi
    
    log_info "后端构建完成"
}

# ==================== 创建默认登录凭据 ====================
create_login_credentials() {
    LOGIN_FILE="${DATA_DIR}/login.json"
    
    if [ ! -f "$LOGIN_FILE" ]; then
        log_info "创建默认登录凭据..."
        
        # 生成随机密码
        RANDOM_PASSWORD=$(openssl rand -base64 12 | tr -dc 'a-zA-Z0-9' | head -c 12)
        if [ -z "$RANDOM_PASSWORD" ]; then
            RANDOM_PASSWORD=$(date +%s | sha256sum | base64 | head -c 12)
        fi
        
        cat > "$LOGIN_FILE" << EOF
{
  "username": "root",
  "password": "${RANDOM_PASSWORD}"
}
EOF
        
        log_info "登录凭据已创建"
    fi
}

# ==================== 启动服务器 ====================
start_server() {
    log_step "启动服务器..."
    
    cd "$SCRIPT_DIR"
    
    echo ""
    echo -e "${BOLD}${GREEN}============================================${NC}"
    echo -e "${BOLD}${GREEN}    服务器启动中...${NC}"
    echo -e "${BOLD}${GREEN}============================================${NC}"
    echo ""
    
    # 启动服务器
    ./mitm-openai-server server --port $PORT
}

# ==================== 清理模式 ====================
clean_data() {
    log_warn "清理模式: 将删除所有数据!"
    echo ""
    read -p "确定要删除所有数据吗? [y/N]: " confirm
    if [ "$confirm" = "y" ] || [ "$confirm" = "Y" ]; then
        log_info "清理数据目录..."
        rm -rf "${DATA_DIR:?}"/*
        log_info "数据已清理"
    else
        log_info "取消清理"
    fi
}

# ==================== 主函数 ====================
main() {
    # 解析命令行参数
    INSTALL_DEPS=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -p|--port)
                PORT="$2"
                shift 2
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -s|--skip-build)
                SKIP_BUILD=true
                shift
                ;;
            -c|--clean)
                CLEAN_START=true
                shift
                ;;
            --install-deps)
                INSTALL_DEPS=true
                shift
                ;;
            *)
                log_error "未知选项: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # 显示欢迎信息
    echo -e "${BOLD}${CYAN}============================================${NC}"
    echo -e "${BOLD}${CYAN}    MITM OpenAI Server 智能启动脚本${NC}"
    echo -e "${BOLD}${CYAN}============================================${NC}"
    echo ""
    
    # 检测操作系统
    detect_os
    
    # 清理模式
    if [ "$CLEAN_START" = true ]; then
        clean_data
    fi
    
    # 检查依赖
    if [ "$INSTALL_DEPS" = true ]; then
        check_go || true
        check_nodejs || true
    else
        check_go || {
            log_error "Go 环境缺失，请使用 --install-deps 自动安装或手动安装"
            exit 1
        }
        check_nodejs || {
            log_error "Node.js 环境缺失，请使用 --install-deps 自动安装或手动安装"
            exit 1
        }
    fi
    
    # 检查数据目录
    check_data_dir
    
    # 创建登录凭据
    create_login_credentials
    
    # 构建应用
    if [ "$SKIP_BUILD" = false ]; then
        build_frontend
        build_backend
    else
        log_info "跳过构建步骤"
        if [ ! -f "./mitm-openai-server" ]; then
            log_error "可执行文件不存在，请先构建"
            exit 1
        fi
    fi
    
    # 清理旧进程
    kill_old_process
    
    # 启动服务器
    start_server
}

# 运行主函数
main "$@"
