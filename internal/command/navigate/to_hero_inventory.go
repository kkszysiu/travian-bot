package navigate

import (
	"context"
	"fmt"
	"time"

	"travian-bot/internal/browser"
	"travian-bot/internal/parser"
)

// ToHeroInventory clicks the hero avatar button and waits for the inventory
// page to fully load (tab active + no loading spinner).
func ToHeroInventory(ctx context.Context, b *browser.Browser) error {
	el, err := b.Element(parser.GetHeroAvatarSelector())
	if err != nil {
		return fmt.Errorf("find hero avatar button: %w", err)
	}
	if err := b.Click(el); err != nil {
		return fmt.Errorf("click hero avatar button: %w", err)
	}

	// Wait for the inventory tab to be active and loaded
	deadline := time.Now().Add(3 * time.Minute)
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for hero inventory page to load")
		}

		html, err := b.PageHTML()
		if err != nil {
			time.Sleep(200 * time.Millisecond)
			continue
		}
		doc, err := parser.DocFromHTML(html)
		if err != nil {
			time.Sleep(200 * time.Millisecond)
			continue
		}

		if parser.IsInventoryPage(doc) && parser.IsInventoryLoaded(doc) {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	time.Sleep(500 * time.Millisecond)
	return nil
}
