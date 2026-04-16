package update

import (
	"fmt"
	"log"
	"strings"
	"time"

	"travian-bot/internal/browser"
	"travian-bot/internal/database"
	"travian-bot/internal/domain/model"
	"travian-bot/internal/event"
	"travian-bot/internal/parser"
)

// UpdateVillageList parses villages from page HTML, syncs them to the database,
// and emits a VillagesModified event.
func UpdateVillageList(b *browser.Browser, db *database.DB, bus *event.Bus, accountID int) error {
	// Brief pause to ensure page content is fully rendered
	time.Sleep(1 * time.Second)

	html, err := b.PageHTML()
	if err != nil {
		return fmt.Errorf("get page html: %w", err)
	}
	doc, err := parser.DocFromHTML(html)
	if err != nil {
		return fmt.Errorf("parse html: %w", err)
	}

	parsed := parser.GetVillages(doc)
	if len(parsed) == 0 {
		// Log a snippet of the sidebar HTML for debugging
		found := false
		for _, marker := range []string{"sidebarBoxVillageList", "sidebarBoxVillagelist"} {
			if idx := strings.Index(html, marker); idx >= 0 {
				end := idx + 500
				if end > len(html) {
					end = len(html)
				}
				log.Printf("[UpdateVillageList] 0 villages parsed but sidebar '%s' found. HTML snippet: %s", marker, html[idx:end])
				found = true
				break
			}
		}
		if !found {
			log.Printf("[UpdateVillageList] 0 villages parsed. Village sidebar NOT found in page HTML (length=%d, URL=%s)", len(html), b.CurrentURL())
		}
		return nil
	}
	log.Printf("[UpdateVillageList] parsed %d village(s)", len(parsed))

	// Get existing villages from DB
	var existing []model.Village
	if err := db.Select(&existing,
		"SELECT id, account_id, name, x, y, is_active, is_under_attack, evasion_state, evasion_target_village_id FROM villages WHERE account_id = ?",
		accountID,
	); err != nil {
		return fmt.Errorf("get existing villages: %w", err)
	}

	existingMap := make(map[int]model.Village, len(existing))
	for _, v := range existing {
		existingMap[v.ID] = v
	}

	parsedMap := make(map[int]parser.VillageInfo, len(parsed))
	for _, v := range parsed {
		parsedMap[v.ID] = v
	}

	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete villages that no longer exist
	for _, v := range existing {
		if _, found := parsedMap[v.ID]; !found {
			if _, err := tx.Exec("DELETE FROM villages WHERE id = ?", v.ID); err != nil {
				return fmt.Errorf("delete village %d: %w", v.ID, err)
			}
		}
	}

	// Insert new or update existing villages
	for _, p := range parsed {
		isActive := 0
		if p.IsActive {
			isActive = 1
		}
		isUnderAttack := 0
		if p.IsUnderAttack {
			isUnderAttack = 1
		}

		if _, found := existingMap[p.ID]; found {
			// Update existing village
			_, err := tx.Exec(
				"UPDATE villages SET name = ?, x = ?, y = ?, is_active = ?, is_under_attack = ? WHERE id = ?",
				p.Name, p.X, p.Y, isActive, isUnderAttack, p.ID,
			)
			if err != nil {
				return fmt.Errorf("update village %d: %w", p.ID, err)
			}
		} else {
			// Insert new village with its parsed ID
			_, err := tx.Exec(
				"INSERT INTO villages (id, account_id, name, x, y, is_active, is_under_attack) VALUES (?, ?, ?, ?, ?, ?, ?)",
				p.ID, accountID, p.Name, p.X, p.Y, isActive, isUnderAttack,
			)
			if err != nil {
				return fmt.Errorf("insert village %d: %w", p.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	// Fill default settings for newly added villages
	for _, p := range parsed {
		if _, found := existingMap[p.ID]; !found {
			if err := db.FillVillageSettingsForNew(accountID, p.ID); err != nil {
				return fmt.Errorf("fill village settings for %d: %w", p.ID, err)
			}
		}
	}

	bus.Emit(event.VillagesModified, accountID)
	return nil
}
