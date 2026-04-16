package navigate

import (
	"context"
	"fmt"
	"time"

	"travian-bot/internal/browser"
	"travian-bot/internal/parser"
)

// ToQuestPage clicks the questmaster button and waits for the quest page to load.
func ToQuestPage(ctx context.Context, b *browser.Browser) error {
	el, err := b.Element(parser.GetQuestMasterSelector())
	if err != nil {
		return fmt.Errorf("find questmaster button: %w", err)
	}
	if err := b.Click(el); err != nil {
		return fmt.Errorf("click questmaster button: %w", err)
	}

	// Wait for quest page
	if err := b.WaitElementVisible(ctx, "div.tasks.tasksVillage"); err != nil {
		return fmt.Errorf("wait for quest page: %w", err)
	}
	time.Sleep(500 * time.Millisecond)
	return nil
}
