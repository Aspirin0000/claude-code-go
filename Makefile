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
	@echo "🔨 构建项目..."
	@go build ./...
	@echo "✅ 构建成功"

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
	@echo "可用目标:"
	@echo "  all                  - 检查简化并构建"
	@echo "  check-simplification - 检查是否有简化代码"
	@echo "  build               - 构建项目"
	@echo "  test                - 运行测试"
	@echo "  check               - 完整检查（简化+构建+测试）"
	@echo "  clean               - 清理生成文件"
	@echo "  deps                - 安装依赖"
	@echo "  fmt                 - 格式化代码"
	@echo "  help                - 显示帮助"
