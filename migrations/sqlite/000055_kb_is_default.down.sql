DROP INDEX IF EXISTS uq_knowledge_bases_tenant_default;
-- SQLite cannot DROP COLUMN portably across versions used here; leave the column.