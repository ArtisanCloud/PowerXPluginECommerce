# ci.mk 汇总 CI 相关快捷目标

.PHONY: ci-integration
ci-integration: ## 运行集成特性 CI 流程
	scripts/ci/integration.sh

.PHONY: ci-all
ci-all: ## 运行完整 CI 流程（act 优先，否则本地执行 test-all）
	@if command -v act >/dev/null 2>&1 ; then \
		echo "检测到 act，使用 GitHub Actions 模拟器执行 $(if $(ACT_OPTS),$(ACT_OPTS),默认配置)"; \
		if [ -n "$(ACT_OPTS)" ]; then \
			act $(ACT_OPTS); \
		else \
			act; \
		fi; \
	else \
		echo "act 未安装，直接执行 make test-all"; \
		$(MAKE) test-all; \
	fi
