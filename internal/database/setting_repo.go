package database

import "travian-bot/internal/domain/enum"

func (db *DB) GetAccountSettings(accountID int) (map[string]int, error) {
	rows, err := db.Queryx(
		"SELECT setting, value FROM accounts_setting WHERE account_id = ?",
		accountID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settings := make(map[string]int)
	// Start with defaults
	for k, v := range enum.DefaultAccountSettings {
		settings[k.String()] = v
	}

	for rows.Next() {
		var setting, value int
		if err := rows.Scan(&setting, &value); err != nil {
			return nil, err
		}
		settings[enum.AccountSetting(setting).String()] = value
	}
	return settings, nil
}

func (db *DB) SaveAccountSettings(accountID int, settings map[string]int) error {
	// Build reverse lookup: string name -> enum value
	nameToEnum := make(map[string]enum.AccountSetting)
	for _, s := range enum.AllAccountSettings {
		nameToEnum[s.String()] = s
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for name, value := range settings {
		settingEnum, ok := nameToEnum[name]
		if !ok {
			continue
		}
		_, err := tx.Exec(
			`INSERT INTO accounts_setting (account_id, setting, value) VALUES (?, ?, ?)
			 ON CONFLICT(account_id, setting) DO UPDATE SET value = excluded.value`,
			accountID, int(settingEnum), value,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (db *DB) GetVillageSettings(villageID int) (map[string]int, error) {
	rows, err := db.Queryx(
		"SELECT setting, value FROM villages_setting WHERE village_id = ?",
		villageID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settings := make(map[string]int)
	// Start with defaults
	for k, v := range enum.DefaultVillageSettings {
		settings[villageSettingName(k)] = v
	}

	for rows.Next() {
		var setting, value int
		if err := rows.Scan(&setting, &value); err != nil {
			return nil, err
		}
		settings[villageSettingName(enum.VillageSetting(setting))] = value
	}
	return settings, nil
}

func (db *DB) SaveVillageSettings(villageID int, settings map[string]int) error {
	nameToEnum := make(map[string]enum.VillageSetting)
	for _, s := range enum.AllVillageSettings {
		nameToEnum[villageSettingName(s)] = s
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for name, value := range settings {
		settingEnum, ok := nameToEnum[name]
		if !ok {
			continue
		}
		_, err := tx.Exec(
			`INSERT INTO villages_setting (village_id, setting, value) VALUES (?, ?, ?)
			 ON CONFLICT(village_id, setting) DO UPDATE SET value = excluded.value`,
			villageID, int(settingEnum), value,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// villageSettingName converts a VillageSetting enum to its string name.
func villageSettingName(s enum.VillageSetting) string {
	names := map[enum.VillageSetting]string{
		enum.VillageSettingUseHeroResourceForBuilding:      "UseHeroResourceForBuilding",
		enum.VillageSettingApplyRomanQueueLogicWhenBuilding: "ApplyRomanQueueLogicWhenBuilding",
		enum.VillageSettingUseSpecialUpgrade:               "UseSpecialUpgrade",
		enum.VillageSettingCompleteImmediately:             "CompleteImmediately",
		enum.VillageSettingTribe:                           "Tribe",
		enum.VillageSettingTrainTroopEnable:                "TrainTroopEnable",
		enum.VillageSettingTrainTroopRepeatTimeMin:         "TrainTroopRepeatTimeMin",
		enum.VillageSettingTrainTroopRepeatTimeMax:         "TrainTroopRepeatTimeMax",
		enum.VillageSettingTrainWhenLowResource:            "TrainWhenLowResource",
		enum.VillageSettingBarrackTroop:                    "BarrackTroop",
		enum.VillageSettingBarrackAmountMin:                "BarrackAmountMin",
		enum.VillageSettingBarrackAmountMax:                "BarrackAmountMax",
		enum.VillageSettingStableTroop:                     "StableTroop",
		enum.VillageSettingStableAmountMin:                 "StableAmountMin",
		enum.VillageSettingStableAmountMax:                 "StableAmountMax",
		enum.VillageSettingGreatBarrackTroop:               "GreatBarrackTroop",
		enum.VillageSettingGreatBarrackAmountMin:           "GreatBarrackAmountMin",
		enum.VillageSettingGreatBarrackAmountMax:           "GreatBarrackAmountMax",
		enum.VillageSettingGreatStableTroop:                "GreatStableTroop",
		enum.VillageSettingGreatStableAmountMin:            "GreatStableAmountMin",
		enum.VillageSettingGreatStableAmountMax:            "GreatStableAmountMax",
		enum.VillageSettingWorkshopTroop:                   "WorkshopTroop",
		enum.VillageSettingWorkshopAmountMin:               "WorkshopAmountMin",
		enum.VillageSettingWorkshopAmountMax:               "WorkshopAmountMax",
		enum.VillageSettingAutoNPCEnable:                   "AutoNPCEnable",
		enum.VillageSettingAutoNPCOverflow:                 "AutoNPCOverflow",
		enum.VillageSettingAutoNPCGranaryPercent:           "AutoNPCGranaryPercent",
		enum.VillageSettingAutoNPCWood:                     "AutoNPCWood",
		enum.VillageSettingAutoNPCClay:                     "AutoNPCClay",
		enum.VillageSettingAutoNPCIron:                     "AutoNPCIron",
		enum.VillageSettingAutoNPCCrop:                     "AutoNPCCrop",
		enum.VillageSettingAutoRefreshEnable:               "AutoRefreshEnable",
		enum.VillageSettingAutoRefreshMin:                  "AutoRefreshMin",
		enum.VillageSettingAutoRefreshMax:                  "AutoRefreshMax",
		enum.VillageSettingAutoClaimQuestEnable:            "AutoClaimQuestEnable",
		enum.VillageSettingCompleteImmediatelyTime:         "CompleteImmediatelyTime",
		enum.VillageSettingAutoSendResourceEnable:          "AutoSendResourceEnable",
		enum.VillageSettingAutoSendResourceRepeatMin:       "AutoSendResourceRepeatMin",
		enum.VillageSettingAutoSendResourceRepeatMax:       "AutoSendResourceRepeatMax",
		enum.VillageSettingAutoSendResourceThreshold:       "AutoSendResourceThreshold",
		enum.VillageSettingAttackEvasionEnable:             "AttackEvasionEnable",
		enum.VillageSettingAttackEvasionSafeVillageID:      "AttackEvasionSafeVillageID",
		enum.VillageSettingAttackEvasionEvacResources:      "AttackEvasionEvacResources",
		enum.VillageSettingAttackEvasionCheckIntervalMin:   "AttackEvasionCheckIntervalMin",
		enum.VillageSettingAttackEvasionCheckIntervalMax:   "AttackEvasionCheckIntervalMax",
	}
	if name, ok := names[s]; ok {
		return name
	}
	return "Unknown"
}
