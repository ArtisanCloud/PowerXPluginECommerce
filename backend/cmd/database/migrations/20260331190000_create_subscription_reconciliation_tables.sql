-- 013-subscription-reconciliation foundational tables (part 1)
CREATE TABLE IF NOT EXISTS subscription_reconciliation_batches (
  id UUID PRIMARY KEY,
  tenant_uuid UUID NOT NULL,
  billing_cycle VARCHAR(32) NOT NULL,
  run_type VARCHAR(16) NOT NULL,
  expected_amount_minor BIGINT NOT NULL DEFAULT 0,
  actual_amount_minor BIGINT NOT NULL DEFAULT 0,
  delta_amount_minor BIGINT NOT NULL DEFAULT 0,
  delta_count INT NOT NULL DEFAULT 0,
  status VARCHAR(24) NOT NULL DEFAULT 'running',
  started_at TIMESTAMPTZ,
  finished_at TIMESTAMPTZ,
  created_by VARCHAR(64),
  metadata JSONB,
  created_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ,
  deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_sr_batches_cycle
ON subscription_reconciliation_batches (tenant_uuid, billing_cycle, run_type);

CREATE TABLE IF NOT EXISTS subscription_reconciliation_deltas (
  id UUID PRIMARY KEY,
  tenant_uuid UUID NOT NULL,
  batch_id UUID NOT NULL,
  subscription_ref VARCHAR(64),
  bill_ref VARCHAR(64),
  payment_ref VARCHAR(64),
  delta_type VARCHAR(48) NOT NULL,
  risk_level VARCHAR(16) NOT NULL,
  expected_amount_minor BIGINT NOT NULL DEFAULT 0,
  actual_amount_minor BIGINT NOT NULL DEFAULT 0,
  delta_amount_minor BIGINT NOT NULL DEFAULT 0,
  reason_code VARCHAR(64),
  status VARCHAR(24) NOT NULL DEFAULT 'open',
  delta_fingerprint VARCHAR(128) NOT NULL,
  detected_at TIMESTAMPTZ,
  metadata JSONB,
  created_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ,
  deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_sr_deltas_batch
ON subscription_reconciliation_deltas (tenant_uuid, batch_id, status);
