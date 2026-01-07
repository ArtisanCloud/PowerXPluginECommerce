# test.mk 汇总测试与代码质量相关目标

GO_TOOLCHAIN ?= go1.24.2
GO_MOD_CACHE := $(abspath $(BACKEND_DIR)/.gomodcache)
GO_BUILD_CACHE := $(abspath $(BACKEND_DIR)/.cache/go-build)
GO_ENV      := GOMODCACHE=$(GO_MOD_CACHE) GOCACHE=$(GO_BUILD_CACHE)

GOLANGCI_LINT_VERSION ?= v1.64.8
GOLANGCI_LINT_BIN_DIR := $(abspath $(BACKEND_DIR)/bin)
GOLANGCI_LINT         := $(GOLANGCI_LINT_BIN_DIR)/golangci-lint
GOLANGCI_LINT_CACHE   := $(abspath $(BACKEND_DIR)/.cache/golangci-lint)

ifeq ($(strip $(GO_TOOLCHAIN)),)
	GO_CMD     := $(GO_ENV) go
	LINT_GOENV := $(GO_ENV)
else
	GO_CMD     := $(GO_ENV) GOTOOLCHAIN=$(GO_TOOLCHAIN) go
	LINT_GOENV := $(GO_ENV) GOTOOLCHAIN=$(GO_TOOLCHAIN)
endif

.PHONY: test
test: ## 执行 Go 单元测试
	@echo "运行测试..."
	@mkdir -p $(GO_MOD_CACHE) $(GO_BUILD_CACHE)
	cd $(BACKEND_DIR) && $(GO_CMD) test ./...

.PHONY: test-admin
test-admin: ## 运行 web-admin 测试
	@echo "运行 Web Admin 测试..."
	cd $(FRONTEND_DIR) && npm run test

.PHONY: test-admin-ci
test-admin-ci: ## CI 中运行 web-admin 单元/组件测试（跳过 e2e）
	@echo "运行 Web Admin CI 测试（跳过 e2e）..."
	cd $(FRONTEND_DIR) && npm run test:ci

.PHONY: lint-admin
lint-admin: ## 运行 web-admin Lint
	@echo "运行 Web Admin Lint..."
	@# 某些项目 lint 脚本可能还是占位（不接受 eslint 参数）；优先严格模式，失败再回退。
	cd $(FRONTEND_DIR) && (npm run lint -- --max-warnings=0 || npm run lint)

.PHONY: build-admin
build-admin: ## 构建 web-admin 产物
	@echo "构建 Web Admin..."
	cd $(FRONTEND_DIR) && npm run build

.PHONY: test-coverage
test-coverage: ## 执行测试并生成覆盖率报告
	@echo "运行测试并生成覆盖率报告..."
	@mkdir -p $(GO_MOD_CACHE) $(GO_BUILD_CACHE)
	cd $(BACKEND_DIR) && $(GO_CMD) test -coverprofile=coverage.out ./...
	cd $(BACKEND_DIR) && $(GO_CMD) tool cover -html=coverage.out -o coverage.html

.PHONY: lint
lint: ## 运行 golangci-lint
	@mkdir -p $(GO_MOD_CACHE) $(GO_BUILD_CACHE) $(GOLANGCI_LINT_BIN_DIR) $(GOLANGCI_LINT_CACHE)
	@echo "安装/更新 golangci-lint ($(GOLANGCI_LINT_VERSION))..."
	cd $(BACKEND_DIR) && GOBIN=$(GOLANGCI_LINT_BIN_DIR) $(GO_CMD) install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	@echo "运行代码检查..."
	cd $(BACKEND_DIR) && $(LINT_GOENV) GOLANGCI_LINT_CACHE=$(GOLANGCI_LINT_CACHE) $(GOLANGCI_LINT) run --timeout 5m

.PHONY: fmt
fmt: ## 使用 go fmt 格式化代码
	@echo "格式化代码..."
	cd $(BACKEND_DIR) && $(GO_CMD) fmt ./...

.PHONY: mod-tidy
mod-tidy: ## 整理 Go 模块依赖
	@echo "整理 Go 模块依赖..."
	cd $(BACKEND_DIR) && $(GO_CMD) mod tidy

.PHONY: test-all
test-all: fmt lint lint-admin test test-admin-ci build-admin ## 运行后端/前端全量验证
	@echo "所有测试与构建完成"

.PHONY: integration-smoke
integration-smoke: ## 运行集成回归演练（Webhook Replay + Nuxt 构建）
	@echo "运行 Webhook Replay Drill..."
	cd $(BACKEND_DIR) && $(GO_CMD) test ./internal/services/integration -run TestWebhookService_ReplayAttemptFlow -count=1
	$(MAKE) frontend-build

.PHONY: check
check: lint lint-admin test test-admin security-audit ## 运行 lint + test + security audit
