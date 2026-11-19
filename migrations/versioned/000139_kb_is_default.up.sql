-- Tenant-scoped default knowledge base (at most one per tenant).
ALTER TABLE knowledge_bases
    ADD COLUMN IF NOT EXISTS is_default BOOLEAN NOT NULL DEFAULT FALSE;

CREATE UNIQUE INDEX IF NOT EXISTS uq_knowledge_bases_tenant_default
    ON knowledge_bases (tenant_id)
    WHERE is_default = TRUE AND deleted_at IS NULL;
