package database

// SetEvasionState updates the evasion state and optionally the target village ID.
func (db *DB) SetEvasionState(villageID, state int, targetVillageID *int) error {
	_, err := db.Exec(
		"UPDATE villages SET evasion_state = ?, evasion_target_village_id = ? WHERE id = ?",
		state, targetVillageID, villageID,
	)
	return err
}

// ClearEvasionState resets evasion state to 0 and target to NULL.
func (db *DB) ClearEvasionState(villageID int) error {
	_, err := db.Exec(
		"UPDATE villages SET evasion_state = 0, evasion_target_village_id = NULL WHERE id = ?",
		villageID,
	)
	return err
}
