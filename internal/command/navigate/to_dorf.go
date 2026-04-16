package navigate

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"travian-bot/internal/browser"
	"travian-bot/internal/parser"
)

// ToDorf navigates to dorf1 (resources) or dorf2 (buildings).
// dorf=0 means stay on current dorf, or go to dorf1 if not on any.
func ToDorf(ctx context.Context, b *browser.Browser, dorf int) error {
	currentURL := b.CurrentURL()
	currentDorf := getCurrentDorf(currentURL)

	if dorf == 0 {
		if currentDorf == 0 {
			dorf = 1
		} else {
			dorf = currentDorf
		}
	}

	// Already on correct dorf
	if currentDorf != 0 && dorf == currentDorf {
		return nil
	}

	// Try click-based navigation first; fall back to direct URL on failure
	// (handles stale contexts after page reloads)
	selector := parser.GetDorfButtonSelector(dorf)
	el, err := b.Element(selector)
	if err == nil {
		err = b.Click(el)
	}
	if err != nil {
		// Context lost or element not found — navigate directly via URL
		if navErr := navigateToDorfURL(b, currentURL, dorf); navErr != nil {
			return fmt.Errorf("navigate to dorf%d: %w", dorf, navErr)
		}
	}

	target := fmt.Sprintf("dorf%d.php", dorf)
	if err := b.WaitPageContains(ctx, target); err != nil {
		return fmt.Errorf("wait for %s: %w", target, err)
	}

	// Wait for logo to be visible (page fully loaded)
	waitCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	b.WaitElementVisible(waitCtx, "#logo")

	return nil
}

// navigateToDorfURL constructs the dorf URL from the current page URL and navigates directly.
func navigateToDorfURL(b *browser.Browser, currentURL string, dorf int) error {
	u, err := url.Parse(currentURL)
	if err != nil {
		return fmt.Errorf("parse current URL: %w", err)
	}
	u.Path = fmt.Sprintf("/dorf%d.php", dorf)
	u.RawQuery = ""
	return b.Navigate(u.String())
}

func getCurrentDorf(url string) int {
	if strings.Contains(url, "dorf1") {
		return 1
	}
	if strings.Contains(url, "dorf2") {
		return 2
	}
	return 0
}
