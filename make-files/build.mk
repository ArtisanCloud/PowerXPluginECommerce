# =========================
# build.mk （完整版本）
# 支持：
#  - Host（被 PowerX 反代运行）：POWERX_PROXY=1, baseURL=/_p/<pluginId>/admin/
#  - Standalone（独立部署/本地预览）：POWERX_PROXY=0, baseURL=/
#  - 前后端构建、运行、打包、检查
# =========================

# ===== 基础信息 =====
PLUGIN_ID           ?= com.powerx.plugins.ecommerce
# 从 plugin.yaml 读取版本（若失败则默认 0.1.0）
VERSION             ?= $(shell awk -F': *' '/^version:/ {print $$2; exit}' plugin.yaml 2>/dev/null || echo "0.1.0")
PLATFORM            ?= host
TARGET_ARCH         ?= amd64

# ===== 目录结构（可按项目调整）=====
# 后端代码在仓库根；如你的 cmd/plugin 在 repo/cmd/plugin，请保持 BACKEND_DIR = .
BACKEND_DIR         ?= .
BUILD_DIR           ?= $(BACKEND_DIR)/bin
ABS_BACKEND_DIR     := $(abspath $(BACKEND_DIR))
ABS_BUILD_DIR       := $(abspath $(BUILD_DIR))
GO_BUILD_CACHE     ?= $(abspath $(BACKEND_DIR)/.cache/go-build)

FRONTEND_DIR        ?= web-admin
FRONTEND_OUTPUT     ?= $(FRONTEND_DIR)/.output

# Dist（install/local 用）
DIST_ROOT           ?= dist
DIST_DIR            ?= $(DIST_ROOT)/$(VERSION)
DIST_BACKEND_BIN    ?= $(DIST_DIR)/backend/bin
DIST_WEBADMIN_DIR   ?= $(DIST_DIR)/web-admin
DIST_WEBADMIN_OUTPUT?= $(DIST_WEBADMIN_DIR)/.output
DIST_COMPACT        ?= auto
DIST_COMPACT_MAX_FILES ?= 900
DIST_VERIFY         ?= 1

# Release（完整发布包）
RELEASE_ROOT        ?= target
RELEASE_DIR         ?= $(RELEASE_ROOT)/$(VERSION)
RELEASE_BACKEND_BIN ?= $(RELEASE_DIR)/backend/bin
RELEASE_WEBADMIN_DIR?= $(RELEASE_DIR)/web-admin
RELEASE_WEBADMIN_OUTPUT ?= $(RELEASE_WEBADMIN_DIR)/.output

# ===== URL / 端口 =====
# Host 构建时写入到前端 baseURL
POWERX_ADMIN_BASE   ?= /_p/$(PLUGIN_ID)/admin/
HOST_PORT           ?= 4100                       # 运行 Host 产物时的本地端口
STANDALONE_PORT     ?= 4200                       # 运行 Standalone 产物时的本地端口
CHECK_PORT          ?= 4999                       # 临时检查端口（不要和上面冲突）

# ===== Go 构建（如不需要可删）=====
.PHONY: build
build: ## 构建后端（本机平台）
	@echo "==> 构建后端二进制（本机平台）..."
	@mkdir -p $(ABS_BUILD_DIR)
	@mkdir -p $(GO_BUILD_CACHE)
	@rm -f $(ABS_BUILD_DIR)/plugin $(ABS_BUILD_DIR)/migrate
	GOCACHE=$(GO_BUILD_CACHE) go build -C $(ABS_BACKEND_DIR) -o $(ABS_BUILD_DIR)/plugin ./cmd/plugin
	@if [ -d "$(ABS_BACKEND_DIR)/cmd/database" ]; then \
	  echo "   构建 migrate（如存在）..."; \
	  GOCACHE=$(GO_BUILD_CACHE) go build -C $(ABS_BACKEND_DIR) -o $(ABS_BUILD_DIR)/migrate ./cmd/database; \
	else \
	  echo "   跳过 migrate（未找到 cmd/database）"; \
	fi

.PHONY: build-linux
build-linux: ## 构建后端（Linux amd64）
	@echo "==> 构建后端二进制（Linux/$(TARGET_ARCH)）..."
	@mkdir -p $(ABS_BUILD_DIR)
	@mkdir -p $(GO_BUILD_CACHE)
	@rm -f $(ABS_BUILD_DIR)/plugin $(ABS_BUILD_DIR)/migrate
	GOOS=linux GOARCH=$(TARGET_ARCH) GOCACHE=$(GO_BUILD_CACHE) go build -C $(ABS_BACKEND_DIR) -o $(ABS_BUILD_DIR)/plugin ./cmd/plugin
	@if [ -d "$(ABS_BACKEND_DIR)/cmd/database" ]; then \
	  echo "   构建 migrate（Linux/$(TARGET_ARCH)）..."; \
	  GOOS=linux GOARCH=$(TARGET_ARCH) GOCACHE=$(GO_BUILD_CACHE) go build -C $(ABS_BACKEND_DIR) -o $(ABS_BUILD_DIR)/migrate ./cmd/database; \
	else \
	  echo "   跳过 migrate（未找到 cmd/database）"; \
	fi

.PHONY: dist-backend
dist-backend:
	@if [ "$(PLATFORM)" = "linux" ]; then \
	  $(MAKE) --no-print-directory build-linux BUILD_DIR="$(BUILD_DIR)" TARGET_ARCH="$(TARGET_ARCH)" GO_BUILD_CACHE="$(GO_BUILD_CACHE)"; \
	else \
	  $(MAKE) --no-print-directory build BUILD_DIR="$(BUILD_DIR)" GO_BUILD_CACHE="$(GO_BUILD_CACHE)"; \
	fi
	@if [ ! -s "$(BUILD_DIR)/plugin" ]; then \
	  echo "❌ dist-backend 失败：未生成有效后端二进制 $(BUILD_DIR)/plugin"; \
	  exit 1; \
	fi

# ===== 前端构建（Host / 被 PowerX 反代）=====
.PHONY: frontend-build
frontend-build: ## 构建 Host 包（POWERX_PROXY=1, baseURL=$(POWERX_ADMIN_BASE)）
	@echo "==> 构建 web-admin（Host 包） POWERX_PROXY=1 baseURL=$(POWERX_ADMIN_BASE)"
	cd $(FRONTEND_DIR) && \
	  POWERX_PROXY=1 \
	  NUXT_PUBLIC_INSIDE_POWERX=1 \
	  POWERX_PLUGIN_ID="$(PLUGIN_ID)" \
	  POWERX_PLUGIN_VERSION="$(VERSION)" \
	  NUXT_PUBLIC_POWERX_PLUGIN_ID="$(PLUGIN_ID)" \
	  NUXT_PUBLIC_POWERX_PLUGIN_VERSION="$(VERSION)" \
	  NUXT_PUBLIC_API_BASE= \
	  NUXT_PUBLIC_API_PREFIX= \
	  POWERX_ADMIN_BASE="$(POWERX_ADMIN_BASE)" \
	  NODE_ENV=production \
	  npm run build

# ===== 前端构建（Standalone / 独立部署）=====
.PHONY: frontend-build-standalone
frontend-build-standalone: ## 构建 Standalone 包（POWERX_PROXY=0, baseURL=/）
	@echo "==> 构建 web-admin（Standalone 包） POWERX_PROXY=0 baseURL=/"
	cd $(FRONTEND_DIR) && \
	  POWERX_PROXY=0 \
	  NUXT_PUBLIC_INSIDE_POWERX=0 \
	  POWERX_PLUGIN_ID="$(PLUGIN_ID)" \
	  POWERX_PLUGIN_VERSION="$(VERSION)" \
	  NUXT_PUBLIC_POWERX_PLUGIN_ID="$(PLUGIN_ID)" \
	  NUXT_PUBLIC_POWERX_PLUGIN_VERSION="$(VERSION)" \
	  NODE_ENV=production \
	  npm run build

# ===== 运行已编译的前端产物（Host）=====
.PHONY: run-frontend
run-frontend: ## 启动 Host 产物（默认端口 $(HOST_PORT)）
	@if [ ! -f "$(FRONTEND_OUTPUT)/server/index.mjs" ]; then \
	  echo "❌ 未找到 $(FRONTEND_OUTPUT)/server/index.mjs"; \
	  echo "   请先执行: make frontend-build"; \
	  exit 1; \
	fi
	@if [ -z "$(HOST_PORT)" ]; then echo "❌ HOST_PORT 为空"; exit 1; fi
	@echo "==> 启动 Host 产物： http://127.0.0.1:$(HOST_PORT)$(POWERX_ADMIN_BASE)"
	cd $(FRONTEND_OUTPUT) && PORT=$(HOST_PORT) NODE_ENV=production node server/index.mjs

# ===== 运行已编译的前端产物（Standalone）=====
.PHONY: run-frontend-standalone
run-frontend-standalone: ## 启动 Standalone 产物（默认端口 $(STANDALONE_PORT)）
	@if [ ! -f "$(FRONTEND_OUTPUT)/server/index.mjs" ]; then \
	  echo "❌ 未找到 $(FRONTEND_OUTPUT)/server/index.mjs"; \
	  echo "   请先执行: make frontend-build-standalone"; \
	  exit 1; \
	fi
	@if [ -z "$(STANDALONE_PORT)" ]; then echo "❌ STANDALONE_PORT 为空"; exit 1; fi
	@echo "==> 启动 Standalone 产物： http://127.0.0.1:$(STANDALONE_PORT)/"
	cd $(FRONTEND_OUTPUT) && PORT=$(STANDALONE_PORT) NODE_ENV=production node server/index.mjs

# ===== 校验产物中的 baseURL（Host 构建）=====
.PHONY: check-base-host
check-base-host: frontend-build ## 构建后临时起 Nitro，抓首页里的 app.baseURL
	@echo "==> 检查 Host 产物 baseURL..."
	@if [ ! -f "$(FRONTEND_OUTPUT)/server/index.mjs" ]; then \
	  echo "❌ 未找到 $(FRONTEND_OUTPUT)/server/index.mjs"; exit 1; fi
	cd $(FRONTEND_OUTPUT) && \
	  PORT=$(CHECK_PORT) NODE_ENV=production node server/index.mjs & echo $$! > .nuxt_pid; \
	  for i in `seq 1 40`; do \
	    sleep 0.25; curl -fsS "http://127.0.0.1:$(CHECK_PORT)$(POWERX_ADMIN_BASE)" >/dev/null 2>&1 && break; \
	    if [ $$i -eq 40 ]; then echo "Nitro 未就绪"; kill $$(cat .nuxt_pid) 2>/dev/null || true; exit 1; fi; \
	  done; \
	  echo -n "HTML 中的 "; \
	  curl -s "http://127.0.0.1:$(CHECK_PORT)$(POWERX_ADMIN_BASE)" | grep -o 'app:{baseURL:"[^"]*"}' | head -1 || true; \
	  kill $$(cat .nuxt_pid) 2>/dev/null || true; rm -f .nuxt_pid

# ===== 校验产物中的 baseURL（Standalone 构建）=====
.PHONY: check-base-standalone
check-base-standalone: frontend-build-standalone
	@echo "==> 检查 Standalone 产物 baseURL..."
	@if [ ! -f "$(FRONTEND_OUTPUT)/server/index.mjs" ]; then \
	  echo "❌ 未找到 $(FRONTEND_OUTPUT)/server/index.mjs"; exit 1; fi
	cd $(FRONTEND_OUTPUT) && \
	  PORT=$(CHECK_PORT) NODE_ENV=production node server/index.mjs & echo $$! > .nuxt_pid; \
	  for i in `seq 1 40`; do \
	    sleep 0.25; curl -fsS "http://127.0.0.1:$(CHECK_PORT)/" >/dev/null 2>&1 && break; \
	    if [ $$i -eq 40 ]; then echo "Nitro 未就绪"; kill $$(cat .nuxt_pid) 2>/dev/null || true; exit 1; fi; \
	  done; \
	  echo -n "HTML 中的 "; \
	  curl -s "http://127.0.0.1:$(CHECK_PORT)/" | grep -o 'app:{baseURL:"[^"]*"}' | head -1 || true; \
	  kill $$(cat .nuxt_pid) 2>/dev/null || true; rm -f .nuxt_pid

# ===== 生成 dist（目录安装包，给 PowerX 的 install/local 用）=====
.PHONY: dist
dist: plugin-yaml-check dist-backend frontend-build
	@echo "==> 生成 dist 安装包目录：$(DIST_DIR)"
	@rm -rf $(DIST_DIR)
	@mkdir -p $(DIST_BACKEND_BIN) $(DIST_WEBADMIN_OUTPUT)
	@echo "写入插件清单 -> $(DIST_DIR)/plugin.yaml (version=$(VERSION))"
	@awk -v ver="$(VERSION)" 'BEGIN{patched=0} /^[[:space:]]*version:[[:space:]]*/ && !patched {print "version: " ver; patched=1; next} {print} END{if(!patched) print "version: " ver}' plugin.yaml > $(DIST_DIR)/plugin.yaml
	@cp $(BUILD_DIR)/plugin $(DIST_BACKEND_BIN)/
	@if [ -f "$(BUILD_DIR)/migrate" ]; then cp $(BUILD_DIR)/migrate $(DIST_BACKEND_BIN)/; fi
	@chmod 0755 $(DIST_BACKEND_BIN)/plugin
	@if [ -f "$(DIST_BACKEND_BIN)/migrate" ]; then chmod 0755 $(DIST_BACKEND_BIN)/migrate; fi
	@if [ -d "backend/etc" ]; then \
	  mkdir -p $(DIST_DIR)/backend/etc; \
	  cp -R backend/etc/. $(DIST_DIR)/backend/etc/; \
	fi
	@if [ -d "config" ]; then \
	  mkdir -p $(DIST_DIR)/config; \
	  cp -R config/. $(DIST_DIR)/config/; \
	fi
	@if [ -d "$(FRONTEND_OUTPUT)" ] && [ -n "$$(ls -A $(FRONTEND_OUTPUT) 2>/dev/null)" ]; then \
	  echo "复制前端构建产物 -> $(DIST_WEBADMIN_OUTPUT)"; \
	  cp -R $(FRONTEND_OUTPUT)/. $(DIST_WEBADMIN_OUTPUT)/; \
	else \
	  echo "⚠️  未找到前端构建产物：$(FRONTEND_OUTPUT)"; \
	fi
	@if [ -d "$(FRONTEND_DIR)/i18n" ]; then \
	  mkdir -p $(DIST_WEBADMIN_DIR)/i18n; \
	  cp -R $(FRONTEND_DIR)/i18n/. $(DIST_WEBADMIN_DIR)/i18n/; \
	fi
	@if [ -d "plugin.d" ]; then \
	  mkdir -p $(DIST_DIR)/plugin.d; \
	  cp -R plugin.d/. $(DIST_DIR)/plugin.d/; \
	fi
	@if [ -d "skills" ]; then \
	  mkdir -p $(DIST_DIR)/skills; \
	  cp -R skills/. $(DIST_DIR)/skills/; \
	fi
	@if [ -d "contracts" ]; then \
	  mkdir -p $(DIST_DIR)/contracts; \
	  cp -R contracts/. $(DIST_DIR)/contracts/; \
	fi
	@if [ -f README.md ]; then cp README.md $(DIST_DIR)/; fi
		@if [ "$(DIST_VERIFY)" = "1" ]; then \
		  $(MAKE) --no-print-directory dist-verify DIST_DIR="$(DIST_DIR)" DIST_BACKEND_BIN="$(DIST_BACKEND_BIN)" DIST_WEBADMIN_OUTPUT="$(DIST_WEBADMIN_OUTPUT)"; \
		fi
		@if [ "$(DIST_COMPACT)" != "0" ]; then \
		  FILE_COUNT=$$(find "$(DIST_DIR)" -type f | wc -l | tr -d ' '); \
		  if [ "$(DIST_COMPACT)" = "1" ] || { [ "$(DIST_COMPACT)" = "auto" ] && [ "$$FILE_COUNT" -gt "$(DIST_COMPACT_MAX_FILES)" ]; }; then \
		    echo "==> dist 文件数 $$FILE_COUNT 超过目录上传阈值，压缩为 PowerX package.tar.gz"; \
		    TMP_DIR=$$(mktemp -d); \
		    mkdir -p "$$TMP_DIR/payload"; \
		    find "$(DIST_DIR)" -mindepth 1 -maxdepth 1 -exec mv {} "$$TMP_DIR/payload/" \; ; \
		    (cd "$$TMP_DIR" && tar -czf "$(abspath $(DIST_DIR))/package.tar.gz" payload); \
		    tar -tzf "$(DIST_DIR)/package.tar.gz" payload/plugin.yaml >/dev/null || { echo "❌ package.tar.gz 缺少 payload/plugin.yaml"; rm -rf "$$TMP_DIR"; exit 1; }; \
		    rm -rf "$$TMP_DIR"; \
		    echo "✅ compact dist ready: $(DIST_DIR)/package.tar.gz"; \
		  fi; \
		fi

.PHONY: dist-verify
dist-verify: ## 校验 dist 安装包必需结构与核心权限注册
	@echo "==> dist 验证（$(DIST_DIR)）"
	@test -f "$(DIST_DIR)/plugin.yaml" || { echo "❌ 缺少 $(DIST_DIR)/plugin.yaml"; exit 1; }
	@test -s "$(DIST_BACKEND_BIN)/plugin" || { echo "❌ 缺少后端二进制 $(DIST_BACKEND_BIN)/plugin"; exit 1; }
	@test -s "$(DIST_BACKEND_BIN)/migrate" || { echo "❌ 缺少迁移二进制 $(DIST_BACKEND_BIN)/migrate"; exit 1; }
	@test -x "$(DIST_BACKEND_BIN)/plugin" || { echo "❌ 后端二进制不可执行 $(DIST_BACKEND_BIN)/plugin"; exit 1; }
	@test -x "$(DIST_BACKEND_BIN)/migrate" || { echo "❌ 迁移二进制不可执行 $(DIST_BACKEND_BIN)/migrate"; exit 1; }
	@awk '/^[[:space:]]*runtime:[[:space:]]*$$/{in_runtime=1; next} in_runtime && /^[^[:space:]]/ {in_runtime=0} in_runtime && /^[[:space:]]*entry:[[:space:]]*backend\/bin\/plugin[[:space:]]*$$/ {found=1} END{exit found?0:1}' "$(DIST_DIR)/plugin.yaml" || { echo "❌ plugin.yaml runtime.entry 必须是 backend/bin/plugin"; exit 1; }
	@rg -q 'POWERX_BIND_ADDR:[[:space:]]*":__POWERX_DYNAMIC_PORT__"' "$(DIST_DIR)/plugin.yaml" || { echo "❌ plugin.yaml 必须使用 POWERX_BIND_ADDR 动态端口占位符"; exit 1; }
	@rg -q 'POWERX_PLUGIN_REGISTRATION_MODE:[[:space:]]*installed' "$(DIST_DIR)/plugin.yaml" || { echo "❌ plugin.yaml 必须设置 POWERX_PLUGIN_REGISTRATION_MODE: installed"; exit 1; }
	@awk '/^[[:space:]]*backend:[[:space:]]*$$/{in_backend=1; next} in_backend && /^[^[:space:]]/ {in_backend=0} in_backend && /^[[:space:]]*port:[[:space:]]*0[[:space:]]*$$/ {found=1} END{exit found?0:1}' "$(DIST_DIR)/plugin.yaml" || { echo "❌ plugin.yaml backend.port 必须是 0"; exit 1; }
	@! rg -q 'port:[[:space:]]*(8078|8086)[[:space:]]*$$' "$(DIST_DIR)/plugin.yaml" || { echo "❌ plugin.yaml 不应固化旧 backend port"; exit 1; }
	@awk '/^[[:space:]]*migrations:[[:space:]]*$$/{in_migrations=1; next} in_migrations && /^[^[:space:]]/ {in_migrations=0} in_migrations && /^[[:space:]]*entry:[[:space:]]*backend\/bin\/migrate[[:space:]]*$$/ {found=1} END{exit found?0:1}' "$(DIST_DIR)/plugin.yaml" || { echo "❌ plugin.yaml migrations.entry 必须是 backend/bin/migrate"; exit 1; }
	@test -f "$(DIST_WEBADMIN_OUTPUT)/server/index.mjs" || { echo "❌ 缺少前端服务入口 $(DIST_WEBADMIN_OUTPUT)/server/index.mjs"; exit 1; }
	@test -f "$(DIST_WEBADMIN_OUTPUT)/public/icon.svg" || { echo "❌ 缺少插件市场图标 $(DIST_WEBADMIN_OUTPUT)/public/icon.svg"; exit 1; }
	@awk '/^[[:space:]]*metadata:[[:space:]]*$$/{in_metadata=1; next} in_metadata && /^[^[:space:]]/ {in_metadata=0} in_metadata && /^[[:space:]]*icon:[[:space:]]*icon\.svg[[:space:]]*$$/ {found=1} END{exit found?0:1}' "$(DIST_DIR)/plugin.yaml" || { echo "❌ plugin.yaml metadata.icon 必须是 icon.svg"; exit 1; }
	@test -f "$(DIST_WEBADMIN_DIR)/i18n/zh-CN/menus.json" || { echo "❌ 缺少宿主菜单中文多语言 $(DIST_WEBADMIN_DIR)/i18n/zh-CN/menus.json"; exit 1; }
	@test -f "$(DIST_WEBADMIN_DIR)/i18n/en/menus.json" || { echo "❌ 缺少宿主菜单英文多语言 $(DIST_WEBADMIN_DIR)/i18n/en/menus.json"; exit 1; }
	@test -f "$(DIST_DIR)/config/event_fabric.yaml" || { echo "❌ 缺少 $(DIST_DIR)/config/event_fabric.yaml"; exit 1; }
	@for f in plugin.d/capabilities.yaml plugin.d/exposure.yaml plugin.d/rbac.yaml; do \
	  test -f "$(DIST_DIR)/$$f" || { echo "❌ 缺少 $(DIST_DIR)/$$f"; exit 1; }; \
	done
	@test -d "$(DIST_DIR)/skills" || { echo "❌ 缺少 $(DIST_DIR)/skills"; exit 1; }
	@find "$(DIST_DIR)/skills" -mindepth 2 -maxdepth 2 -name SKILL.md -print -quit | rg -q . || { echo "❌ skills 目录没有标准 SKILL.md 包"; exit 1; }
	@cd $(BACKEND_DIR) && GOCACHE=$(GO_BUILD_CACHE) go run ./cmd/skillcheck --root "$(abspath $(DIST_DIR))"
	@test -d "$(DIST_DIR)/contracts/capabilities" || { echo "❌ 缺少 $(DIST_DIR)/contracts/capabilities"; exit 1; }
	@for key in capabilities exposure rbac; do \
	  awk -v key="$$key" '/^[[:space:]]*catalogs:[[:space:]]*$$/{in_catalogs=1; next} in_catalogs && /^[^[:space:]]/ {in_catalogs=0} in_catalogs && $$0 ~ "^[[:space:]]*" key ":" {found=1} END{exit found?0:1}' "$(DIST_DIR)/plugin.yaml" || { echo "❌ plugin.yaml 缺少 catalogs.$$key"; exit 1; }; \
	done
	@rg -q "resource: com.powerx.plugins.ecommerce:pricing.coupon.template" "$(DIST_DIR)/plugin.d/rbac.yaml" || { echo "❌ rbac 缺少优惠券模板权限"; exit 1; }
	@rg -q "path: /admin/coupons/templates" "$(DIST_DIR)/plugin.d/rbac.yaml" || { echo "❌ rbac 缺少优惠券模板接口路径"; exit 1; }
	@rg -q "path: /admin/coupons/usage-logs" "$(DIST_DIR)/plugin.d/rbac.yaml" || { echo "❌ rbac 缺少优惠券流水接口路径"; exit 1; }
	@rg -q "path: /admin/pricing/pricebooks" "$(DIST_DIR)/plugin.d/rbac.yaml" || { echo "❌ rbac 缺少价格手册接口路径"; exit 1; }
	@echo "✅ dist 验证通过"

# ===== 生成 release（完整发布包）=====
.PHONY: release
release: build frontend-build
	@echo "==> 生成 release 发布目录：$(RELEASE_DIR)"
	@rm -rf $(RELEASE_DIR)
	@mkdir -p $(RELEASE_BACKEND_BIN) $(RELEASE_WEBADMIN_OUTPUT)
	@echo "写入插件清单 -> $(RELEASE_DIR)/plugin.yaml (version=$(VERSION))"
	@awk -v ver="$(VERSION)" 'BEGIN{patched=0} /^[[:space:]]*version:[[:space:]]*/ && !patched {print "version: " ver; patched=1; next} {print} END{if(!patched) print "version: " ver}' plugin.yaml > $(RELEASE_DIR)/plugin.yaml
	@cp $(BUILD_DIR)/plugin $(RELEASE_BACKEND_BIN)/
	@if [ -f "$(BUILD_DIR)/migrate" ]; then cp $(BUILD_DIR)/migrate $(RELEASE_BACKEND_BIN)/; fi
	@chmod 0755 $(RELEASE_BACKEND_BIN)/plugin
	@if [ -f "$(RELEASE_BACKEND_BIN)/migrate" ]; then chmod 0755 $(RELEASE_BACKEND_BIN)/migrate; fi
	@if [ -d "backend/etc" ]; then \
	  mkdir -p $(RELEASE_DIR)/backend/etc; \
	  cp -R backend/etc/. $(RELEASE_DIR)/backend/etc/; \
	fi
	@if [ -d "config" ]; then \
	  mkdir -p $(RELEASE_DIR)/config; \
	  cp -R config/. $(RELEASE_DIR)/config/; \
	fi
	@cp -R $(FRONTEND_OUTPUT)/. $(RELEASE_WEBADMIN_OUTPUT)/
	@if [ -d "$(FRONTEND_DIR)/i18n" ]; then \
	  mkdir -p $(RELEASE_WEBADMIN_DIR)/i18n; \
	  cp -R $(FRONTEND_DIR)/i18n/. $(RELEASE_WEBADMIN_DIR)/i18n/; \
	fi
	@if [ -d "plugin.d" ]; then \
	  mkdir -p $(RELEASE_DIR)/plugin.d; \
	  cp -R plugin.d/. $(RELEASE_DIR)/plugin.d/; \
	fi
	@if [ -d "skills" ]; then \
	  mkdir -p $(RELEASE_DIR)/skills; \
	  cp -R skills/. $(RELEASE_DIR)/skills/; \
	fi
	@if [ -d "contracts" ]; then \
	  mkdir -p $(RELEASE_DIR)/contracts; \
	  cp -R contracts/. $(RELEASE_DIR)/contracts/; \
	fi
	@if [ -f README.md ]; then cp README.md $(RELEASE_DIR)/; fi
	@rg -q 'POWERX_BIND_ADDR:[[:space:]]*":__POWERX_DYNAMIC_PORT__"' "$(RELEASE_DIR)/plugin.yaml" || { echo "❌ release 验证失败：plugin.yaml 必须使用 POWERX_BIND_ADDR 动态端口占位符"; exit 1; }
	@rg -q 'POWERX_PLUGIN_REGISTRATION_MODE:[[:space:]]*installed' "$(RELEASE_DIR)/plugin.yaml" || { echo "❌ release 验证失败：plugin.yaml 必须设置 POWERX_PLUGIN_REGISTRATION_MODE: installed"; exit 1; }
	@awk '/^[[:space:]]*backend:[[:space:]]*$$/{in_backend=1; next} in_backend && /^[^[:space:]]/ {in_backend=0} in_backend && /^[[:space:]]*port:[[:space:]]*0[[:space:]]*$$/ {found=1} END{exit found?0:1}' "$(RELEASE_DIR)/plugin.yaml" || { echo "❌ release 验证失败：plugin.yaml backend.port 必须是 0"; exit 1; }
	@! rg -q 'port:[[:space:]]*(8078|8086)[[:space:]]*$$' "$(RELEASE_DIR)/plugin.yaml" || { echo "❌ release 验证失败：plugin.yaml 不应固化旧 backend port"; exit 1; }
	@test -d "$(RELEASE_DIR)/skills" || { echo "❌ release 验证失败：缺少 $(RELEASE_DIR)/skills"; exit 1; }
	@find "$(RELEASE_DIR)/skills" -mindepth 2 -maxdepth 2 -name SKILL.md -print -quit | rg -q . || { echo "❌ release 验证失败：skills 目录没有标准 SKILL.md 包"; exit 1; }

# ===== 打包 zip =====
.PHONY: package
package: dist
	@echo "==> 打包 dist 为 zip（install/local 用）..."
	@rm -f $(PLUGIN_ID)-$(VERSION).zip
	@cd $(DIST_ROOT) && zip -r ../$(PLUGIN_ID)-$(VERSION).zip $(VERSION)
	@echo "✅ 输出：$(PLUGIN_ID)-$(VERSION).zip"

.PHONY: package-release
package-release: release
	@echo "==> 打包 release 为 zip（发布包）..."
	@rm -f $(PLUGIN_ID)-$(VERSION)-release.zip
	@cd $(RELEASE_ROOT) && zip -r ../$(PLUGIN_ID)-$(VERSION)-release.zip $(VERSION)
	@echo "✅ 输出：$(PLUGIN_ID)-$(VERSION)-release.zip"

# ===== 清理 =====
.PHONY: clean
clean:
	@echo "==> 清理 build 产物..."
	@rm -rf $(BUILD_DIR)

.PHONY: dist-clean
dist-clean:
	@echo "==> 清理 dist..."
	@rm -rf $(DIST_ROOT)

.PHONY: release-clean
release-clean:
	@echo "==> 清理 target..."
	@rm -rf $(RELEASE_ROOT)
