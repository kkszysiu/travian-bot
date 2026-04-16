package navigate

import (
	"context"
	"fmt"
	"time"

	"github.com/go-rod/rod"

	"travian-bot/internal/browser"
	"travian-bot/internal/parser"
)

// SwitchVillage clicks on a village in the sidebar to make it active.
// Retries a few times if the sidebar hasn't loaded yet (e.g., after relogin).
func SwitchVillage(ctx context.Context, b *browser.Browser, villageID int) error {
	selectors := parser.VillageSidebarSelectors(villageID)

	// Try multiple times — the sidebar is React-rendered and may take time to load
	var el *rod.Element
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		el, err = b.ElementBySelectors(selectors)
		if err == nil {
			break
		}
		// Refresh the page to force sidebar to reload
		if attempt == 0 {
			_ = b.Navigate(b.CurrentURL())
			time.Sleep(2 * time.Second)
		} else {
			time.Sleep(3 * time.Second)
		}
	}
	if err != nil {
		return fmt.Errorf("find village %d in sidebar: %w", villageID, err)
	}

	// Click the anchor inside the list entry
	link, err := el.Element("a")
	if err != nil {
		return fmt.Errorf("find village link: %w", err)
	}

	if err := b.Click(link); err != nil {
		return fmt.Errorf("click village %d: %w", villageID, err)
	}

	// Wait for village to become active
	waitCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	activeSelectors := parser.VillageSidebarActiveSelectors(villageID)
	for _, sel := range activeSelectors {
		if err := b.WaitElementVisible(waitCtx, sel); err == nil {
			return nil
		}
	}
	return fmt.Errorf("village %d did not become active", villageID)
}
