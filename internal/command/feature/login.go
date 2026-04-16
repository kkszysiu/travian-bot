package feature

import (
	"context"
	"fmt"
	"time"

	"travian-bot/internal/browser"
	"travian-bot/internal/database"
	"travian-bot/internal/parser"
)

// Login performs the login flow: enter credentials and click login button.
func Login(ctx context.Context, b *browser.Browser, db *database.DB, accountID int) error {
	// Check if already logged in
	html, err := b.PageHTML()
	if err != nil {
		return fmt.Errorf("get page html: %w", err)
	}
	doc, err := parser.DocFromHTML(html)
	if err != nil {
		return fmt.Errorf("parse html: %w", err)
	}
	if parser.IsIngamePage(doc) {
		return nil // Already logged in
	}

	// Dismiss any cookie consent overlays that might block interaction
	b.DismissCookieConsent()
	time.Sleep(500 * time.Millisecond)

	// Get login credentials
	username, password, err := getLoginInfo(db, accountID)
	if err != nil {
		return fmt.Errorf("get login info: %w", err)
	}

	// Try multiple username input selectors (modern Travian pages vary)
	usernameSelectors := []string{
		"input[name='name']",
		"input[name='usernameOrEmail']",
		"input[name='username']",
		"input[name='email']",
		"input[type='text'][placeholder]",
		"input[type='email']",
		"#loginScene input[type='text']",
	}
	usernameEl, err := b.ElementBySelectors(usernameSelectors)
	if err != nil {
		// Dump page HTML snippet for debugging
		if pageHTML, htmlErr := b.PageHTML(); htmlErr == nil {
			snippet := pageHTML
			if len(snippet) > 2000 {
				snippet = snippet[:2000]
			}
			return fmt.Errorf("find username input (tried %d selectors), page URL: %s, HTML snippet: %.500s", len(usernameSelectors), b.CurrentURL(), snippet)
		}
		return fmt.Errorf("find username input: %w", err)
	}
	if err := b.Input(usernameEl, username); err != nil {
		return fmt.Errorf("input username: %w", err)
	}

	// Try multiple password input selectors
	passwordSelectors := []string{
		"input[name='password']",
		"input[type='password']",
		"#loginScene input[type='password']",
	}
	passwordEl, err := b.ElementBySelectors(passwordSelectors)
	if err != nil {
		return fmt.Errorf("find password input: %w", err)
	}
	if err := b.Input(passwordEl, password); err != nil {
		return fmt.Errorf("input password: %w", err)
	}

	// Try multiple login button selectors
	loginSelectors := []string{
		"#loginScene button.green",
		"button[type='submit'].green",
		"button[type='submit']",
		"button.green",
		"form button[type='submit']",
	}
	loginBtn, err := b.ElementBySelectors(loginSelectors)
	if err != nil {
		return fmt.Errorf("find login button: %w", err)
	}
	if err := b.Click(loginBtn); err != nil {
		return fmt.Errorf("click login button: %w", err)
	}

	// Wait for navigation to dorf page
	if err := b.WaitPageContains(ctx, "dorf"); err != nil {
		return fmt.Errorf("wait for login redirect: %w", err)
	}

	return nil
}

func getLoginInfo(db *database.DB, accountID int) (string, string, error) {
	var access struct {
		Username string `db:"username"`
		Password string `db:"password"`
	}
	err := db.Get(&access,
		"SELECT username, password FROM accesses WHERE account_id = ? ORDER BY last_used DESC LIMIT 1",
		accountID,
	)
	if err != nil {
		return "", "", fmt.Errorf("no access found for account %d: %w", accountID, err)
	}
	return access.Username, access.Password, nil
}
