package navigate

import (
	"context"
	"fmt"
	"strings"

	"travian-bot/internal/browser"
	"travian-bot/internal/parser"
)

// SwitchTab switches to the building tab at the given index (0-based).
// Navigates by extracting the tab's href and loading it directly.
func SwitchTab(ctx context.Context, b *browser.Browser, tabIndex int) error {
	html, err := b.PageHTML()
	if err != nil {
		return fmt.Errorf("get page HTML: %w", err)
	}
	doc, err := parser.DocFromHTML(html)
	if err != nil {
		return fmt.Errorf("parse HTML: %w", err)
	}

	tabCount := parser.CountTab(doc)
	if tabCount == 0 {
		return nil // No tabs present on this page
	}
	if tabIndex < 0 || tabIndex >= tabCount {
		return fmt.Errorf("tab index %d out of range (0..%d)", tabIndex, tabCount-1)
	}

	// Already on the requested tab
	if parser.IsTabActive(doc, tabIndex) {
		return nil
	}

	// Get the tab's href and navigate directly
	href := parser.GetTabHref(doc, tabIndex)
	if href == "" {
		return fmt.Errorf("tab %d has no href", tabIndex)
	}

	// Build absolute URL from relative href
	currentURL := b.CurrentURL()
	baseURL := currentURL
	if idx := strings.Index(currentURL, "/build.php"); idx >= 0 {
		baseURL = currentURL[:idx]
	} else if idx := strings.Index(currentURL, "/dorf"); idx >= 0 {
		baseURL = currentURL[:idx]
	}
	targetURL := baseURL + href

	if err := b.Navigate(targetURL); err != nil {
		return fmt.Errorf("navigate to tab %d: %w", tabIndex, err)
	}

	return nil
}
