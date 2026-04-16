package feature

import (
	"context"
	"fmt"
	"time"

	"travian-bot/internal/browser"
	"travian-bot/internal/command/navigate"
	"travian-bot/internal/command/update"
	"travian-bot/internal/database"
	"travian-bot/internal/domain/enum"
	"travian-bot/internal/event"
	"travian-bot/internal/parser"
)

// EvadeTroops sends ALL available troops from the current village to a target village as reinforcement.
func EvadeTroops(ctx context.Context, b *browser.Browser, db *database.DB, bus *event.Bus,
	villageID int, targetX, targetY int) error {

	// Navigate to rally point
	if err := navigate.ToDorf(ctx, b, 2); err != nil {
		return fmt.Errorf("navigate to dorf2: %w", err)
	}
	if err := update.UpdateBuildings(b, db, bus, villageID); err != nil {
		return fmt.Errorf("update buildings: %w", err)
	}
	if err := navigate.ToBuildingByType(ctx, b, db, villageID, int(enum.BuildingRallyPoint)); err != nil {
		return fmt.Errorf("navigate to rally point: %w", err)
	}

	// Switch to "Send troops" tab (tt=2 in Travian)
	if err := navigate.SwitchTab(ctx, b, 2); err != nil {
		return fmt.Errorf("switch to send troops tab: %w", err)
	}
	time.Sleep(500 * time.Millisecond)

	// Parse page for available troops
	html, err := b.PageHTML()
	if err != nil {
		return fmt.Errorf("get rally point html: %w", err)
	}
	doc, err := parser.DocFromHTML(html)
	if err != nil {
		return fmt.Errorf("parse rally point html: %w", err)
	}

	slots := parser.GetRallyPointTroopSlots(doc)
	if len(slots) == 0 {
		return nil // No troops available to send
	}

	// Input max troops for each slot — use TryElement with short timeout
	// since we already know which inputs exist from the parser
	for _, slot := range slots {
		inputEl := b.TryElement(parser.GetRallyPointTroopInputSelector(slot.InputName), 5*time.Second)
		if inputEl == nil {
			continue
		}
		if err := b.Input(inputEl, fmt.Sprintf("%d", slot.Available)); err != nil {
			return fmt.Errorf("input troop %s: %w", slot.InputName, err)
		}
		time.Sleep(200 * time.Millisecond)
	}

	// Input target coordinates
	xEl := b.TryElement(parser.GetRallyPointCoordXSelector(), 5*time.Second)
	if xEl == nil {
		return fmt.Errorf("coord X input not found")
	}
	if err := b.Input(xEl, fmt.Sprintf("%d", targetX)); err != nil {
		return fmt.Errorf("input coord X: %w", err)
	}
	time.Sleep(200 * time.Millisecond)

	yEl := b.TryElement(parser.GetRallyPointCoordYSelector(), 5*time.Second)
	if yEl == nil {
		return fmt.Errorf("coord Y input not found")
	}
	if err := b.Input(yEl, fmt.Sprintf("%d", targetY)); err != nil {
		return fmt.Errorf("input coord Y: %w", err)
	}
	time.Sleep(200 * time.Millisecond)

	// Select reinforcement radio (eventType=5)
	radioEl := b.TryElement(parser.GetReinforcementRadioSelector(), 5*time.Second)
	if radioEl == nil {
		return fmt.Errorf("reinforcement radio not found")
	}
	if err := b.Click(radioEl); err != nil {
		return fmt.Errorf("click reinforcement radio: %w", err)
	}
	time.Sleep(200 * time.Millisecond)

	// Click send button (button#ok)
	sendBtn := b.TryElement(parser.GetRallyPointSendButtonSelector(), 5*time.Second)
	if sendBtn == nil {
		return fmt.Errorf("send button not found")
	}
	if err := b.Click(sendBtn); err != nil {
		return fmt.Errorf("click send button: %w", err)
	}

	// Wait for confirmation page to load
	time.Sleep(2 * time.Second)

	// The confirm button on the confirmation page has id="confirmSendTroops"
	// and class="rallyPointConfirm"
	confirmSelectors := []string{
		"button#confirmSendTroops",           // Exact ID from Travian confirm page
		"button.rallyPointConfirm",           // By class name
		"button#s1",                          // Fallback for other Travian versions
		"button#ok",                          // Fallback
	}
	confirmBtn, err := b.ElementBySelectors(confirmSelectors)
	if err != nil {
		// No confirm button found — possibly troops were sent directly without confirmation
		return nil
	}
	if err := b.Click(confirmBtn); err != nil {
		return fmt.Errorf("click confirm button: %w", err)
	}

	time.Sleep(1 * time.Second)
	return nil
}
