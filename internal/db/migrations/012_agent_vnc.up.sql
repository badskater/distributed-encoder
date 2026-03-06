-- Migration 012: Add VNC port to agents table.
-- Stores the TCP port the agent's VNC server is listening on.
-- 0 means VNC is not configured/running on that agent.
ALTER TABLE agents ADD COLUMN IF NOT EXISTS vnc_port INTEGER NOT NULL DEFAULT 0;
