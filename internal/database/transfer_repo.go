package database

// TransferRuleDTO is returned by GetTransferRules, including the target village name.
type TransferRuleDTO struct {
	ID              int    `json:"id" db:"id"`
	VillageID       int    `json:"villageId" db:"village_id"`
	Position        int    `json:"position" db:"position"`
	TargetVillageID int    `json:"targetVillageId" db:"target_village_id"`
	TargetName      string `json:"targetName" db:"target_name"`
	Wood            int    `json:"wood" db:"wood"`
	Clay            int    `json:"clay" db:"clay"`
	Iron            int    `json:"iron" db:"iron"`
	Crop            int    `json:"crop" db:"crop"`
}

// TransferRuleInput is the DTO for adding a transfer rule.
type TransferRuleInput struct {
	VillageID       int `json:"villageId"`
	TargetVillageID int `json:"targetVillageId"`
	Wood            int `json:"wood"`
	Clay            int `json:"clay"`
	Iron            int `json:"iron"`
	Crop            int `json:"crop"`
}

func (db *DB) GetTransferRules(villageID int) ([]TransferRuleDTO, error) {
	var rules []TransferRuleDTO
	err := db.Select(&rules, `
		SELECT t.id, t.village_id, t.position, t.target_village_id,
		       v.name AS target_name, t.wood, t.clay, t.iron, t.crop
		FROM transfer_rules t
		JOIN villages v ON v.id = t.target_village_id
		WHERE t.village_id = ?
		ORDER BY t.position`, villageID)
	if err != nil {
		return nil, err
	}
	if rules == nil {
		rules = []TransferRuleDTO{}
	}
	return rules, nil
}

func (db *DB) AddTransferRule(input TransferRuleInput) error {
	var maxPos int
	db.Get(&maxPos, "SELECT COALESCE(MAX(position), 0) FROM transfer_rules WHERE village_id = ?", input.VillageID)

	_, err := db.Exec(
		"INSERT INTO transfer_rules (village_id, position, target_village_id, wood, clay, iron, crop) VALUES (?, ?, ?, ?, ?, ?, ?)",
		input.VillageID, maxPos+1, input.TargetVillageID, input.Wood, input.Clay, input.Iron, input.Crop,
	)
	return err
}

func (db *DB) DeleteTransferRule(ruleID int) error {
	_, err := db.Exec("DELETE FROM transfer_rules WHERE id = ?", ruleID)
	return err
}

func (db *DB) DeleteAllTransferRules(villageID int) error {
	_, err := db.Exec("DELETE FROM transfer_rules WHERE village_id = ?", villageID)
	return err
}

func (db *DB) HasTransferRules(villageID int) (bool, error) {
	var count int
	err := db.Get(&count, "SELECT COUNT(*) FROM transfer_rules WHERE village_id = ?", villageID)
	return count > 0, err
}
