CREATE TABLE IF NOT EXISTS transfer_rules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    village_id INTEGER NOT NULL,
    position INTEGER NOT NULL DEFAULT 0,
    target_village_id INTEGER NOT NULL,
    wood INTEGER NOT NULL DEFAULT 0,
    clay INTEGER NOT NULL DEFAULT 0,
    iron INTEGER NOT NULL DEFAULT 0,
    crop INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (village_id) REFERENCES villages(id) ON DELETE CASCADE,
    FOREIGN KEY (target_village_id) REFERENCES villages(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_transfer_rules_village_id ON transfer_rules(village_id);
