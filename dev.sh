#!/bin/bash

# --- New API 增强型开发启动脚本 (支持热更新 & 端口检查) ---

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}>>> 正在初始化 New API 开发环境...${NC}"

# 端口定义
BACKEND_PORT=3000
FRONTEND_PORT=5173

# 端口检查函数
check_port() {
    local port=$1
    local name=$2
    local pid=$(lsof -ti :$port)
    
    if [ ! -z "$pid" ]; then
        echo -e "${RED}警告: 端口 $port ($name) 已被占用 (PID: $pid)${NC}"
        local proc_info=$(ps -p $pid -o command=)
        echo -e "${YELLOW}占用程序: $proc_info${NC}"
        
        read -p "是否尝试强制杀掉该进程并继续？[y/N] " confirm
        if [[ "$confirm" =~ ^[Yy]$ ]]; then
            echo -e "${YELLOW}正在清理端口 $port...${NC}"
            kill -9 $pid
            sleep 1
        else
            echo -e "${RED}启动中止。请手动处理端口占用或修改配置。${NC}"
            exit 1
        fi
    fi
}

# 执行端口检查
check_port $BACKEND_PORT "后端服务"
check_port $FRONTEND_PORT "前端开发服务器"

# 1. 检查基础环境
command -v go >/dev/null 2>&1 || { echo -e "${RED}错误: 未找到 Go 环境。请先安装 Go 1.22+${NC}" >&2; exit 1; }
command -v bun >/dev/null 2>&1 || { echo -e "${RED}错误: 未找到 Bun 环境。建议安装 Bun 以获得最佳体验。${NC}" >&2; exit 1; }

# 2. 检查并安装 Air (后端热更新)
GOPATH=$(go env GOPATH)
AIR_BIN="$GOPATH/bin/air"

if ! command -v air >/dev/null 2>&1 && [ ! -f "$AIR_BIN" ]; then
    echo -e "${YELLOW}>>> 未找到 Air，正在尝试安装以支持后端热更新...${NC}"
    go install github.com/air-verse/air@latest
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}>>> Air 安装成功！${NC}"
    else
        echo -e "${YELLOW}>>> Air 安装失败，后端将回退到普通模式 (不支持热更新)。${NC}"
    fi
fi

# 3. 初始化后端依赖
echo -e "${BLUE}>>> 检查后端依赖...${NC}"
go mod download

# 4. 初始化前端依赖 (classic 主题)
echo -e "${BLUE}>>> 检查前端依赖 (使用 Bun)...${NC}"
cd web/classic
if [ ! -d "node_modules" ]; then
    echo -e "${YELLOW}>>> 正在安装前端依赖，这可能需要一点时间...${NC}"
    bun install
fi
cd ../..

# 5. 确保 web/dist 存在 (Go embed 必须)
if [ ! -d "web/dist" ]; then
    echo -e "${BLUE}>>> 创建临时 dist 目录...${NC}"
    mkdir -p web/dist
    touch web/dist/index.html
fi

# 启动流程
echo -e "${GREEN}>>> 准备就绪，启动服务...${NC}"

# 启动后端
if command -v air >/dev/null 2>&1; then
    air &
elif [ -f "$AIR_BIN" ]; then
    "$AIR_BIN" &
else
    go run main.go &
fi
BACKEND_PID=$!

# 启动前端
cd web/classic
bun run dev &
FRONTEND_PID=$!
cd ../..

echo -e "${GREEN}>>> 服务已全部启动！${NC}"
echo -e "${BLUE}>>> 后端: http://localhost:$BACKEND_PORT${NC}"
echo -e "${BLUE}>>> 前端: http://localhost:$FRONTEND_PORT (支持热更新)${NC}"
echo -e "${YELLOW}>>> 按 Ctrl+C 停止所有服务${NC}"

# 清理逻辑 (使用函数而非内联命令，避免 trap 嵌套问题)
cleanup() {
    echo -e "\n${BLUE}>>> 正在清理并退出...${NC}"
    # 先发 SIGINT 让进程自己清理子进程
    kill -INT $BACKEND_PID $FRONTEND_PID 2>/dev/null
    sleep 1
    # 再发 SIGTERM
    kill -TERM $BACKEND_PID $FRONTEND_PID 2>/dev/null
    sleep 1
    # 如果还有残留，强制杀掉（含子进程）
    kill -KILL $BACKEND_PID $FRONTEND_PID 2>/dev/null
    # 清理残留的后台子进程（air 启动的 go run、bun 启动的 vite node 等）
    pkill -P $BACKEND_PID 2>/dev/null
    pkill -P $FRONTEND_PID 2>/dev/null
    exit 0
}
trap cleanup INT TERM

wait
