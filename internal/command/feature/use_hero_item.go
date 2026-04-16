package feature

import (
	"context"
	"fmt"
	"time"

	"travian-bot/internal/browser"
	"travian-bot/internal/domain/enum"
	"travian-bot/internal/parser"
)

// UseHeroItem activates a specific hero inventory item with the given amount.
// The caller must have already navigated to the hero inventory page.
// It clicks the item slot, enters the amount, and confirms the dialog.
func UseHeroItem(ctx context.Context, b *browser.Browser, item enum.HeroItem, amount int64) error {
	if err := clickHeroItem(ctx, b, item); err != nil {
		return fmt.Errorf("click hero item %d: %w", item, err)
	}
	time.Sleep(500 * time.Millisecond)

	if err := enterItemAmount(b, amount); err != nil {
		return fmt.Errorf("enter amount %d for item %d: %w", amount, item, err)
	}
	time.Sleep(500 * time.Millisecond)

	if err := confirmUseItem(ctx, b); err != nil {
		return fmt.Errorf("confirm use item %d: %w", item, err)
	}
	time.Sleep(500 * time.Millisecond)

	return nil
}

// clickHeroItem finds the item slot for the given hero item type and clicks it.
func clickHeroItem(ctx context.Context, b *browser.Browser, item enum.HeroItem) error {
	html, err := b.PageHTML()
	if err != nil {
		return fmt.Errorf("get page html: %w", err)
	}
	doc, err := parser.DocFromHTML(html)
	if err != nil {
		return fmt.Errorf("parse html: %w", err)
	}

	slot := parser.GetItemSlot(doc, item)
	if slot == nil {
		return fmt.Errorf("item slot not found for type %d", item)
	}

	// Build a unique selector using the item class on the child element.
	// The item slot contains a child with the item type as a CSS class number.
	// We use an XPath that finds the heroItem div containing a child with the
	// matching class.
	xpath := fmt.Sprintf(
		`//div[contains(@class,'heroItems')]//div[contains(@class,'heroItem') and not(contains(@class,'empty'))]//*[contains(@class,'%d')]/..`,
		int(item),
	)

	el, err := b.ElementByXPath(xpath)
	if err != nil {
		return fmt.Errorf("find item element by xpath: %w", err)
	}
	if err := b.Click(el); err != nil {
		return fmt.Errorf("click item element: %w", err)
	}

	// Wait for inventory to finish loading after the click
	if err := waitForInventoryLoaded(ctx, b); err != nil {
		return fmt.Errorf("wait for inventory after click: %w", err)
	}

	return nil
}

// enterItemAmount types the desired amount into the consumable item dialog's input.
func enterItemAmount(b *browser.Browser, amount int64) error {
	el, err := b.Element(parser.GetAmountInputSelector())
	if err != nil {
		return fmt.Errorf("find amount input: %w", err)
	}
	if err := b.Input(el, fmt.Sprintf("%d", amount)); err != nil {
		return fmt.Errorf("input amount: %w", err)
	}
	return nil
}

// confirmUseItem clicks the confirm button in the hero item use dialog and
// waits for the inventory to finish reloading.
func confirmUseItem(ctx context.Context, b *browser.Browser) error {
	el, err := b.Element(parser.GetConfirmUseItemButtonSelector())
	if err != nil {
		return fmt.Errorf("find confirm button: %w", err)
	}
	if err := b.Click(el); err != nil {
		return fmt.Errorf("click confirm button: %w", err)
	}

	if err := waitForInventoryLoaded(ctx, b); err != nil {
		return fmt.Errorf("wait for inventory after confirm: %w", err)
	}

	return nil
}

// waitForInventoryLoaded polls the page until the inventory wrapper is loaded
// (no loading class) or a timeout is reached.
func waitForInventoryLoaded(ctx context.Context, b *browser.Browser) error {
	deadline := time.Now().Add(30 * time.Second)
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for inventory to finish loading")
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

		if parser.IsInventoryLoaded(doc) {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
}
