-- Knowledge-base text preprocessing options applied before chunking.
ALTER TABLE knowledge_bases
    ADD COLUMN IF NOT EXISTS text_preprocess_config JSONB NOT NULL DEFAULT '{}'::jsonb;
