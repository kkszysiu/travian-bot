package database

import "travian-bot/internal/domain/model"

type VillageListItem struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	X             int    `json:"x"`
	Y             int    `json:"y"`
	IsActive      bool   `json:"isActive"`
	IsUnderAttack bool   `json:"isUnderAttack"`
	EvasionState  int    `json:"evasionState"`
}

func (db *DB) GetAccountIDForVillage(villageID int) (int, error) {
	var accountID int
	err := db.Get(&accountID, "SELECT account_id FROM villages WHERE id = ?", villageID)
	return accountID, err
}

func (db *DB) GetAccountIDForJob(jobID int) (int, error) {
	var accountID int
	err := db.Get(&accountID,
		"SELECT v.account_id FROM jobs j JOIN villages v ON j.village_id = v.id WHERE j.id = ?",
		jobID)
	return accountID, err
}

func (db *DB) GetVillages(accountID int) ([]VillageListItem, error) {
	var villages []model.Village
	if err := db.Select(&villages,
		"SELECT id, account_id, name, x, y, is_active, is_under_attack, evasion_state, evasion_target_village_id FROM villages WHERE account_id = ? ORDER BY id",
		accountID,
	); err != nil {
		return nil, err
	}

	items := make([]VillageListItem, len(villages))
	for i, v := range villages {
		items[i] = VillageListItem{
			ID:            v.ID,
			Name:          v.Name,
			X:             v.X,
			Y:             v.Y,
			IsActive:      v.IsActive,
			IsUnderAttack: v.IsUnderAttack,
			EvasionState:  v.EvasionState,
		}
	}
	return items, nil
}
