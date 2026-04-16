package update

import (
	"fmt"

	"travian-bot/internal/browser"
	"travian-bot/internal/database"
	"travian-bot/internal/parser"
)

// UpdateAccountInfo parses gold, silver, plus account status from page HTML and saves to database.
func UpdateAccountInfo(b *browser.Browser, db *database.DB, accountID int) error {
	html, err := b.PageHTML()
	if err != nil {
		return fmt.Errorf("get page html: %w", err)
	}
	doc, err := parser.DocFromHTML(html)
	if err != nil {
		return fmt.Errorf("parse html: %w", err)
	}

	gold := parser.GetGold(doc)
	silver := parser.GetSilver(doc)
	hasPlusAccount := parser.HasPlusAccount(doc)

	plusInt := 0
	if hasPlusAccount {
		plusInt = 1
	}

	// Check if account info already exists
	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM accounts_info WHERE account_id = ?", accountID)
	if err != nil {
		return fmt.Errorf("check account info: %w", err)
	}

	if count == 0 {
		_, err = db.Exec(
			"INSERT INTO accounts_info (account_id, gold, silver, has_plus_account) VALUES (?, ?, ?, ?)",
			accountID, gold, silver, plusInt,
		)
	} else {
		_, err = db.Exec(
			"UPDATE accounts_info SET gold = ?, silver = ?, has_plus_account = ? WHERE account_id = ?",
			gold, silver, plusInt, accountID,
		)
	}
	if err != nil {
		return fmt.Errorf("save account info: %w", err)
	}
	return nil
}
