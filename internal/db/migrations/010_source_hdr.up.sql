-- Migration 010: Dolby Vision and HDR10+ metadata on sources

-- Add HDR type and Dolby Vision profile columns to sources.
-- hdr_type: empty string = SDR/unknown, otherwise 'hdr10', 'hdr10+', 'dolby_vision', 'hlg'
-- dv_profile: 0 = no DV, otherwise the DV profile number (5, 7, 8, 9, ...)
ALTER TABLE sources
    ADD COLUMN IF NOT EXISTS hdr_type   TEXT     NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS dv_profile SMALLINT NOT NULL DEFAULT 0;

-- Extend the analysis_results type check to include 'hdr_detect'.
ALTER TABLE analysis_results
    DROP CONSTRAINT IF EXISTS analysis_results_type_check;

ALTER TABLE analysis_results
    ADD CONSTRAINT analysis_results_type_check
    CHECK (type IN ('histogram', 'vmaf', 'scene_detect', 'hdr_detect'));
