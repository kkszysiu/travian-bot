package feature

import (
	"context"
	"fmt"
	"strings"
	"time"

	"travian-bot/internal/browser"
	"travian-bot/internal/command/navigate"
	"travian-bot/internal/parser"
)

// ClaimQuest navigates to the quest page and claims all available quests.
func ClaimQuest(ctx context.Context, b *browser.Browser) error {
	// Navigate to quest page
	if err := navigate.ToQuestPage(ctx, b); err != nil {
		// pointer-events: none means the questmaster button is disabled (no quests)
		if strings.Contains(err.Error(), "pointer-events") {
			return nil
		}
		return fmt.Errorf("to quest page: %w", err)
	}

	// Keep claiming quests while available
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		html, err := b.PageHTML()
		if err != nil {
			return fmt.Errorf("get page html: %w", err)
		}
		doc, err := parser.DocFromHTML(html)
		if err != nil {
			return fmt.Errorf("parse html: %w", err)
		}

		if !parser.HasQuestCollectButton(doc) {
			// Try switching to tab 1 (first quest tab)
			if err := navigate.SwitchTab(ctx, b, 1); err != nil {
				return nil // No quests to claim
			}
			time.Sleep(500 * time.Millisecond)

			html, err = b.PageHTML()
			if err != nil {
				return nil
			}
			doc, err = parser.DocFromHTML(html)
			if err != nil {
				return nil
			}

			if !parser.HasQuestCollectButton(doc) {
				return nil // Still no quests
			}
		}

		el, err := b.Element(parser.GetQuestCollectButtonSelector())
		if err != nil {
			return nil // Done
		}
		if err := b.Click(el); err != nil {
			return fmt.Errorf("click collect button: %w", err)
		}
		time.Sleep(500 * time.Millisecond)

		// Check if there are more
		html, err = b.PageHTML()
		if err != nil {
			return nil
		}
		doc, err = parser.DocFromHTML(html)
		if err != nil {
			return nil
		}
		if !parser.IsQuestClaimable(doc) {
			return nil
		}
	}
}
