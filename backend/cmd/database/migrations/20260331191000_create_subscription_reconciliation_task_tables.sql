-- 013-subscription-reconciliation foundational tables (part 2)
CREATE TABLE IF NOT EXISTS subscription_reconciliation_delta_tasks (
  id UUID PRIMARY KEY,
  tenant_uuid UUID NOT NULL,
  delta_id UUID NOT NULL,
  delta_fingerprint VARCHAR(128) NOT NULL,
  assignee VARCHAR(64),
  priority VARCHAR(24),
  sla_level VARCHAR(16) NOT NULL,
  sla_deadline TIMESTAMPTZ,
  status VARCHAR(24) NOT NULL DEFAULT 'pending',
  resolution VARCHAR(32),
  resolution_note TEXT,
  closed_at TIMESTAMPTZ,
  metadata JSONB,
  created_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ,
  deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_sr_delta_tasks_delta
ON subscription_reconciliation_delta_tasks (tenant_uuid, delta_id, status);

CREATE TABLE IF NOT EXISTS subscription_reconciliation_governance_policies (
  id UUID PRIMARY KEY,
  tenant_uuid UUID NOT NULL,
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  retry_windows JSONB,
  escalation_threshold INT NOT NULL DEFAULT 3,
  notify_channels JSONB,
  version INT NOT NULL DEFAULT 1,
  effective_from TIMESTAMPTZ,
  effective_to TIMESTAMPTZ,
  metadata JSONB,
  created_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ,
  deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS subscription_reconciliation_execution_logs (
  id UUID PRIMARY KEY,
  tenant_uuid UUID NOT NULL,
  subscription_ref VARCHAR(64) NOT NULL,
  action_type VARCHAR(24) NOT NULL,
  attempt_no INT NOT NULL DEFAULT 0,
  scheduled_at TIMESTAMPTZ,
  executed_at TIMESTAMPTZ,
  result VARCHAR(16) NOT NULL,
  failure_reason TEXT,
  operator_type VARCHAR(16) NOT NULL DEFAULT 'system',
  operator_id VARCHAR(64),
  metadata JSONB,
  created_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ,
  deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_sr_execution_logs_subscription
ON subscription_reconciliation_execution_logs (tenant_uuid, subscription_ref, action_type);
