package navigate

import (
	"context"
	"fmt"
	"time"

	"travian-bot/internal/browser"
	"travian-bot/internal/parser"
)

// ToAdventurePage clicks the hero adventure button and waits for the adventure list page to load.
func ToAdventurePage(ctx context.Context, b *browser.Browser) error {
	el, err := b.Element(parser.GetHeroAdventureButtonSelector())
	if err != nil {
		return fmt.Errorf("find adventure button: %w", err)
	}
	if err := b.Click(el); err != nil {
		return fmt.Errorf("click adventure button: %w", err)
	}

	// Wait for adventure table
	if err := b.WaitElementVisible(ctx, "table.adventureList"); err != nil {
		return fmt.Errorf("wait for adventure page: %w", err)
	}
	time.Sleep(500 * time.Millisecond)
	return nil
}
