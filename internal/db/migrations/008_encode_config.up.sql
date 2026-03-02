-- Migration 008: add encode_config to jobs

ALTER TABLE jobs ADD COLUMN encode_config JSONB NOT NULL DEFAULT '{}';
