package database

type FarmItem struct {
	ID       int    `json:"id" db:"id"`
	Name     string `json:"name" db:"name"`
	IsActive bool   `json:"isActive" db:"is_active"`
}

func (db *DB) GetFarmLists(accountID int) ([]FarmItem, error) {
	var farms []FarmItem
	err := db.Select(&farms,
		"SELECT id, name, is_active FROM farm_lists WHERE account_id = ? ORDER BY id",
		accountID,
	)
	if err != nil {
		return nil, err
	}
	return farms, nil
}
