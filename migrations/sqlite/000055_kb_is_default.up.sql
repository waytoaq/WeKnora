-- Tenant-scoped default knowledge base (at most one per tenant).
ALTER TABLE knowledge_bases ADD COLUMN is_default INTEGER NOT NULL DEFAULT 0;

CREATE UNIQUE INDEX IF NOT EXISTS uq_knowledge_bases_tenant_default
    ON knowledge_bases (tenant_id)
    WHERE is_default = 1 AND deleted_at IS NULL;
