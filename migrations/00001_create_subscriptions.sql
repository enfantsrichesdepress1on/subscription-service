-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS subscriptions (
    id BIGSERIAL PRIMARY KEY,
    service_name TEXT NOT NULL,
    price INTEGER NOT NULL CHECK (price > 0),
    user_id UUID NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ NULL,
    CHECK (end_date IS NULL OR end_date >= start_date),
    CHECK (date_trunc('month', start_date)::date = start_date),
    CHECK (end_date IS NULL OR date_trunc('month', end_date)::date = end_date)
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_subscriptions_service_name ON subscriptions(service_name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_subscriptions_period ON subscriptions(start_date, end_date) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS subscriptions;
