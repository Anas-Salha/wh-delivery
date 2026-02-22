-- Webhook registrations
CREATE TABLE webhooks (
    id BIGSERIAL PRIMARY KEY,
    client_id BIGINT NOT NULL,
    callback_url TEXT NOT NULL,
    signing_secret TEXT NOT NULL,
    event_types TEXT[] NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    retry_config JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_webhooks_client_event ON webhooks(client_id, event_types);

-- Event log (for auditing and replay)
CREATE TABLE events (
    id BIGSERIAL PRIMARY KEY,
    source_id BIGINT NOT NULL,
    idempotency_key TEXT NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE (source_id, idempotency_key)
);

CREATE INDEX idx_events_source_created ON events(source_id, created_at);

-- Delivery tracking
CREATE TABLE deliveries (
    id BIGSERIAL PRIMARY KEY,
    event_id BIGINT NOT NULL REFERENCES events(id),
    webhook_id BIGINT NOT NULL REFERENCES webhooks(id),
    status VARCHAR(20) NOT NULL,  -- pending, success, failed
    attempts INT DEFAULT 0,
    last_attempt_at TIMESTAMP,
    last_error TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_deliveries_status ON deliveries(status, webhook_id);
