-- Knowledge-base graph schema: attributes, entities, relations, and ingest tasks.
CREATE TABLE IF NOT EXISTS kb_graph_attributes (
    id TEXT PRIMARY KEY,
    tenant_id INTEGER NOT NULL,
    knowledge_base_id TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL,
    code TEXT NOT NULL,
    value_type TEXT NOT NULL DEFAULT 'free',
    data_type TEXT NOT NULL DEFAULT 'string',
    is_label INTEGER NOT NULL DEFAULT 0,
    is_static INTEGER NOT NULL DEFAULT 1,
    description TEXT NOT NULL DEFAULT '',
    enum_values TEXT NOT NULL DEFAULT '[]',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_kb_graph_attributes_kb_code
    ON kb_graph_attributes(tenant_id, knowledge_base_id, code);
CREATE INDEX IF NOT EXISTS idx_kb_graph_attributes_kb
    ON kb_graph_attributes(tenant_id, knowledge_base_id);

CREATE TABLE IF NOT EXISTS kb_graph_entities (
    id TEXT PRIMARY KEY,
    tenant_id INTEGER NOT NULL,
    knowledge_base_id TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL,
    code TEXT NOT NULL,
    identity_key TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    attribute_ids TEXT NOT NULL DEFAULT '[]',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_kb_graph_entities_kb_code
    ON kb_graph_entities(tenant_id, knowledge_base_id, code);
CREATE INDEX IF NOT EXISTS idx_kb_graph_entities_kb
    ON kb_graph_entities(tenant_id, knowledge_base_id);

CREATE TABLE IF NOT EXISTS kb_graph_relations (
    id TEXT PRIMARY KEY,
    tenant_id INTEGER NOT NULL,
    knowledge_base_id TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL,
    code TEXT NOT NULL,
    source_entity_id TEXT NOT NULL,
    target_entity_id TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    attribute_ids TEXT NOT NULL DEFAULT '[]',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_kb_graph_relations_kb_code
    ON kb_graph_relations(tenant_id, knowledge_base_id, code);
CREATE INDEX IF NOT EXISTS idx_kb_graph_relations_kb
    ON kb_graph_relations(tenant_id, knowledge_base_id);

CREATE TABLE IF NOT EXISTS kb_graph_ingest_tasks (
    id TEXT PRIMARY KEY,
    tenant_id INTEGER NOT NULL,
    knowledge_base_id TEXT NOT NULL,
    task_kind TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    source_type TEXT NOT NULL DEFAULT 'table_mapping',
    source_ref TEXT NOT NULL DEFAULT '{}',
    progress INTEGER NOT NULL DEFAULT 0,
    error_message TEXT NOT NULL DEFAULT '',
    started_at DATETIME,
    finished_at DATETIME,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_kb_graph_ingest_tasks_kb
    ON kb_graph_ingest_tasks(tenant_id, knowledge_base_id, created_at);
CREATE INDEX IF NOT EXISTS idx_kb_graph_ingest_tasks_status
    ON kb_graph_ingest_tasks(tenant_id, knowledge_base_id, status);
