FRONTEND_DIR = ./web
BACKEND_DIR = .

.PHONY: all build-frontend start-backend dev

all: build-frontend start-backend

# 生产环境构建：构建前端静态文件
build-frontend:
	@echo "Building frontend..."
	@cd $(FRONTEND_DIR) && bun install && DISABLE_ESLINT_PLUGIN='true' VITE_REACT_APP_VERSION=$$(cat VERSION) bun run build

# 普通启动模式（无热更新）
start-backend:
	@echo "Starting backend dev server..."
	@cd $(BACKEND_DIR) && go run main.go

# 推荐的开发模式：全栈热更新
dev:
	@chmod +x dev.sh
	@./dev.sh
