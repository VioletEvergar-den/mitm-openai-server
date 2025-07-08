#!/bin/bash

# 脚本错误时退出
set -e

# 定义颜色代码
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

echo -e "${BOLD}${CYAN}============================================${NC}"
echo -e "${BOLD}${CYAN}    MITM OpenAI Server 启动脚本${NC}"
echo -e "${BOLD}${CYAN}============================================${NC}"
echo

# 检查依赖
echo -e "${YELLOW}检查依赖...${NC}"

# 检查Node.js和npm
if ! command -v node &> /dev/null; then
    echo -e "${RED}错误: 未找到Node.js，请先安装Node.js${NC}"
    exit 1
fi

if ! command -v npm &> /dev/null; then
    echo -e "${RED}错误: 未找到npm，请先安装npm${NC}"
    exit 1
fi

# 检查Go
if ! command -v go &> /dev/null; then
    echo -e "${RED}错误: 未找到Go，请先安装Go${NC}"
    exit 1
fi

echo -e "${GREEN}✓ 依赖检查通过${NC}"
echo

# 清理旧的构建文件
echo -e "${YELLOW}清理旧的构建文件...${NC}"
rm -rf ./react-ui/dist
rm -f ./mitm-openai-server
echo -e "${GREEN}✓ 清理完成${NC}"
echo

# 构建前端
echo -e "${YELLOW}构建前端应用...${NC}"
cd ./react-ui

# 安装依赖
echo -e "  ${BLUE}安装前端依赖...${NC}"
npm install --legacy-peer-deps

# 构建前端
echo -e "  ${BLUE}编译前端代码...${NC}"
npm run build

# 检查构建结果
if [ ! -d "dist" ]; then
    echo -e "${RED}错误: 前端构建失败，没有生成dist目录${NC}"
    exit 1
fi

echo -e "${GREEN}✓ 前端构建完成${NC}"
cd ..
echo

# 构建后端
echo -e "${YELLOW}构建后端应用...${NC}"
echo -e "  ${BLUE}编译Go代码...${NC}"
go build -o mitm-openai-server

# 检查构建结果
if [ ! -f "./mitm-openai-server" ]; then
    echo -e "${RED}错误: 后端构建失败${NC}"
    exit 1
fi

echo -e "${GREEN}✓ 后端构建完成${NC}"
echo

# 杀死之前的实例
echo -e "${YELLOW}检查并停止之前的实例...${NC}"

# 函数：杀死占用指定端口的进程
kill_process_on_port() {
    local port=$1
    local pids=$(lsof -ti:$port 2>/dev/null || true)
    
    if [ -n "$pids" ]; then
        echo -e "  ${BLUE}发现占用端口 $port 的进程: $pids${NC}"
        echo -e "  ${BLUE}正在停止这些进程...${NC}"
        
        # 先尝试优雅停止
        for pid in $pids; do
            if kill -TERM $pid 2>/dev/null; then
                echo -e "    ${GREEN}✓ 发送TERM信号给进程 $pid${NC}"
            fi
        done
        
        # 等待2秒让进程优雅退出
        sleep 2
        
        # 检查是否还有进程在运行，如果有则强制杀死
        local remaining_pids=$(lsof -ti:$port 2>/dev/null || true)
        if [ -n "$remaining_pids" ]; then
            echo -e "    ${YELLOW}强制停止剩余进程...${NC}"
            for pid in $remaining_pids; do
                if kill -KILL $pid 2>/dev/null; then
                    echo -e "      ${GREEN}✓ 强制停止进程 $pid${NC}"
                fi
            done
        fi
        
        # 再次检查
        sleep 1
        local final_pids=$(lsof -ti:$port 2>/dev/null || true)
        if [ -z "$final_pids" ]; then
            echo -e "  ${GREEN}✓ 端口 $port 已释放${NC}"
        else
            echo -e "  ${RED}警告: 端口 $port 仍被占用，可能需要手动处理${NC}"
        fi
    else
        echo -e "  ${GREEN}✓ 端口 $port 未被占用${NC}"
    fi
}

# 同时也杀死可能存在的mitm-openai-server进程
echo -e "  ${BLUE}检查mitm-openai-server进程...${NC}"
pkill -f "mitm-openai-server.*server" 2>/dev/null && echo -e "  ${GREEN}✓ 停止了现有的mitm-openai-server进程${NC}" || echo -e "  ${GREEN}✓ 没有发现运行中的mitm-openai-server进程${NC}"

# 杀死占用8081端口的进程
kill_process_on_port 8081

echo -e "${GREEN}✓ 实例清理完成${NC}"
echo

# 启动服务器
echo -e "${BOLD}${GREEN}启动服务器...${NC}"
echo -e "${PURPLE}注意: 服务器启动后会显示登录信息${NC}"
echo -e "${PURPLE}主页访问: http://localhost:8081/${NC}"
echo -e "${PURPLE}直接登录: http://localhost:8081/ui/login${NC}"
echo -e "${PURPLE}按 Ctrl+C 可以停止服务器${NC}"
echo

# 启动服务器
./mitm-openai-server server --port 8081 