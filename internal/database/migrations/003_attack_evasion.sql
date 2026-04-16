-- 003_attack_evasion.sql: Add evasion tracking columns to villages table

ALTER TABLE villages ADD COLUMN evasion_state INTEGER NOT NULL DEFAULT 0;
ALTER TABLE villages ADD COLUMN evasion_target_village_id INTEGER DEFAULT NULL;
