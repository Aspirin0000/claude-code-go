.PHONY: all build check-simplification test clean

# 默认目标
all: check-simplification build

# 检查简化行为
check-simplification:
	@echo "🔍 检查简化行为..."
	@go run scripts/verify_no_simplification.go
	@echo "✅ 零简化政策验证通过"

# 构建项目
build: check-simplification
	@echo "🔨 Building project..."
	@go build ./...
	@go build -o claude ./cmd/claude
	@echo "✅ Build successful! Binary: ./claude"

# 安装到系统
install: build
	@echo "📦 Installing claude to /usr/local/bin..."
	@sudo cp claude /usr/local/bin/
	@echo "✅ Installed successfully! Run 'claude --help' to get started."

# 卸载
uninstall:
	@echo "🗑️  Removing claude from /usr/local/bin..."
	@sudo rm -f /usr/local/bin/claude
	@echo "✅ Uninstalled successfully!"

# 构建并运行
run: build
	@echo "🚀 Running..."
	@./claude

# 开发模式运行
dev:
	@echo "🚀 Running in dev mode..."
	@go run ./cmd/claude

# 运行测试
test:
	@echo "🧪 运行测试..."
	@go test ./... -v

# 清理生成文件
clean:
	@echo "🧹 清理..."
	@go clean
	@rm -f claude

# 安装依赖
deps:
	@echo "📦 安装依赖..."
	@go mod tidy

# 格式化代码
fmt:
	@echo "📝 格式化代码..."
	@go fmt ./...

# 运行静态检查
lint:
	@echo "🔍 静态检查..."
	@golangci-lint run || echo "请安装 golangci-lint: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"

# 完整检查（简化检查 + 构建 + 测试）
check: check-simplification build test
	@echo "✅ 所有检查通过"

# 帮助
help:
	@echo "Available targets:"
	@echo "  all                  - Check simplification and build"
	@echo "  check-simplification - Check for simplified code"
	@echo "  build               - Build the project"
	@echo "  install             - Install claude to /usr/local/bin"
	@echo "  uninstall           - Remove claude from /usr/local/bin"
	@echo "  test                - Run tests"
	@echo "  check               - Full check (simplification + build + test)"
	@echo "  clean               - Clean generated files"
	@echo "  deps                - Install dependencies"
	@echo "  fmt                 - Format code"
	@echo "  run                 - Build and run"
	@echo "  dev                 - Run with go run"
	@echo "  help                - Show this help"
