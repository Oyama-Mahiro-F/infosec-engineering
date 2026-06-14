# =============================================================================
# 乒乓球拍防伪溯源系统 - Makefile
# =============================================================================

.PHONY: help up down logs test clean build

help: ## 显示帮助信息
	@echo "乒乓球拍防伪溯源系统 - 可用命令："
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

up: ## 启动所有Docker服务
	docker compose up -d
	@echo "✅ 服务已启动"
	@echo "   前端:  http://localhost:3000/consumer"
	@echo "   API:   http://localhost:8080/api/v1/health"

down: ## 停止所有Docker服务
	docker compose down

logs: ## 查看API服务日志
	docker compose logs -f api

restart: ## 重启API服务
	docker compose restart api

ps: ## 查看服务状态
	docker compose ps

test-contract: ## 运行智能合约测试
	cd test && go test -v -run TestContract

test-api: ## 运行API集成测试
	cd test && go test -v -run TestAPI

test-security: ## 运行安全测试
	cd test && go test -v -run TestSecurity

test-all: ## 运行全部测试
	cd test && go test -v ./...

bench: ## 运行性能基准测试
	cd test && go test -bench=. -benchmem

build: ## 编译API服务
	cd api && go build -o bin/paddle-trace-api ./cmd/main.go

build-linux: ## 交叉编译Linux版本
	cd api && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-ldflags="-s -w" -o bin/paddle-trace-api-linux ./cmd/main.go

lint: ## 运行代码静态分析
	cd api && golangci-lint run ./...

clean: ## 清理编译产物和数据
	rm -rf api/bin/
	rm -rf data/
	docker compose down -v
	@echo "✅ 清理完成"

dev: ## 本地开发模式（仅启动基础设施）
	docker compose up -d xchain-node1 postgres redis
	@echo "✅ 基础设施已启动，可本地运行 go run ./api/cmd/main.go"
