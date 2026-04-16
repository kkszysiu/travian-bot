package update

import (
	"fmt"

	"travian-bot/internal/browser"
	"travian-bot/internal/database"
	"travian-bot/internal/event"
	"travian-bot/internal/parser"
)

// UpdateStorage parses resource amounts and capacity from page HTML and upserts to the database.
func UpdateStorage(b *browser.Browser, db *database.DB, bus *event.Bus, villageID int) error {
	html, err := b.PageHTML()
	if err != nil {
		return fmt.Errorf("get page html: %w", err)
	}
	doc, err := parser.DocFromHTML(html)
	if err != nil {
		return fmt.Errorf("parse html: %w", err)
	}

	wood := int(parser.GetWood(doc))
	clay := int(parser.GetClay(doc))
	iron := int(parser.GetIron(doc))
	crop := int(parser.GetCrop(doc))
	warehouse := int(parser.GetWarehouseCapacity(doc))
	granary := int(parser.GetGranaryCapacity(doc))
	freeCrop := int(parser.GetFreeCrop(doc))

	// Check if storage record already exists
	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM storages WHERE village_id = ?", villageID)
	if err != nil {
		return fmt.Errorf("check storage: %w", err)
	}

	if count == 0 {
		_, err = db.Exec(
			"INSERT INTO storages (village_id, wood, clay, iron, crop, warehouse, granary, free_crop) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			villageID, wood, clay, iron, crop, warehouse, granary, freeCrop,
		)
	} else {
		_, err = db.Exec(
			"UPDATE storages SET wood = ?, clay = ?, iron = ?, crop = ?, warehouse = ?, granary = ?, free_crop = ? WHERE village_id = ?",
			wood, clay, iron, crop, warehouse, granary, freeCrop, villageID,
		)
	}
	if err != nil {
		return fmt.Errorf("save storage: %w", err)
	}
	bus.Emit(event.StorageModified, villageID)
	return nil
}
