package service

import (
	"travian-bot/internal/database"
	"travian-bot/internal/domain/enum"
)

// GetAccountSettingValue returns a single setting value for an account, with default fallback.
func GetAccountSettingValue(db *database.DB, accountID int, setting enum.AccountSetting) (int, error) {
	var value int
	err := db.Get(&value,
		"SELECT value FROM accounts_setting WHERE account_id = ? AND setting = ?",
		accountID, int(setting),
	)
	if err != nil {
		// Return default
		if def, ok := enum.DefaultAccountSettings[setting]; ok {
			return def, nil
		}
		return 0, nil
	}
	return value, nil
}

// GetVillageSettingValue returns a single setting value for a village, with default fallback.
func GetVillageSettingValue(db *database.DB, villageID int, setting enum.VillageSetting) (int, error) {
	var value int
	err := db.Get(&value,
		"SELECT value FROM villages_setting WHERE village_id = ? AND setting = ?",
		villageID, int(setting),
	)
	if err != nil {
		if def, ok := enum.DefaultVillageSettings[setting]; ok {
			return def, nil
		}
		return 0, nil
	}
	return value, nil
}
