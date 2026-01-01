# ci.mk 汇总 CI 相关快捷目标

.PHONY: ci-integration
ci-integration: ## 运行集成特性 CI 流程
	scripts/ci/integration.sh

.PHONY: ci-all
ci-all: ## 运行完整 CI 流程（ACT_OPTS/act 存在则走 act，否则本地 test-all）
	@if [ -n "$(ACT_OPTS)" ]; then \
		$(MAKE) ci-act ACT_OPTS="$(ACT_OPTS)"; \
	elif command -v act >/dev/null 2>&1 ; then \
		$(MAKE) ci-act; \
	else \
		echo "执行本地 test-all，与 PowerXPlugin 流程保持一致..."; \
		$(MAKE) test-all; \
	fi

.PHONY: ci-act
ci-act: ## 使用 act 在容器内跑 GitHub Actions（支持 ACT_OPTS）
	@if ! command -v act >/dev/null 2>&1 ; then \
		echo "未找到 act，请先安装：https://github.com/nektos/act"; \
		exit 1; \
	fi
	@echo "使用 act 执行 .github/workflows/ci.yml (job=build)..."
	@echo "ACT_OPTS=$(ACT_OPTS)"
	@act -W .github/workflows/ci.yml -j build $(ACT_OPTS)
