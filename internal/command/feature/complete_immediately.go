package feature

import (
	"context"
	"fmt"
	"time"

	"travian-bot/internal/browser"
	"travian-bot/internal/command/navigate"
	"travian-bot/internal/parser"
)

// CompleteImmediately uses gold to complete the current building queue immediately.
// Navigates to dorf, finds the complete button, clicks it, and confirms.
func CompleteImmediately(ctx context.Context, b *browser.Browser) error {
	// Navigate to dorf (any dorf page shows the queue)
	if err := navigate.ToDorf(ctx, b, 0); err != nil {
		return fmt.Errorf("navigate to dorf: %w", err)
	}
	time.Sleep(500 * time.Millisecond)

	html, err := b.PageHTML()
	if err != nil {
		return fmt.Errorf("get page html: %w", err)
	}
	doc, err := parser.DocFromHTML(html)
	if err != nil {
		return fmt.Errorf("parse html: %w", err)
	}

	oldCount := parser.CountQueueBuilding(doc)
	if oldCount == 0 {
		return nil // Nothing to complete
	}

	// Click the complete button
	el, err := b.Element(parser.GetCompleteButtonSelector())
	if err != nil {
		return fmt.Errorf("find complete button: %w", err)
	}
	if err := b.Click(el); err != nil {
		return fmt.Errorf("click complete button: %w", err)
	}

	// Wait for confirm dialog
	time.Sleep(1 * time.Second)

	// Click confirm button
	confirmEl, err := b.Element(parser.GetConfirmButtonSelector())
	if err != nil {
		return fmt.Errorf("find confirm button: %w", err)
	}
	if err := b.Click(confirmEl); err != nil {
		return fmt.Errorf("click confirm button: %w", err)
	}

	// Wait for queue to change
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		time.Sleep(500 * time.Millisecond)
		html, err = b.PageHTML()
		if err != nil {
			continue
		}
		doc, err = parser.DocFromHTML(html)
		if err != nil {
			continue
		}
		newCount := parser.CountQueueBuilding(doc)
		if newCount != oldCount {
			return nil
		}
	}

	return fmt.Errorf("queue count did not change after complete immediately")
}
