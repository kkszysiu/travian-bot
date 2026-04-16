package update

import (
	"fmt"

	"travian-bot/internal/browser"
	"travian-bot/internal/database"
	"travian-bot/internal/parser"
)

// UpdateInventory parses hero inventory items from the current page HTML
// and syncs them to the database for the given account.
func UpdateInventory(b *browser.Browser, db *database.DB, accountID int) error {
	html, err := b.PageHTML()
	if err != nil {
		return fmt.Errorf("get page html: %w", err)
	}
	doc, err := parser.DocFromHTML(html)
	if err != nil {
		return fmt.Errorf("parse html: %w", err)
	}

	items := parser.GetHeroItems(doc)

	dtos := make([]database.HeroItemDTO, len(items))
	for i, item := range items {
		dtos[i] = database.HeroItemDTO{
			Type:   int(item.Type),
			Amount: item.Amount,
		}
	}

	if err := db.UpdateHeroItems(accountID, dtos); err != nil {
		return fmt.Errorf("update hero items in database: %w", err)
	}

	return nil
}
