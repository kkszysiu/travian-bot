package feature

import (
	"context"
	"fmt"
	"time"

	"travian-bot/internal/browser"
	"travian-bot/internal/parser"
)

// DisableContextualHelp navigates to options and disables contextual help.
func DisableContextualHelp(ctx context.Context, b *browser.Browser) error {
	// Click options button
	optBtn, err := b.Element(parser.GetOptionButtonSelector())
	if err != nil {
		return fmt.Errorf("find options button: %w", err)
	}
	if err := b.Click(optBtn); err != nil {
		return fmt.Errorf("click options button: %w", err)
	}

	// Wait for options page
	waitCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := b.WaitElementVisible(waitCtx, parser.GetHideContextualHelpSelector()); err != nil {
		return fmt.Errorf("wait for options page: %w", err)
	}

	// Click hide contextual help checkbox
	checkbox, err := b.Element(parser.GetHideContextualHelpSelector())
	if err != nil {
		return fmt.Errorf("find contextual help checkbox: %w", err)
	}
	if err := b.Click(checkbox); err != nil {
		return fmt.Errorf("click contextual help checkbox: %w", err)
	}

	// Click submit
	submitBtn, err := b.Element(parser.GetSubmitButtonSelector())
	if err != nil {
		return fmt.Errorf("find submit button: %w", err)
	}
	if err := b.Click(submitBtn); err != nil {
		return fmt.Errorf("click submit button: %w", err)
	}

	return nil
}
