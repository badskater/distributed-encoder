-- Rollback migration 010

ALTER TABLE sources
    DROP COLUMN IF EXISTS hdr_type,
    DROP COLUMN IF EXISTS dv_profile;

-- Restore original type check (hdr_detect removed).
ALTER TABLE analysis_results
    DROP CONSTRAINT IF EXISTS analysis_results_type_check;

ALTER TABLE analysis_results
    ADD CONSTRAINT analysis_results_type_check
    CHECK (type IN ('histogram', 'vmaf', 'scene_detect'));
