package browser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/stealth"
)

// Browser wraps go-rod with the interface needed for Travian automation.
type Browser struct {
	browser *rod.Browser
	page    *rod.Page

	userAgent string
	headless  bool
	proxyHost string
	proxyPort int
}

// Config holds the settings to launch a browser.
type Config struct {
	UserAgent     string
	Headless      bool
	ProxyHost     string
	ProxyPort     int
	ProxyUsername string
	ProxyPassword string
	ProfilePath   string
}

// New creates and launches a new browser instance.
func New(cfg Config) (*Browser, error) {
	b := &Browser{
		userAgent: cfg.UserAgent,
		headless:  cfg.Headless,
		proxyHost: cfg.ProxyHost,
		proxyPort: cfg.ProxyPort,
	}

	// Find or download Chrome
	path, _ := launcher.LookPath()
	l := launcher.New().Bin(path)

	// Cache directory — use persistent path (same approach as C# version)
	if cfg.ProfilePath != "" {
		home, _ := os.UserHomeDir()
		var baseDir string
		switch runtime.GOOS {
		case "darwin":
			baseDir = filepath.Join(home, "Library", "Application Support", "travian-bot")
		case "linux":
			configDir, _ := os.UserConfigDir()
			baseDir = filepath.Join(configDir, "travian-bot")
		default:
			appData := os.Getenv("APPDATA")
			if appData == "" {
				appData = filepath.Join(home, "AppData", "Roaming")
			}
			baseDir = filepath.Join(appData, "travian-bot")
		}
		cacheDir := filepath.Join(baseDir, "cache", cfg.ProfilePath)
		os.MkdirAll(cacheDir, 0755)

		// Remove stale Chrome lock files left behind by crashed processes
		for _, lockFile := range []string{"SingletonLock", "SingletonCookie", "SingletonSocket"} {
			os.Remove(filepath.Join(cacheDir, lockFile))
		}

		l = l.UserDataDir(cacheDir)
	}

	if cfg.Headless {
		l = l.Headless(true)
	} else {
		l = l.Headless(false)
	}

	if cfg.ProxyHost != "" {
		l = l.Proxy(fmt.Sprintf("%s:%d", cfg.ProxyHost, cfg.ProxyPort))
	}

	// Anti-detection flags
	l = l.Set("disable-blink-features", "AutomationControlled").
		Set("disable-features", "UserAgentClientHint").
		Set("no-default-browser-check").
		Set("no-first-run").
		Set("mute-audio").
		Set("disable-background-timer-throttling").
		Set("disable-backgrounding-occluded-windows")

	u, err := l.Launch()
	if err != nil {
		return nil, fmt.Errorf("launch browser: %w", err)
	}

	browser := rod.New().ControlURL(u)
	if err := browser.Connect(); err != nil {
		return nil, fmt.Errorf("connect browser: %w", err)
	}
	b.browser = browser

	// Create stealth page
	page, err := stealth.Page(browser)
	if err != nil {
		browser.Close()
		return nil, fmt.Errorf("create stealth page: %w", err)
	}
	b.page = page

	// Set user agent
	if cfg.UserAgent != "" {
		err = page.SetUserAgent(&proto.NetworkSetUserAgentOverride{
			UserAgent: cfg.UserAgent,
		})
		if err != nil {
			browser.Close()
			return nil, fmt.Errorf("set user agent: %w", err)
		}
	}

	// Set up proxy authentication if needed
	if cfg.ProxyHost != "" && cfg.ProxyUsername != "" {
		go browser.MustHandleAuth(cfg.ProxyUsername, cfg.ProxyPassword)()
	}

	return b, nil
}

// Navigate opens a URL and waits for load.
func (b *Browser) Navigate(url string) error {
	if err := b.page.Navigate(url); err != nil {
		return fmt.Errorf("navigate to %s: %w", url, err)
	}
	if err := b.page.WaitLoad(); err != nil {
		return fmt.Errorf("wait load: %w", err)
	}
	return nil
}

// Refresh reloads the current page.
func (b *Browser) Refresh() error {
	return b.page.Reload()
}

// PageHTML returns the current page's HTML source.
func (b *Browser) PageHTML() (string, error) {
	return b.page.HTML()
}

// CurrentURL returns the current page URL.
func (b *Browser) CurrentURL() string {
	info, err := b.page.Info()
	if err != nil {
		return ""
	}
	return info.URL
}

// Element waits for a CSS selector and returns the element (up to 3 min timeout).
func (b *Browser) Element(selector string) (*rod.Element, error) {
	page := b.page.Timeout(3 * time.Minute)
	el, err := page.Element(selector)
	if err != nil {
		return nil, fmt.Errorf("element %q not found: %w", selector, err)
	}
	return el, nil
}

// ElementByXPath waits for an XPath expression and returns the element (up to 3 min timeout).
func (b *Browser) ElementByXPath(xpath string) (*rod.Element, error) {
	page := b.page.Timeout(3 * time.Minute)
	el, err := page.ElementX(xpath)
	if err != nil {
		return nil, fmt.Errorf("element xpath %q not found: %w", xpath, err)
	}
	return el, nil
}

// Click clicks an element.
func (b *Browser) Click(el *rod.Element) error {
	return el.Click(proto.InputMouseButtonLeft, 1)
}

// Input clears and types into an element.
func (b *Browser) Input(el *rod.Element, text string) error {
	if err := el.SelectAllText(); err != nil {
		// Element might not support text selection, just clear
	}
	return el.Input(text)
}

// WaitPageContains waits until the URL contains the given substring.
func (b *Browser) WaitPageContains(ctx context.Context, urlSubstr string) error {
	timeout := 3 * time.Minute
	deadline := time.Now().Add(timeout)
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for URL to contain %q, current: %s", urlSubstr, b.CurrentURL())
		}
		if strings.Contains(b.CurrentURL(), urlSubstr) {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
}

// WaitElementVisible waits for an element to be visible by CSS selector.
func (b *Browser) WaitElementVisible(ctx context.Context, selector string) error {
	timeout := 3 * time.Minute
	deadline := time.Now().Add(timeout)
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for element %q", selector)
		}
		// Use a short timeout per attempt so we don't block the loop
		if el := b.TryElement(selector, 2*time.Second); el != nil {
			visible, _ := el.Visible()
			if visible {
				return nil
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
}

// ElementBySelectors tries multiple CSS selectors in order, returning the first match.
// Uses a short timeout per selector instead of waiting 3 minutes for each.
func (b *Browser) ElementBySelectors(selectors []string) (*rod.Element, error) {
	for _, sel := range selectors {
		if el := b.TryElement(sel, 3*time.Second); el != nil {
			return el, nil
		}
	}
	return nil, fmt.Errorf("no element found for selectors: %v", selectors)
}

// TryElement attempts to find an element with a short timeout, returning nil if not found.
func (b *Browser) TryElement(selector string, timeout time.Duration) *rod.Element {
	page := b.page.Timeout(timeout)
	el, err := page.Element(selector)
	if err != nil {
		return nil
	}
	return el
}

// DismissCookieConsent tries to close common cookie consent overlays.
func (b *Browser) DismissCookieConsent() {
	// Try common cookie consent button selectors
	selectors := []string{
		".cmpboxbtn.cmpboxbtnyes",       // Quantcast/CMP "Accept" button
		"#CybotCookiebotDialogBodyLevelButtonLevelOptinAllowAll", // Cookiebot
		".cc-btn.cc-dismiss",             // Cookie Consent plugin
		"button.consent-accept",          // Generic
		"[data-testid='cookie-accept']",  // Generic test id
		".cmpboxbtn",                     // CMP first button (usually accept)
	}
	for _, sel := range selectors {
		if el := b.TryElement(sel, 2*time.Second); el != nil {
			el.Click(proto.InputMouseButtonLeft, 1)
			time.Sleep(500 * time.Millisecond)
			return
		}
	}
	// Also try dismissing via JS (common CMP frameworks)
	b.page.Eval(`() => {
		try {
			if (window.__cmpConsent) window.__cmpConsent();
			if (window.CookieConsent && window.CookieConsent.acceptAll) window.CookieConsent.acceptAll();
		} catch(e) {}
	}`)
}

// Screenshot captures the current page.
func (b *Browser) Screenshot() (string, error) {
	dir := filepath.Join(os.TempDir(), "travian-bot", "screenshots")
	os.MkdirAll(dir, 0755)
	filename := filepath.Join(dir, fmt.Sprintf("%s.png", time.Now().Format("2006-01-02_15-04-05")))

	data, err := b.page.Screenshot(true, nil)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return "", err
	}
	return filename, nil
}

// ExecuteJS runs JavaScript on the page.
func (b *Browser) ExecuteJS(script string) error {
	_, err := b.page.Eval(script)
	return err
}

// Close shuts down the browser.
func (b *Browser) Close() {
	if b.page != nil {
		b.page.Close()
	}
	if b.browser != nil {
		b.browser.Close()
	}
}

// Page returns the underlying rod page for advanced operations.
func (b *Browser) Page() *rod.Page {
	return b.page
}
