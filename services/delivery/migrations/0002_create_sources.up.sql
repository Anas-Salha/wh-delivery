-- Source registrations
CREATE TABLE sources (
    id BIGSERIAL PRIMARY KEY,
    source_name TEXT NOT NULL UNIQUE,
    api_key TEXT NOT NULL UNIQUE,
    webhook_secret TEXT NOT NULL,
    allowed_event_types TEXT[],
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_sources_api_key ON sources(api_key);
CREATE INDEX idx_sources_status ON sources(status);
