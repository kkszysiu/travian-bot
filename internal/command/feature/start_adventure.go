package feature

import (
	"context"
	"fmt"
	"time"

	"travian-bot/internal/browser"
	"travian-bot/internal/command/navigate"
	"travian-bot/internal/parser"
)

// StartAdventure navigates to the adventure page, checks if an adventure can be started,
// and starts the first available adventure. Returns the adventure duration in seconds.
func StartAdventure(ctx context.Context, b *browser.Browser) (int, error) {
	// Navigate to adventure page
	if err := navigate.ToAdventurePage(ctx, b); err != nil {
		return 0, fmt.Errorf("to adventure page: %w", err)
	}

	html, err := b.PageHTML()
	if err != nil {
		return 0, fmt.Errorf("get page html: %w", err)
	}
	doc, err := parser.DocFromHTML(html)
	if err != nil {
		return 0, fmt.Errorf("parse html: %w", err)
	}

	if !parser.CanStartAdventure(doc) {
		return 0, nil // No adventure available
	}

	// Click the adventure button
	el, err := b.Element(parser.GetAdventureButtonSelector())
	if err != nil {
		return 0, fmt.Errorf("find adventure button: %w", err)
	}
	if err := b.Click(el); err != nil {
		return 0, fmt.Errorf("click adventure button: %w", err)
	}

	// Wait for continue button
	if err := b.WaitElementVisible(ctx, parser.GetContinueButtonSelector()); err != nil {
		return 0, fmt.Errorf("wait for continue button: %w", err)
	}

	// Get duration before clicking continue
	html, err = b.PageHTML()
	if err != nil {
		return 0, fmt.Errorf("get duration page html: %w", err)
	}
	doc, err = parser.DocFromHTML(html)
	if err != nil {
		return 0, fmt.Errorf("parse duration html: %w", err)
	}
	duration := parser.GetAdventureDuration(doc)

	// Click continue button to start the adventure
	continueEl, err := b.Element(parser.GetContinueButtonSelector())
	if err != nil {
		return 0, fmt.Errorf("find continue button: %w", err)
	}
	if err := b.Click(continueEl); err != nil {
		return 0, fmt.Errorf("click continue button: %w", err)
	}

	time.Sleep(500 * time.Millisecond)
	return duration, nil
}
