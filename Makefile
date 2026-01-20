.PHONY: all build run test clean docker-up docker-down dev

# 默认目标
all: build

# 构建后端
build:
	go build -o bin/server ./cmd/server

# 运行后端
run:
	go run ./cmd/server

# 运行测试
test:
	go test -v ./...

# 清理构建产物
clean:
	rm -rf bin/
	rm -rf frontend/.next/

# Docker Compose 启动所有服务
docker-up:
	docker-compose up -d

# Docker Compose 停止所有服务
docker-down:
	docker-compose down

# Docker Compose 重新构建并启动
docker-rebuild:
	docker-compose up -d --build

# 开发模式：启动后端
dev-backend:
	go run ./cmd/server

# 开发模式：启动前端
dev-frontend:
	cd frontend && npm run dev

# 开发模式：并行启动前后端
dev:
	@echo "启动后端..."
	@go run ./cmd/server &
	@echo "启动前端..."
	@cd frontend && npm run dev

# 安装依赖
deps:
	go mod tidy
	cd frontend && npm install

# 格式化代码
fmt:
	go fmt ./...
	goimports -w .

# 代码检查
lint:
	golangci-lint run ./...

# 生成 API 文档
docs:
	swag init -g cmd/server/main.go -o docs/

# 帮助信息
help:
	@echo "Browser Automation Studio - Makefile 命令"
	@echo ""
	@echo "构建和运行:"
	@echo "  make build         - 构建后端二进制"
	@echo "  make run           - 运行后端服务"
	@echo "  make dev           - 开发模式（前后端并行）"
	@echo "  make dev-backend   - 仅启动后端"
	@echo "  make dev-frontend  - 仅启动前端"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-up     - 启动所有 Docker 服务"
	@echo "  make docker-down   - 停止所有 Docker 服务"
	@echo "  make docker-rebuild- 重新构建并启动"
	@echo ""
	@echo "开发工具:"
	@echo "  make deps          - 安装依赖"
	@echo "  make test          - 运行测试"
	@echo "  make fmt           - 格式化代码"
	@echo "  make lint          - 代码检查"
	@echo "  make clean         - 清理构建产物"
