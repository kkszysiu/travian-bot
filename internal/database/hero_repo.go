package database

import "fmt"

type HeroItemDTO struct {
	Type   int `json:"type" db:"type"`
	Amount int `json:"amount" db:"amount"`
}

func (db *DB) GetHeroItems(accountID int) ([]HeroItemDTO, error) {
	var items []HeroItemDTO
	err := db.Select(&items,
		"SELECT type, amount FROM hero_items WHERE account_id = ? ORDER BY type",
		accountID,
	)
	if err != nil {
		return nil, err
	}
	return items, nil
}

// GetHeroItemAmount returns the amount of a specific hero item type for an account.
// Returns 0 if the item is not found.
func (db *DB) GetHeroItemAmount(accountID int, itemType int) (int, error) {
	var amount int
	err := db.Get(&amount,
		"SELECT amount FROM hero_items WHERE account_id = ? AND type = ?",
		accountID, itemType,
	)
	if err != nil {
		return 0, nil // Item not found, return 0
	}
	return amount, nil
}

// UpdateHeroItems syncs the hero items in the database with the provided list.
// It deletes items no longer present, inserts new ones, and updates existing amounts.
func (db *DB) UpdateHeroItems(accountID int, items []HeroItemDTO) error {
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get existing items
	var existing []struct {
		ID     int `db:"id"`
		Type   int `db:"type"`
		Amount int `db:"amount"`
	}
	err = tx.Select(&existing,
		"SELECT id, type, amount FROM hero_items WHERE account_id = ?",
		accountID,
	)
	if err != nil {
		return fmt.Errorf("select existing items: %w", err)
	}

	// Build lookup of new items by type
	newByType := make(map[int]int, len(items))
	for _, item := range items {
		newByType[item.Type] = item.Amount
	}

	// Build lookup of existing items by type
	existingByType := make(map[int]int, len(existing))
	existingIDByType := make(map[int]int, len(existing))
	for _, e := range existing {
		existingByType[e.Type] = e.Amount
		existingIDByType[e.Type] = e.ID
	}

	// Delete items no longer present
	for _, e := range existing {
		if _, found := newByType[e.Type]; !found {
			if _, err := tx.Exec("DELETE FROM hero_items WHERE id = ?", e.ID); err != nil {
				return fmt.Errorf("delete hero item %d: %w", e.ID, err)
			}
		}
	}

	// Insert new items and update existing ones
	for _, item := range items {
		if _, found := existingByType[item.Type]; found {
			// Update
			if _, err := tx.Exec(
				"UPDATE hero_items SET amount = ? WHERE id = ?",
				item.Amount, existingIDByType[item.Type],
			); err != nil {
				return fmt.Errorf("update hero item type %d: %w", item.Type, err)
			}
		} else {
			// Insert
			if _, err := tx.Exec(
				"INSERT INTO hero_items (account_id, type, amount) VALUES (?, ?, ?)",
				accountID, item.Type, item.Amount,
			); err != nil {
				return fmt.Errorf("insert hero item type %d: %w", item.Type, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
