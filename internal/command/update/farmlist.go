package update

import (
	"fmt"

	"travian-bot/internal/browser"
	"travian-bot/internal/database"
	"travian-bot/internal/domain/model"
	"travian-bot/internal/event"
	"travian-bot/internal/parser"
)

// UpdateFarmList parses farm lists from the current page HTML, syncs them to the
// database (deleting removed lists, inserting new ones, updating existing ones),
// and emits a FarmsModified event.
// Uses the Travian farm list ID (data-list attribute) as the database ID,
// matching how the C# version works.
func UpdateFarmList(b *browser.Browser, db *database.DB, bus *event.Bus, accountID int) error {
	html, err := b.PageHTML()
	if err != nil {
		return fmt.Errorf("get page html: %w", err)
	}
	doc, err := parser.DocFromHTML(html)
	if err != nil {
		return fmt.Errorf("parse html: %w", err)
	}

	parsed := parser.GetFarmLists(doc)
	if len(parsed) == 0 {
		// Don't delete existing farm lists when we can't parse any —
		// this likely means the page is not on the farm list tab.
		return nil
	}

	// Get existing farm lists from DB
	var existing []model.Farm
	if err := db.Select(&existing,
		"SELECT id, account_id, name, is_active FROM farm_lists WHERE account_id = ?",
		accountID,
	); err != nil {
		return fmt.Errorf("get existing farm lists: %w", err)
	}

	existingByID := make(map[int]model.Farm, len(existing))
	for _, f := range existing {
		existingByID[f.ID] = f
	}

	parsedIDs := make(map[int]bool, len(parsed))
	for _, p := range parsed {
		parsedIDs[p.ID] = true
	}

	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete farm lists that no longer exist on the page
	for _, f := range existing {
		if !parsedIDs[f.ID] {
			if _, err := tx.Exec("DELETE FROM farm_lists WHERE id = ?", f.ID); err != nil {
				return fmt.Errorf("delete farm list %d (%s): %w", f.ID, f.Name, err)
			}
		}
	}

	// Insert new or update existing farm lists (using Travian ID as the DB ID)
	for _, p := range parsed {
		if _, found := existingByID[p.ID]; found {
			// Update existing farm list name
			_, err := tx.Exec(
				"UPDATE farm_lists SET name = ? WHERE id = ?",
				p.Name, p.ID,
			)
			if err != nil {
				return fmt.Errorf("update farm list %d: %w", p.ID, err)
			}
		} else {
			// Insert new farm list with Travian ID and is_active = 1
			_, err := tx.Exec(
				"INSERT INTO farm_lists (id, account_id, name, is_active) VALUES (?, ?, ?, 1)",
				p.ID, accountID, p.Name,
			)
			if err != nil {
				return fmt.Errorf("insert farm list %d %q: %w", p.ID, p.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	bus.Emit(event.FarmsModified, accountID)
	return nil
}
