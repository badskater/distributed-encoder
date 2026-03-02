-- Migration 008 rollback

ALTER TABLE jobs DROP COLUMN IF EXISTS encode_config;
