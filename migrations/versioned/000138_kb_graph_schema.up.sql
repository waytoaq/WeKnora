-- Knowledge-base graph schema: attributes, entities, relations, and ingest tasks.
CREATE TABLE IF NOT EXISTS kb_graph_attributes (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    knowledge_base_id VARCHAR(36) NOT NULL,
    category VARCHAR(64) NOT NULL DEFAULT '',
    name VARCHAR(200) NOT NULL,
    code VARCHAR(128) NOT NULL,
    value_type VARCHAR(32) NOT NULL DEFAULT 'free',
    data_type VARCHAR(32) NOT NULL DEFAULT 'string',
    is_label BOOLEAN NOT NULL DEFAULT FALSE,
    is_static BOOLEAN NOT NULL DEFAULT TRUE,
    description TEXT NOT NULL DEFAULT '',
    enum_values JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_kb_graph_attributes_kb_code
    ON kb_graph_attributes(tenant_id, knowledge_base_id, code);
CREATE INDEX IF NOT EXISTS idx_kb_graph_attributes_kb
    ON kb_graph_attributes(tenant_id, knowledge_base_id);

CREATE TABLE IF NOT EXISTS kb_graph_entities (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    knowledge_base_id VARCHAR(36) NOT NULL,
    category VARCHAR(32) NOT NULL DEFAULT '',
    name VARCHAR(200) NOT NULL,
    code VARCHAR(128) NOT NULL,
    identity_key VARCHAR(128) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    attribute_ids JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_kb_graph_entities_kb_code
    ON kb_graph_entities(tenant_id, knowledge_base_id, code);
CREATE INDEX IF NOT EXISTS idx_kb_graph_entities_kb
    ON kb_graph_entities(tenant_id, knowledge_base_id);

CREATE TABLE IF NOT EXISTS kb_graph_relations (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    knowledge_base_id VARCHAR(36) NOT NULL,
    category VARCHAR(64) NOT NULL DEFAULT '',
    name VARCHAR(200) NOT NULL,
    code VARCHAR(128) NOT NULL,
    source_entity_id VARCHAR(36) NOT NULL,
    target_entity_id VARCHAR(36) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    attribute_ids JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_kb_graph_relations_kb_code
    ON kb_graph_relations(tenant_id, knowledge_base_id, code);
CREATE INDEX IF NOT EXISTS idx_kb_graph_relations_kb
    ON kb_graph_relations(tenant_id, knowledge_base_id);

CREATE TABLE IF NOT EXISTS kb_graph_ingest_tasks (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    knowledge_base_id VARCHAR(36) NOT NULL,
    task_kind VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    source_type VARCHAR(32) NOT NULL DEFAULT 'table_mapping',
    source_ref JSONB NOT NULL DEFAULT '{}',
    progress INT NOT NULL DEFAULT 0,
    error_message TEXT NOT NULL DEFAULT '',
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_kb_graph_ingest_tasks_kb
    ON kb_graph_ingest_tasks(tenant_id, knowledge_base_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_kb_graph_ingest_tasks_status
    ON kb_graph_ingest_tasks(tenant_id, knowledge_base_id, status);
