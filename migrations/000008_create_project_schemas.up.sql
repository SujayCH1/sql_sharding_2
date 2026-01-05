-- =========================================
-- Schema state enum
-- =========================================
CREATE TYPE schema_state AS ENUM (
    'draft',
    'pending',
    'applying',
    'applied',
    'failed'
);

-- =========================================
-- Project schema versions (DDL log)
-- =========================================
CREATE TABLE project_schemas (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL,

    version INTEGER NOT NULL,
    state schema_state NOT NULL DEFAULT 'draft',

    -- Raw DDL SQL exactly as provided by user
    ddl_sql TEXT NOT NULL,

    -- Error message if schema failed
    error_message TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    committed_at TIMESTAMPTZ,
    applied_at TIMESTAMPTZ,

    CONSTRAINT fk_project_schemas_project
        FOREIGN KEY (project_id)
        REFERENCES projects(id)
        ON DELETE CASCADE,

    CONSTRAINT uq_project_schema_version
        UNIQUE (project_id, version)
);

CREATE INDEX idx_project_schemas_project
    ON project_schemas(project_id);

CREATE INDEX idx_project_schemas_state
    ON project_schemas(project_id, state);
