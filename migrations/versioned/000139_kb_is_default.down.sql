DROP INDEX IF EXISTS uq_knowledge_bases_tenant_default;
ALTER TABLE knowledge_bases DROP COLUMN IF EXISTS is_default;
