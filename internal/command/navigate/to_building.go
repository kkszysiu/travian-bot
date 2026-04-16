package navigate

import (
	"context"
	"fmt"
	"net/url"

	"travian-bot/internal/browser"
)

// ToBuilding navigates directly to the build page for the given location.
func ToBuilding(ctx context.Context, b *browser.Browser, location int) error {
	currentURL := b.CurrentURL()
	baseURL := extractBaseURL(currentURL)
	if baseURL == "" {
		return fmt.Errorf("cannot determine base URL from %q", currentURL)
	}

	buildURL := fmt.Sprintf("%s/build.php?id=%d", baseURL, location)
	if err := b.Navigate(buildURL); err != nil {
		return fmt.Errorf("navigate to building at location %d: %w", location, err)
	}

	if err := b.WaitPageContains(ctx, "build"); err != nil {
		return fmt.Errorf("wait for build page: %w", err)
	}

	return nil
}

// extractBaseURL returns the scheme + host portion of a URL.
func extractBaseURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	if u.Scheme == "" || u.Host == "" {
		return ""
	}
	return fmt.Sprintf("%s://%s", u.Scheme, u.Host)
}
