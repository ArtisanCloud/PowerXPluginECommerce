COMPAT_CONFIG ?= contracts/compatibility.yaml
COMPAT_OUTPUT ?= build/compat
COMPAT_SCRIPT := $(abspath scripts/check-compatibility.mjs)
COMPAT_NODE ?= node

.PHONY: check-compat
check-compat:
	@echo "[compat] Generating compatibility report using $(COMPAT_CONFIG)"
	@mkdir -p $(COMPAT_OUTPUT)
	@$(COMPAT_NODE) $(COMPAT_SCRIPT) --config "$(abspath $(COMPAT_CONFIG))" --out "$(abspath $(COMPAT_OUTPUT))"

.PHONY: check-powerx-align
check-powerx-align:
	@echo "[align] checking deprecated env vars..."
	@if rg -n "POWERX_RBAC_DELEGATE" . -S --glob '!make-files/compat.mk'; then \
		echo "[align] FAIL: deprecated POWERX_RBAC_DELEGATE still exists"; \
		exit 1; \
	else \
		echo "[align] PASS: POWERX_RBAC_DELEGATE not found"; \
	fi
	@echo "[align] checking proxy/apikey coupling..."
	@if rg -n "POWERX_PROXY.*apikey|apikey.*POWERX_PROXY" backend -S; then \
		echo "[align] FAIL: found POWERX_PROXY/apikey coupling"; \
		exit 1; \
	else \
		echo "[align] PASS: no POWERX_PROXY/apikey coupling"; \
	fi
