SYNC_LIFECYCLE_SRC := docs/lifecycle
SYNC_LIFECYCLE_DEST := docs/integration/01_plugin_lifecycle
SPEC_CONTRACT_DIR := specs/003-channel-master/contracts
SPECTRAL ?= npx --yes @stoplight/spectral-cli

.PHONY: sync-lifecycle-docs
sync-lifecycle-docs:
	@echo "[docs] Syncing $(SYNC_LIFECYCLE_SRC) -> $(SYNC_LIFECYCLE_DEST)"
	@mkdir -p $(SYNC_LIFECYCLE_DEST)
	@rsync -a $(SYNC_LIFECYCLE_SRC)/ $(SYNC_LIFECYCLE_DEST)/

.PHONY: lint-contracts
lint-contracts:
	@echo "[docs] Linting OpenAPI contracts via Spectral"
	@$(SPECTRAL) lint $(SPEC_CONTRACT_DIR)/*.yaml
