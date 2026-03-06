-- Migration 012 rollback: Remove VNC port from agents table.
ALTER TABLE agents DROP COLUMN IF EXISTS vnc_port;
